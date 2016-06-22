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
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	"github.com/topfreegames/khan/models"
)

func TestPlayerHandler(t *testing.T) {
	g := Goblin(t)

	testDb, err := models.GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Create Player Handler", func() {
		g.It("Should create player", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			payload := map[string]interface{}{
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": "{\"x\": 1}",
			}
			res := PostJSON(a, GetGameRoute(game.PublicID, "/players"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["publicID"]).Equal(payload["publicID"].(string))

			dbPlayer, err := models.GetPlayerByPublicID(
				a.Db, game.PublicID, payload["publicID"].(string),
			)
			AssertNotError(g, err)
			g.Assert(dbPlayer.GameID).Equal(game.PublicID)
			g.Assert(dbPlayer.PublicID).Equal(payload["publicID"])
			g.Assert(dbPlayer.Name).Equal(payload["name"])
			g.Assert(dbPlayer.Metadata).Equal(payload["metadata"])
		})

		g.It("Should not create player if missing parameters", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players")
			res := PostJSON(a, route, t, map[string]interface{}{})

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("publicID is required, name is required, metadata is required")
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
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			payload := map[string]interface{}{
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": "metadata-it-not-a-json-and-will-break",
			}
			res := PostJSON(a, GetGameRoute(game.PublicID, "/players"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})

	g.Describe("Update Player Handler", func() {
		g.It("Should update player", func() {
			a := GetDefaultTestApp()
			player, err := models.CreatePlayerFactory(a.Db, "")
			AssertNotError(g, err)

			metadata := "{\"y\": 10}"
			payload := map[string]interface{}{
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

		g.It("Should not update player if missing parameters", func() {
			a := GetDefaultTestApp()
			route := GetGameRoute("game-id", "/players/player-id")
			res := PutJSON(a, route, t, map[string]interface{}{})

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("name is required, metadata is required")
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
			player, err := models.CreatePlayerFactory(a.Db, "")
			AssertNotError(g, err)

			metadata := ""

			payload := map[string]interface{}{
				"publicID": player.PublicID,
				"name":     player.Name,
				"metadata": metadata,
			}
			route := GetGameRoute(player.GameID, fmt.Sprintf("/players/%s", player.PublicID))
			res := PutJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})

	g.Describe("Player Hooks", func() {
		g.It("Should call create player hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/playercreated",
			}, models.PlayerCreatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/playercreated"}, 52525)

			app := GetDefaultTestApp()
			time.Sleep(time.Second)

			gameID := hooks[0].GameID
			payload := map[string]interface{}{
				"publicID": randomdata.FullName(randomdata.RandomGender),
				"name":     randomdata.FullName(randomdata.RandomGender),
				"metadata": "{\"x\": 1}",
			}
			res := PostJSON(app, GetGameRoute(gameID, "/players"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["publicID"]).Equal(payload["publicID"].(string))

			g.Assert(len(*responses)).Equal(1)

			player := (*responses)[0]
			g.Assert(player["gameID"]).Equal(gameID)
			g.Assert(player["publicID"]).Equal(payload["publicID"])
			g.Assert(player["name"]).Equal(payload["name"])
			g.Assert(player["metadata"]).Equal(payload["metadata"])
		})

		g.It("Should call update player hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/updated",
			}, models.PlayerUpdatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/updated"}, 52525)

			player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{"GameID": hooks[0].GameID}).(*models.Player)
			err = testDb.Insert(player)
			AssertNotError(g, err)

			app := GetDefaultTestApp()
			time.Sleep(time.Second)

			gameID := hooks[0].GameID
			payload := map[string]interface{}{
				"publicID": player.PublicID,
				"name":     player.Name,
				"metadata": player.Metadata,
			}
			res := PutJSON(app, GetGameRoute(gameID, fmt.Sprintf("/players/%s", player.PublicID)), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			g.Assert(len(*responses)).Equal(1)

			playerPayload := (*responses)[0]
			g.Assert(playerPayload["gameID"]).Equal(gameID)
			g.Assert(playerPayload["publicID"]).Equal(payload["publicID"])
			g.Assert(playerPayload["name"]).Equal(payload["name"])
			g.Assert(playerPayload["metadata"]).Equal(payload["metadata"])
		})
	})
}
