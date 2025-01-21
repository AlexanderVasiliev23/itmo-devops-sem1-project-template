package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"project_sem/internal/config"
	"project_sem/internal/migrate"
)

func ConnectToDB(pgConfig config.PGConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", pgConfig.User, pgConfig.Pass, pgConfig.Name)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrate.Migrate(db, migrate.Migrations); err != nil {
		return nil, fmt.Errorf("ошибка миграции базы данных: %w", err)
	}

	return db, nil
}
