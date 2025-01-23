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

	addedItemsCount := 0

	err = e.transactionProvider.RunInTransaction(r.Context(), func(ctx context.Context) error {
		for _, f := range zipReader.File {
			if filepath.Ext(f.Name) != ".csv" {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				return err
			}

			csvReader := csv.NewReader(rc)
			for {
				record, err := csvReader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				if len(record) != 5 {
					return &InvalidFileData{
						Message: "Ошибка чтения CSV файла: неправильный формат, должно быть 5 колонок",
					}
				}

				if record[0] == "id" {
					continue
				}

				id, err := strconv.Atoi(record[0])
				if err != nil {
					return &InvalidFileData{
						Message: "Ошибка чтения CSV файла: неправильный id",
					}
				}

				price, err := strconv.ParseFloat(record[3], 64)
				if err != nil {
					return &InvalidFileData{
						Message: "Ошибка чтения CSV файла: неправильный формат цены",
					}
				}

				createDate, err := time.Parse(time.DateOnly, record[4])
				if err != nil {
					return &InvalidFileData{
						Message: "Ошибка чтения CSV файла: неправильный формат даты",
					}
				}

				priceModel := model.PriceModel{
					ID:         id,
					Name:       record[1],
					Category:   record[2],
					Price:      price,
					CreateDate: createDate,
				}

				addedItemsCount++

				if err := e.priceProvider.Insert(ctx, priceModel); err != nil {
					return err
				}
			}

			rc.Close()
		}

		return nil
	})
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

	list, err := e.priceProvider.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := struct {
		TotalItems      int    `json:"total_items"`
		TotalCategories int    `json:"total_categories"`
		TotalPrice      string `json:"total_price"`
	}{
		TotalItems:      addedItemsCount,
		TotalCategories: list.UniqueCategoriesCount(),
		TotalPrice:      fmt.Sprintf("%.2f", list.TotalPrice()),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка маршалинга JSON", http.StatusInternalServerError)
		return
	}
}
