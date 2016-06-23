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
	"sort"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

// AssertError asserts that the specified error is not nil
func AssertError(g *G, err error) {
	g.Assert(err == nil).IsFalse("Expected error to exist, but it was nil")
}

// AssertNotError asserts that the specified error is nil
func AssertNotError(g *G, err error) {
	if err != nil {
		g.Assert(err == nil).IsTrue(err.Error())
	}
}

func TestClanHandler(t *testing.T) {
	t.Parallel()
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	AssertNotError(g, err)

	g.Describe("Create Clan Handler", func() {
		g.It("Should create clan", func() {
			player, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := util.JSON{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         util.JSON{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["publicID"]).Equal(clanPublicID)

			dbClan, err := models.GetClanByPublicID(a.Db, player.GameID, clanPublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.GameID).Equal(player.GameID)
			g.Assert(dbClan.OwnerID).Equal(player.ID)
			g.Assert(dbClan.PublicID).Equal(payload["publicID"])
			g.Assert(dbClan.Name).Equal(payload["name"])
			g.Assert(dbClan.Metadata).Equal(payload["metadata"])
			g.Assert(dbClan.AllowApplication).Equal(payload["allowApplication"])
			g.Assert(dbClan.AutoJoin).Equal(payload["autoJoin"])
		})

		g.It("Should not create clan if missing parameters", func() {
			player, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := util.JSON{
				"publicID":         clanPublicID,
				"allowApplication": true,
				"autoJoin":         true,
			}
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("name is required, ownerPublicID is required, metadata is required"))
		})

		g.It("Should not create clan if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			res := PostBody(a, GetGameRoute(gameID, "/clans"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create clan if owner does not exist", func() {
			gameID := randomdata.FullName(randomdata.RandomGender)
			a := GetDefaultTestApp()
			payload := util.JSON{
				"publicID":         randomdata.FullName(randomdata.RandomGender),
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         util.JSON{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			res := PostJSON(a, GetGameRoute(gameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", payload["ownerPublicID"]))
		})

		g.It("Should not create clan if invalid data", func() {
			player, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			payload := util.JSON{
				"publicID":         randomdata.FullName(randomdata.RandomGender),
				"name":             strings.Repeat("a", 256),
				"ownerPublicID":    player.PublicID,
				"metadata":         util.JSON{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(255)")
		})
	})

	g.Describe("Leave Clan Handler", func() {
		g.It("Should leave a clan and transfer ownership", func() {
			clan, owner, _, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()
			ownerPublicID := owner.PublicID

			a := GetDefaultTestApp()
			payload := util.JSON{
				"ownerPublicID": ownerPublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/leave", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.OwnerID).Equal(memberships[0].PlayerID)
		})

		g.It("Should not leave a clan if missing parameters", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/leave", "random-id"))
			res := PostJSON(a, route, t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("ownerPublicID is required")
		})

		g.It("Should not leave a clan if invalid payload", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/leave", "random-id"))
			res := PostBody(a, route, t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not leave a clan if player is not the owner", func() {
			clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()
			payload := util.JSON{
				"ownerPublicID": players[0].PublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/leave", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})
	})

	g.Describe("Transfer Clan Ownership Handler", func() {
		g.It("Should transfer a clan ownership", func() {
			clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()
			ownerPublicID := owner.PublicID
			playerPublicID := players[0].PublicID

			a := GetDefaultTestApp()
			payload := util.JSON{
				"ownerPublicID":  ownerPublicID,
				"playerPublicID": playerPublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/transfer-ownership", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.OwnerID).Equal(players[0].ID)
		})

		g.It("Should not transfer a clan ownership if missing parameters", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "public-id"))
			res := PostJSON(a, route, t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("ownerPublicID is required, playerPublicID is required")

		})

		g.It("Should not transfer a clan ownership if invalid payload", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "random-id"))
			res := PostBody(a, route, t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not transfer a clan ownership if player is not the owner", func() {
			clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()
			payload := util.JSON{
				"ownerPublicID":  players[0].PublicID,
				"playerPublicID": players[0].PublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/transfer-ownership", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})
	})

	g.Describe("Update Clan Handler", func() {
		g.It("Should update clan", func() {
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := owner.PublicID
			metadata := util.JSON{"new": "metadata"}

			a := GetDefaultTestApp()
			payload := util.JSON{
				"name":             clanName,
				"ownerPublicID":    ownerPublicID,
				"metadata":         metadata,
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, gameID, publicID)
			AssertNotError(g, err)
			g.Assert(dbClan.GameID).Equal(gameID)
			g.Assert(dbClan.PublicID).Equal(publicID)
			g.Assert(dbClan.Name).Equal(clanName)
			g.Assert(dbClan.OwnerID).Equal(owner.ID)
			g.Assert(dbClan.Metadata).Equal(metadata)
		})

		g.It("Should not update clan if missing parameters", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("gameID", fmt.Sprintf("/clans/%s", "publicID"))
			res := PutJSON(a, route, t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("name is required, ownerPublicID is required, metadata is required, allowApplication is required, autoJoin is required")
		})

		g.It("Should not update clan if invalid payload", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("/clans/%s", "random-id"))
			res := PutBody(a, route, t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not update clan if player is not the owner", func() {
			clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			metadata := util.JSON{"x": 1}

			a := GetDefaultTestApp()
			payload := util.JSON{
				"name":             clan.Name,
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         metadata,
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})

		g.It("Should not update clan if invalid data", func() {
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := util.JSON{
				"name":             strings.Repeat("s", 256),
				"ownerPublicID":    owner.PublicID,
				"metadata":         util.JSON{"x": 1},
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(255)")
		})
	})

	g.Describe("List All Clans Handler", func() {
		g.It("Should get all clans", func() {
			player, expectedClans, err := models.GetTestClans(testDb, "", "", 10)
			AssertNotError(g, err)
			sort.Sort(models.ClanByName(expectedClans))

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(player.GameID, "/clans"), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()
			for index, clanObj := range result["clans"].([]interface{}) {
				clan := clanObj.(map[string]interface{}) // Can't be util.JSON
				g.Assert(clan["name"]).Equal(expectedClans[index].Name)
				clanMetadata := clan["metadata"].(map[string]interface{})
				for k, v := range clanMetadata {
					g.Assert(v).Equal(expectedClans[index].Metadata[k])
				}
				g.Assert(clan["publicID"]).Equal(expectedClans[index].PublicID)
				g.Assert(clan["ID"]).Equal(nil)
			}
		})

		g.It("Should return empty list if invalid game query", func() {
			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute("invalid-query-game-id", "/clans"), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()
			g.Assert(len(result["clans"].([]interface{}))).Equal(0)
		})
	})
	g.Describe("Retrieve Clan Handler", func() {
		g.It("Should get details for clan", func() {
			clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			g.Assert(result["name"]).Equal(clan.Name)
			resultMetadata := result["metadata"].(map[string]interface{})
			for k, v := range resultMetadata {
				g.Assert(v).Equal(clan.Metadata[k])
			}
			g.Assert(result["publicID"]).Equal(nil)
		})

		g.It("Should get clan members", func() {
			clan, _, _, _, err := models.GetClanWithMemberships(
				testDb, 10, "clan-details-api", "clan-details-api-clan",
			)
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			g.Assert(result["members"] == nil).IsFalse()
		})
	})
	g.Describe("Search Clan Handler", func() {
		g.It("Should search for a clan", func() {
			player, expectedClans, err := models.GetTestClans(
				testDb, "", "clan-apisearch-clan", 10,
			)
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(player.GameID, "clan-search?term=APISEARCH"), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			clans := result["clans"].([]interface{})
			g.Assert(len(clans)).Equal(10)

			for index, expectedClan := range expectedClans {
				clan := clans[index].(map[string]interface{}) // Can't be util.JSON
				g.Assert(clan["name"]).Equal(expectedClan.Name)
			}
		})
	})
}
