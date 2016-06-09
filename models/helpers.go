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
func GetTestDB() (*gorp.DbMap, error) {
	return GetDB("localhost", "khan_test", 5432, "disable", "khan_test", "")
}

//GetDB returns a DbMap connection to the database specified in the arguments
func GetDB(host string, user string, port int, sslmode string, dbName string, password string) (*gorp.DbMap, error) {
	if db == nil {
		var err error
		db, err = InitDb(host, user, port, sslmode, dbName, password)
		if err != nil {
			db = nil
			return nil, err
		}
	}

	return db, nil
}

//InitDb initializes a connection to the database
func InitDb(host string, user string, port int, sslmode string, dbName string, password string) (*gorp.DbMap, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s port=%d sslmode=%s dbname=%s",
		host, user, port, sslmode, dbName,
	)
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbmap.AddTableWithName(Player{}, "players").SetKeys(true, "ID")
	dbmap.AddTableWithName(Clan{}, "clans").SetKeys(true, "ID")
	dbmap.AddTableWithName(Membership{}, "memberships").SetKeys(true, "ID")

	return dbmap, nil
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
