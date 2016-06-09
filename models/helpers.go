package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" //This is required to use postgres with database/sql
	"gopkg.in/gorp.v1"
)

var db *gorp.DbMap

//GetTestDB returns a connection to the test database
func GetTestDB() *gorp.DbMap {
	return GetDB("localhost", "khan_test", 5432, "disable", "khan_test", "")
}

func GetDB(host string, user string, port int, sslmode string, dbName string, password string) *gorp.DbMap {
	if db == nil {
		db = InitDb(host, user, port, sslmode, dbName, password)
	}

	return db
}

//InitDb initializes a connection to the database
func InitDb(host string, user string, port int, sslmode string, dbName string, password string) *gorp.DbMap {
	connStr := fmt.Sprintf(
		"host=%s user=%s port=%d sslmode=%s dbname=%s",
		host, user, port, sslmode, dbName,
	)
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}
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
