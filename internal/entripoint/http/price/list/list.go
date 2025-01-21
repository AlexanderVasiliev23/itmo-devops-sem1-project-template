package list

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"project_sem/internal/dataprovider"
	"project_sem/internal/model"
	"time"
)

type Entrypoint struct {
	priceProvider *dataprovider.PriceProvider
}

func NewListEntrypoint(priceProvider *dataprovider.PriceProvider) *Entrypoint {
	return &Entrypoint{
		priceProvider: priceProvider,
	}
}

func (u Entrypoint) Handle(w http.ResponseWriter, r *http.Request) {
	prices, err := u.priceProvider.List(r.Context())
	if err != nil {
		http.Error(w, "Ошибка получения данных из бд", http.StatusInternalServerError)
		return
	}

	csvBuf := new(bytes.Buffer)
	err = writePricesToCSV(prices, csvBuf)
	if err != nil {
		http.Error(w, "Ошибка создания CSV файла", http.StatusInternalServerError)
		return
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	csvWriter, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Ошибка создания ZIP архива", http.StatusInternalServerError)
		return
	}

	if _, err = csvWriter.Write(csvBuf.Bytes()); err != nil {
		http.Error(w, "Ошибка записи CSV файла в ZIP архив", http.StatusInternalServerError)
		return
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Ошибка закрытия ZIP архива", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=prices.zip")

	if _, err := w.Write(zipBuffer.Bytes()); err != nil {
		http.Error(w, "Ошибка отправки ZIP файла", http.StatusInternalServerError)
		return
	}
}

func writePricesToCSV(prices model.PriceModels, buf *bytes.Buffer) error {
	writer := csv.NewWriter(buf)
	defer writer.Flush()

	headers := []string{"id", "name", "category", "price", "create_date"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, price := range prices {
		record := []string{
			fmt.Sprintf("%d", price.ID),
			price.Name,
			price.Category,
			fmt.Sprintf("%.2f", price.Price),
			price.CreateDate.Format(time.DateOnly),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
