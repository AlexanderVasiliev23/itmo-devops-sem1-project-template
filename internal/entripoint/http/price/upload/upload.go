package upload

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"project_sem/internal/dataprovider"
	"project_sem/internal/model"
	"project_sem/internal/utils/postgres"
	"strconv"
	"time"
)

type Entrypoint struct {
	priceProvider       *dataprovider.PriceProvider
	transactionProvider *postgres.TransactionProvider
}

func NewUploadEntrypoint(
	priceProvider *dataprovider.PriceProvider,
	transactionProvider *postgres.TransactionProvider,
) *Entrypoint {
	return &Entrypoint{
		priceProvider:       priceProvider,
		transactionProvider: transactionProvider,
	}
}

type InvalidFileData struct {
	Message string
}

func (e *InvalidFileData) Error() string {
	return e.Message
}

func (e *Entrypoint) Handle(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		log.Println(err)
		return
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, file); err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		http.Error(w, "Ошибка открытия ZIP архива", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	newPriceModels, err := parseArchive(zipReader)
	if err != nil {
		var e *InvalidFileData
		switch {
		case errors.As(err, &e):
			http.Error(w, e.Message, http.StatusBadRequest)
		default:
			http.Error(w, "Неизвестная ошибка", http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	var allPriceModels model.PriceModels

	// в рамках одной транзакции сначала вставляем новые записи (prices),
	// а затем из базы получаем все модели (prices), включая новые только что вставленные
	err = e.transactionProvider.RunInTransaction(r.Context(), func(ctx context.Context) error {
		if err := e.priceProvider.InsertBatch(ctx, newPriceModels); err != nil {
			return err
		}

		allPriceModels, err = e.priceProvider.List(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		http.Error(w, "Неизвестная ошибка", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := struct {
		TotalItems      int    `json:"total_items"`
		TotalCategories int    `json:"total_categories"`
		TotalPrice      string `json:"total_price"`
	}{
		TotalItems:      len(allPriceModels),
		TotalCategories: allPriceModels.UniqueCategoriesCount(),
		TotalPrice:      fmt.Sprintf("%.2f", allPriceModels.TotalPrice()),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка маршалинга JSON", http.StatusInternalServerError)
		return
	}
}

func parseArchive(zipReader *zip.Reader) (model.PriceModels, error) {
	var out model.PriceModels

	for _, f := range zipReader.File {
		if filepath.Ext(f.Name) != ".csv" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		csvReader := csv.NewReader(rc)
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if len(record) > 0 && record[0] == "id" {
				continue
			}

			priceModel, err := mapCsvRecordToModel(record)
			if err != nil {
				return nil, err
			}
			out = append(out, priceModel)
		}

		rc.Close()
	}

	return out, nil
}

func mapCsvRecordToModel(record []string) (*model.PriceModel, error) {
	if len(record) != 5 {
		return nil, &InvalidFileData{
			Message: "Ошибка чтения CSV файла: неправильный формат, должно быть 5 колонок",
		}
	}

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return nil, &InvalidFileData{
			Message: "Ошибка чтения CSV файла: неправильный id",
		}
	}

	price, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return nil, &InvalidFileData{
			Message: "Ошибка чтения CSV файла: неправильный формат цены",
		}
	}

	createDate, err := time.Parse(time.DateOnly, record[4])
	if err != nil {
		return nil, &InvalidFileData{
			Message: "Ошибка чтения CSV файла: неправильный формат даты",
		}
	}

	return &model.PriceModel{
		ID:         id,
		Name:       record[1],
		Category:   record[2],
		Price:      price,
		CreateDate: createDate,
	}, nil
}
