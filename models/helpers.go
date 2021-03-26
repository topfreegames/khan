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

	"github.com/go-gorp/gorp"
	_ "github.com/lib/pq" //This is required to use postgres with database/sql
	egorp "github.com/topfreegames/extensions/v9/gorp"
	"github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/khan/util"
)

// DB is the contract for all the operations we use from either a connection or transaction
// This is required for automatic transactions
type DB interface {
	Get(interface{}, ...interface{}) (interface{}, error)
	Select(interface{}, string, ...interface{}) ([]interface{}, error)
	SelectOne(interface{}, string, ...interface{}) error
	SelectInt(string, ...interface{}) (int64, error)
	Insert(...interface{}) error
	Update(...interface{}) (int64, error)
	Delete(...interface{}) (int64, error)
	Exec(string, ...interface{}) (sql.Result, error)
}

var _db interfaces.Database

// GetDefaultDB returns a connection to the default database
func GetDefaultDB() (interfaces.Database, error) {
	return GetDB("localhost", "khan", 5433, "disable", "khan", "")
}

// GetPerfDB returns a connection to the perf database
func GetPerfDB() (interfaces.Database, error) {
	return GetDB("localhost", "khan_perf", 5433, "disable", "khan_perf", "")
}

// GetDB returns a DbMap connection to the database specified in the arguments
func GetDB(host string, user string, port int, sslmode string, dbName string, password string) (interfaces.Database, error) {
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

// InitDb initializes a connection to the database
func InitDb(host string, user string, port int, sslmode string, dbName string, password string) (interfaces.Database, error) {
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

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)

	dbmap := &gorp.DbMap{
		Db:            db,
		Dialect:       gorp.PostgresDialect{},
		TypeConverter: util.TypeConverter{},
	}

	dbmap.AddTableWithName(Game{}, "games").SetKeys(true, "ID")
	dbmap.AddTableWithName(Player{}, "players").SetKeys(true, "ID")
	dbmap.AddTableWithName(EncryptedPlayer{}, "encrypted_players")
	dbmap.AddTableWithName(Clan{}, "clans").SetKeys(true, "ID")
	dbmap.AddTableWithName(Membership{}, "memberships").SetKeys(true, "ID")
	dbmap.AddTableWithName(Hook{}, "hooks").SetKeys(true, "ID")

	// dbmap.TraceOn("[gorp]", log.New(os.Stdout, "KHAN:", log.Lmicroseconds))
	return egorp.New(dbmap, dbName), nil
}

// Returns value or 0
func nullOrInt(value sql.NullInt64) int64 {
	if value.Valid {
		v, err := value.Value()
		if err == nil {
			return v.(int64)
		}
	}
	return 0
}

// Returns value or ""
func nullOrString(value sql.NullString) string {
	if value.Valid {
		v, err := value.Value()
		if err == nil {
			return v.(string)
		}
	}
	return ""
}

// Returns value or false
func nullOrBool(value sql.NullBool) bool {
	if value.Valid {
		v, err := value.Value()
		if err == nil {
			return v.(bool)
		}
	}
	return false
}
