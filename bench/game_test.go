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

	"github.com/Pallinder/go-randomdata"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var gameResult *http.Response

func getGamePayload(publicID, name string) map[string]interface{} {
	if publicID == "" {
		publicID = randomdata.FullName(randomdata.RandomGender)
	}
	if name == "" {
		name = randomdata.FullName(randomdata.RandomGender)
	}
	return map[string]interface{}{
		"publicID":                      publicID,
		"name":                          name,
		"membershipLevels":              map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3},
		"metadata":                      map[string]interface{}{"x": "a"},
		"minLevelToAcceptApplication":   1,
		"minLevelToCreateInvitation":    1,
		"minLevelToRemoveMember":        1,
		"minLevelOffsetToRemoveMember":  1,
		"minLevelOffsetToPromoteMember": 1,
		"minLevelOffsetToDemoteMember":  1,
		"maxMembers":                    100,
		"maxClansPerPlayer":             1,
		"cooldownAfterDeny":             30,
		"cooldownAfterDelete":           30,
	}
}

func BenchmarkCreateGame(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gameID := uuid.NewV4().String()

		route := getRoute("/games")
		res, err := postTo(route, getGamePayload(gameID, gameID))
		validateResp(res, err)
		res.Body.Close()

		gameResult = res
	}
}

func BenchmarkUpdateGame(b *testing.B) {
	db, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	var games []*models.Game
	for i := 0; i < b.N; i++ {
		game := fixtures.GameFactory.MustCreateWithOption(map[string]interface{}{
			"PublicID":          uuid.NewV4().String(),
			"MaxClansPerPlayer": 999999,
		}).(*models.Game)
		err := db.Insert(game)
		if err != nil {
			panic(err.Error())
		}
		games = append(games, game)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/games/%s", games[i].PublicID))
		res, err := putTo(route, getGamePayload(games[i].PublicID, games[i].Name))
		validateResp(res, err)
		res.Body.Close()

		result = res
	}
}
