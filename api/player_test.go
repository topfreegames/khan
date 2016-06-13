// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
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
			publicID := randomdata.FullName(randomdata.RandomGender)
			playerName := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":   gameID,
				"publicID": publicID,
				"name":     playerName,
				"metadata": metadata,
			}
			res := PostJSON(a, GetGameRoute(gameID, "/players"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbPlayer, err := models.GetPlayerByPublicID(a.Db, gameID, publicID)
			AssertNotError(g, err)
			g.Assert(dbPlayer.GameID).Equal(gameID)
			g.Assert(dbPlayer.PublicID).Equal(publicID)
			g.Assert(dbPlayer.Name).Equal(playerName)
			g.Assert(dbPlayer.Metadata).Equal(metadata)
		})

		g.It("Should not create player if invalid payload", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players")
			res := PostBody(a, route, t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not create player if invalid data", func() {
			gameID := "game-id-is-too-large-for-this-field-should-be-less-than-36-chars"
			publicID := randomdata.FullName(randomdata.RandomGender)
			playerName := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":   gameID,
				"publicID": publicID,
				"name":     playerName,
				"metadata": metadata,
			}
			res := PostJSON(a, GetGameRoute(gameID, "/players"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(36)")
		})
	})

	g.Describe("Update Player Handler", func() {
		g.It("Should update player", func() {
			a := GetDefaultTestApp()
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := a.Db.Insert(player)
			AssertNotError(g, err)

			metadata := "{\"y\": 10}"
			payload := map[string]interface{}{
				"gameID":   player.GameID,
				"publicID": player.PublicID,
				"name":     player.Name,
				"metadata": metadata,
			}

			route := GetGameRoute(player.GameID, fmt.Sprintf("/players/%s", player.PublicID))
			res := PutJSON(a, route, t, payload)
			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbPlayer, err := models.GetPlayerByPublicID(a.Db, player.GameID, player.PublicID)
			AssertNotError(g, err)
			g.Assert(dbPlayer.GameID).Equal(player.GameID)
			g.Assert(dbPlayer.PublicID).Equal(player.PublicID)
			g.Assert(dbPlayer.Name).Equal(player.Name)
			g.Assert(dbPlayer.Metadata).Equal(metadata)
		})

		g.It("Should not update player if invalid payload", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players/fake")
			res := PutBody(a, route, t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not update player if invalid data", func() {
			a := GetDefaultTestApp()

			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := a.Db.Insert(player)
			AssertNotError(g, err)

			metadata := ""

			payload := map[string]interface{}{
				"gameID":   player.GameID,
				"publicID": player.PublicID,
				"name":     player.Name,
				"metadata": metadata,
			}
			route := GetGameRoute("game-id", fmt.Sprintf("/players/%s", player.PublicID))
			res := PutJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})
}
