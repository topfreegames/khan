// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	"github.com/topfreegames/khan/es"
	"github.com/topfreegames/khan/models"

	"github.com/uber-go/zap"
)

func GetTestES() *es.Client {
	return es.GetTestClient("localhost", 9200, "khan", false, zap.New(zap.NewJSONEncoder(), zap.ErrorLevel), false)
}

func DestroyTestES() {
	es.DestroyClient()
}

// GetTestDB returns a connection to the test database
func GetTestDB() (models.DB, error) {
	return models.GetDB("localhost", "khan_test", 5433, "disable", "khan_test", "")
}

// GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB() models.DB {
	faultyDb, _ := models.InitDb("localhost", "khan_tet", 5433, "disable", "khan_test", "")
	return faultyDb
}
