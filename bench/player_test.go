// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"fmt"
	"net/http"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
	khanTesting "github.com/topfreegames/khan/testing"
)

var playerResult *http.Response

func BenchmarkCreatePlayer(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, _, err := getGameAndPlayer(db, mongoDB)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/players", game.PublicID))
		res, err := postTo(route, getPlayerPayload(uuid.NewV4().String()))
		validateResp(res, err)
		res.Body.Close()

		playerResult = res
	}
}

func BenchmarkUpdatePlayer(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, _, err := getGameAndPlayer(db, mongoDB)
	if err != nil {
		panic(err.Error())
	}

	var players []*models.Player
	for i := 0; i < b.N; i++ {
		player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
			"GameID": game.PublicID,
		}).(*models.Player)
		err = db.Insert(player)
		if err != nil {
			panic(err.Error())
		}
		players = append(players, player)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		playerPublicID := players[i].PublicID
		route := getRoute(fmt.Sprintf("/games/%s/players/%s", game.PublicID, playerPublicID))
		res, err := putTo(route, getPlayerPayload(playerPublicID))
		validateResp(res, err)
		res.Body.Close()

		playerResult = res
	}
}

func BenchmarkRetrievePlayer(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	gameID := uuid.NewV4().String()
	_, player, err := fixtures.GetTestPlayerWithMemberships(db, gameID, 50, 20, 30, 80)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/players/%s", gameID, player.PublicID))
		res, err := get(route)
		validateResp(res, err)
		res.Body.Close()

		playerResult = res
	}
}
