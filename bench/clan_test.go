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

	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

var result *http.Response

func BenchmarkCreateClan(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, _, err := getGameAndPlayer(db)
	if err != nil {
		panic(err.Error())
	}

	var players []*models.Player
	for i := 0; i < b.N; i++ {
		player := models.PlayerFactory.MustCreateWithOption(util.JSON{
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
		route := getRoute(fmt.Sprintf("/games/%s/clans", game.PublicID))
		res, err := postTo(route, getClanPayload(players[i].PublicID, uuid.NewV4().String()))
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkUpdateClan(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, owner, err := getGameAndPlayer(db)
	if err != nil {
		panic(err.Error())
	}

	clans, err := createClans(db, game, owner, b.N)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clanPublicID := clans[i].PublicID
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s", game.PublicID, clanPublicID))
		res, err := putTo(route, getClanPayload(owner.PublicID, clanPublicID))
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkRetrieveClan(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	gameID := uuid.NewV4().String()
	_, clan, _, _, _, err := models.GetClanWithMemberships(
		db, 50, 0, 0, 0, gameID, uuid.NewV4().String(),
	)

	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s", gameID, clan.PublicID))
		res, err := get(route)
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}
