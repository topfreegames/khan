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
	"strings"
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
			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":   "api-cr-1",
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": "{\"x\": 1}",
			}
			res := PostJSON(a, GetGameRoute(payload["gameID"].(string), "/players"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbPlayer, err := models.GetPlayerByPublicID(
				a.Db, payload["gameID"].(string), payload["publicID"].(string),
			)
			AssertNotError(g, err)
			g.Assert(dbPlayer.GameID).Equal(payload["gameID"])
			g.Assert(dbPlayer.PublicID).Equal(payload["publicID"])
			g.Assert(dbPlayer.Name).Equal(payload["name"])
			g.Assert(dbPlayer.Metadata).Equal(payload["metadata"])
		})

		g.It("Should not create player if invalid payload", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players")
			res := PostBody(a, route, t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create player if invalid data", func() {
			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":   "game-id-is-too-large-for-this-field-should-be-less-than-36-chars",
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": "{\"x\": 1}",
			}
			res := PostJSON(a, GetGameRoute(payload["gameID"].(string), "/players"), t, payload)

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
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
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
