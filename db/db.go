package db

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DBName        = "hugobot.sqlite"
	DBPragma      = ` PRAGMA foreign_keys = ON; `
	DBBasePathEnv = "HUGOBOT_DB_PATH"
)

var (
	DBOptions = map[string]string{
		"_journal_mode": "WAL",
	}

	DB *Database
)

type Database struct {
	Handle *sqlx.DB
}

func (d *Database) Open() error {

	dsnOptions := &url.Values{}
	for k, v := range DBOptions {
		dsnOptions.Set(k, v)
	}

	// Get db base path
	path, set := os.LookupEnv(DBBasePathEnv)
	if !set {
		path = "."
	}
	path = filepath.Join(path, DBName)
	//path = fmt.Sprintf("%s/%s", path, DBName)

	dsn := fmt.Sprintf("file:%s?%s", path, dsnOptions.Encode())

	log.Printf("Opening sqlite db %s\n", dsn)

	var err error
	d.Handle, err = sqlx.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Execute Pragmas
	d.Handle.MustExec(DBPragma)

	return nil
}

type AutoIncr struct {
	ID      int64     `json:"id"`
	Created time.Time `json:"created"`
}

func init() {
	DB = &Database{}
	DB.Open()
}
