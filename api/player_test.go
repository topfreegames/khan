// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/models"
)

func TestPlayerHandler(t *testing.T) {
	g := Goblin(t)

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Create Player Handler", func() {
		g.It("Should create player", func() {
			gameID := "api-cr-1"
			playerID := randomdata.FullName(randomdata.RandomGender)
			playerName := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]string{
				"gameID":   gameID,
				"playerID": playerID,
				"name":     playerName,
				"metadata": metadata,
			}
			res := PostJSON(a, "/players", t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbPlayer, err := models.GetPlayerByPlayerID(gameID, playerID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbPlayer.GameID).Equal(gameID)
			g.Assert(dbPlayer.PlayerID).Equal(playerID)
			g.Assert(dbPlayer.Name).Equal(playerName)
			g.Assert(dbPlayer.Metadata).Equal(metadata)
		})
	})
}
