package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

//go:embed migrations
var Migrations embed.FS

func Migrate(db *sql.DB, path fs.FS) error {
	goose.SetBaseFS(path)

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up migrations: %w", err)
	}

	return nil
}
