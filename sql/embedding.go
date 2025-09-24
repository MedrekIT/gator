package embedding

import (
	"embed"
	"database/sql"
	"github.com/pressly/goose/v3"
)

//go:embed schema/*.sql
var embedMigrations embed.FS

func DbEmbedding(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "schema"); err != nil {
		return err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	return nil
}
