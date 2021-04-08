// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	egorp "github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/khan/models"
)

// GetTestDB returns a connection to the test database
func GetTestDB() (egorp.Database, error) {
	return models.GetDB("localhost", "khan_test", 5433, "disable", "khan_test", "")
}

// GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB() models.DB {
	faultyDb, _ := models.InitDb("localhost", "khan_tet", 5433, "disable", "khan_test", "")
	return faultyDb
}
