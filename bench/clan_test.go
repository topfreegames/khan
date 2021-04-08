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
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
	khanTesting "github.com/topfreegames/khan/testing"
)

var result *http.Response

func BenchmarkCreateClan(b *testing.B) {
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
		route := getRoute(fmt.Sprintf("/games/%s/clans", game.PublicID))
		res, err := postTo(route, getClanPayload(players[i].PublicID, uuid.NewV4().String()))
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkUpdateClan(b *testing.B) {
	fixtures.ConfigureAndStartGoWorkers()

	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, owner, err := getGameAndPlayer(db, mongoDB)
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
	_, clan, _, _, _, err := fixtures.GetClanWithMemberships(
		db, 50, 50, 50, 50, gameID, uuid.NewV4().String(),
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

func BenchmarkRetrieveClanSummary(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	gameID := uuid.NewV4().String()
	_, clan, _, _, _, err := fixtures.GetClanWithMemberships(
		db, 50, 50, 50, 50, gameID, uuid.NewV4().String(),
	)

	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/summary", gameID, clan.PublicID))
		res, err := get(route)
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkRetrieveClansSummary(b *testing.B) {
	fixtures.ConfigureAndStartGoWorkers()

	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, owner, err := getGameAndPlayer(db, mongoDB)
	if err != nil {
		panic(err.Error())
	}

	clans, err := createClans(db, game, owner, 500)
	if err != nil {
		panic(err.Error())
	}

	var clanIDs []string
	for _, clan := range clans {
		clanIDs = append(clanIDs, clan.PublicID)
	}
	qs := strings.Join(clanIDs, ",")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans-summary?clanPublicIds=%s", game.PublicID, qs))
		res, err := get(route)
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkSearchClan(b *testing.B) {
	fixtures.ConfigureAndStartGoWorkers()

	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, owner, err := getGameAndPlayer(db, mongoDB)
	if err != nil {
		panic(err.Error())
	}

	clans, err := createClans(db, game, owner, b.N)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans/search?term=%s", game.PublicID, clans[0].PublicID))
		res, err := get(route)
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkListClans(b *testing.B) {
	fixtures.ConfigureAndStartGoWorkers()

	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, owner, err := getGameAndPlayer(db, mongoDB)
	if err != nil {
		panic(err.Error())
	}

	_, err = createClans(db, game, owner, b.N)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans", game.Name))
		res, err := get(route)
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkLeaveClan(b *testing.B) {
	fixtures.ConfigureAndStartGoWorkers()

	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	mongoDB, err := khanTesting.GetTestMongo()
	if err != nil {
		panic(err.Error())
	}

	game, owner, err := getGameAndPlayer(db, mongoDB)
	if err != nil {
		panic(err.Error())
	}

	clans, err := createClans(db, game, owner, b.N)
	if err != nil {
		panic(err.Error())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/leave", game.PublicID, clans[i].PublicID))
		res, err := postTo(route, map[string]interface{}{
			"ownerPublicID": owner.PublicID,
		})
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}

func BenchmarkTransferOwnership(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	game, clan, owner, members, _, err := fixtures.GetClanWithMemberships(
		db, 20, 0, 0, 0, "", "",
	)
	if err != nil {
		panic(err.Error())
	}

	player1 := owner.PublicID
	player2 := members[0].PublicID

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s/clans/%s/transfer-ownership", game.PublicID, clan.PublicID))
		res, err := postTo(route, map[string]interface{}{
			"ownerPublicID":  player1,
			"playerPublicID": player2,
		})
		validateResp(res, err)
		res.Body.Close()

		altPlayer := player1
		player1 = player2
		player2 = altPlayer

		result = res
	}
}
