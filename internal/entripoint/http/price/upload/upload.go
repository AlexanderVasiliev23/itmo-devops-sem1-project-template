package upload

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"project_sem/internal/dataprovider"
	"project_sem/internal/model"
	"strconv"
	"time"
)

type Entrypoint struct {
	priceProvider *dataprovider.PriceProvider
}

func NewUploadEntrypoint(priceProvider *dataprovider.PriceProvider) *Entrypoint {
	return &Entrypoint{
		priceProvider: priceProvider,
	}
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

	var (
		uniqueCategories = make(map[string]struct{})
		totalItems       int
		totalPrice       float64
	)

	for _, f := range zipReader.File {
		if filepath.Ext(f.Name) != ".csv" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			http.Error(w, "Ошибка открытия файла", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		csvReader := csv.NewReader(rc)
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, "Ошибка чтения CSV файла", http.StatusBadRequest)
				log.Println(err)
				return
			}

			if len(record) != 5 {
				http.Error(w, "Ошибка чтения CSV файла: неправильный формат, должно быть 5 колонок", http.StatusBadRequest)
				log.Println(err)
				return
			}

			if record[0] == "id" {
				continue
			}

			id, err := strconv.Atoi(record[0])
			if err != nil {
				http.Error(w, "Ошибка чтения CSV файла: неправильный id", http.StatusBadRequest)
				log.Println(err)
				return
			}

			price, err := strconv.ParseFloat(record[3], 64)
			if err != nil {
				http.Error(w, "Ошибка чтения CSV файла: неправильный формат цены", http.StatusBadRequest)
				log.Println(err)
				return
			}

			createDate, err := time.Parse(time.DateOnly, record[4])
			if err != nil {
				http.Error(w, "Ошибка чтения CSV файла: неправильный формат даты", http.StatusBadRequest)
				log.Println(err)
				return
			}

			priceModel := model.PriceModel{
				ID:         id,
				Name:       record[1],
				Category:   record[2],
				Price:      price,
				CreateDate: createDate,
			}

			totalItems++
			uniqueCategories[priceModel.Category] = struct{}{}
			totalPrice += price

			if err := e.priceProvider.Insert(r.Context(), priceModel); err != nil {
				http.Error(w, "Ошибка сохранения в базу данных", http.StatusInternalServerError)
				log.Println(err)
				return
			}
		}

		rc.Close()
	}

	w.Header().Set("Content-Type", "application/json")

	resp := struct {
		TotalItems      int    `json:"total_items"`
		TotalCategories int    `json:"total_categories"`
		TotalPrice      string `json:"total_price"`
	}{
		TotalItems:      totalItems,
		TotalCategories: len(uniqueCategories),
		TotalPrice:      fmt.Sprintf("%.2f", totalPrice),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Ошибка маршалинга JSON", http.StatusInternalServerError)
		return
	}
}
