package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //This is required to use postgres with gorm
)

var db *gorm.DB

//GetTestDB returns a connection to the test database
func GetTestDB() *gorm.DB {
	if db == nil {
		connStr := "host=localhost user=khan_test port=5432 sslmode=disable dbname=khan_test"
		db, _ = gorm.Open("postgres", connStr)
	}

	return db
}
