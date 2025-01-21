package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"project_sem/internal/config"
	"project_sem/internal/dataprovider"
	"project_sem/internal/entripoint/http/price/list"
	"project_sem/internal/entripoint/http/price/upload"
	"project_sem/internal/utils/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.ConnectToDB(cfg.PGConfig)
	if err != nil {
		log.Fatal(err)
	}

	priceProvider := dataprovider.NewPriceProvider(db)

	uploadEntrypoint := upload.NewUploadEntrypoint(priceProvider)
	listEntrypoint := list.NewListEntrypoint(priceProvider)

	r := mux.NewRouter()
	r.HandleFunc("/api/v0/prices", uploadEntrypoint.Handle).Methods(http.MethodPost)
	r.HandleFunc("/api/v0/prices", listEntrypoint.Handle).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServerPort), r))
}
