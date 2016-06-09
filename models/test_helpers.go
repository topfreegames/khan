package models

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" //This is required to use postgres with database/sql
	"gopkg.in/gorp.v1"
)

var db *gorp.DbMap

//GetTestDB returns a connection to the test database
func GetTestDB() *gorp.DbMap {
	if db == nil {
		db = InitDb()
	}

	return db
}

//InitDb initializes a connection to the database
func InitDb() *gorp.DbMap {
	connStr := "host=localhost user=khan_test port=5432 sslmode=disable dbname=khan_test"
	db, err := sql.Open("postgres", connStr)
	checkErr(err, "sql.Open failed")

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbmap.AddTableWithName(Player{}, "players").SetKeys(true, "ID")
	dbmap.AddTableWithName(Clan{}, "clans").SetKeys(true, "ID")
	dbmap.AddTableWithName(Membership{}, "memberships").SetKeys(true, "ID")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
