// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" //This is required to use postgres with database/sql
	"gopkg.in/gorp.v1"
)

//DB is the contract for all the operations we use from either a connection or transaction
//This is required for automatic transactions
type DB interface {
	Get(interface{}, ...interface{}) (interface{}, error)
	Select(interface{}, string, ...interface{}) ([]interface{}, error)
	SelectOne(interface{}, string, ...interface{}) error
	SelectInt(string, ...interface{}) (int64, error)
	Insert(...interface{}) error
	Update(...interface{}) (int64, error)
	Delete(...interface{}) (int64, error)
}

var _db DB

//GetTestDB returns a connection to the test database
func GetTestDB() (DB, error) {
	return GetDB("localhost", "khan_test", 5432, "disable", "khan_test", "")
}

//GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB() DB {
	faultyDb, _ := InitDb("localhost", "khan_tet", 5432, "disable", "khan_test", "")
	return faultyDb
}

//GetDB returns a DbMap connection to the database specified in the arguments
func GetDB(host string, user string, port int, sslmode string, dbName string, password string) (DB, error) {
	if _db == nil {
		var err error
		_db, err = InitDb(host, user, port, sslmode, dbName, password)
		if err != nil {
			_db = nil
			return nil, err
		}
	}

	return _db, nil
}

//InitDb initializes a connection to the database
func InitDb(host string, user string, port int, sslmode string, dbName string, password string) (DB, error) {
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
