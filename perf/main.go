// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package main

import (
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

func createTestData(db models.DB) {
	game := models.GameFactory.MustCreateWithOption(util.JSON{
		"PublicID": "default-game",
	}).(*models.Game)
	db.Insert(game)

	owner := models.PlayerFactory.MustCreateWithOption(util.JSON{
		"GameID":   game.PublicID,
		"PublicID": "default-owner",
	}).(*models.Player)
	db.Insert(owner)
}

func main() {
	testDb, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	createTestData(testDb)
}
