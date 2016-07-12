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
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
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
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	AssertNotError(g, err)

	g.Describe("Create Clan Handler", func() {
		g.It("Should create clan", func() {
			_, player, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
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
			_, player, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			clanPublicID := randomdata.FullName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("name is required, ownerPublicID is required, metadata is required"))
		})

		g.It("Should not create clan if invalid payload", func() {
			gameID := "gameID"
			a := GetDefaultTestApp()
			res := PostBody(a, GetGameRoute(gameID, "/clans"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create clan if owner does not exist", func() {
			game, _, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			payload := map[string]interface{}{
				"publicID":         randomdata.FullName(randomdata.RandomGender),
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, GetGameRoute(game.PublicID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", payload["ownerPublicID"]))
		})

		g.It("Should not create clan if invalid data", func() {
			_, player, err := models.CreatePlayerFactory(testDb, "")
			AssertNotError(g, err)

			payload := map[string]interface{}{
				"publicID":         randomdata.FullName(randomdata.RandomGender),
				"name":             strings.Repeat("a", 256),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(255)")
		})
	})

	g.Describe("Leave Clan Handler", func() {
		g.It("Should leave a clan and transfer ownership", func() {
			_, clan, _, _, memberships, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/leave", clan.PublicID))
			a := GetDefaultTestApp()
			res := PostJSON(a, route, t, map[string]interface{}{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.OwnerID).Equal(memberships[0].PlayerID)
		})

		g.It("Should not leave a clan if invalid clan", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/leave", "random-id"))
			a := GetDefaultTestApp()
			res := PostBody(a, route, t, "")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "Clan was not found with id: random-id")).IsTrue()
		})
	})

	g.Describe("Transfer Clan Ownership Handler", func() {
		g.It("Should transfer a clan ownership", func() {
			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()
			ownerPublicID := owner.PublicID
			playerPublicID := players[0].PublicID

			payload := map[string]interface{}{
				"ownerPublicID":  ownerPublicID,
				"playerPublicID": playerPublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/transfer-ownership", clan.PublicID))
			a := GetDefaultTestApp()
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.OwnerID).Equal(players[0].ID)
		})

		g.It("Should not transfer a clan ownership if missing parameters", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "public-id"))
			a := GetDefaultTestApp()
			res := PostJSON(a, route, t, map[string]interface{}{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("playerPublicID is required")

		})

		g.It("Should not transfer a clan ownership if invalid payload", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "random-id"))
			a := GetDefaultTestApp()
			res := PostBody(a, route, t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})
	})

	g.Describe("Update Clan Handler", func() {
		g.It("Should update clan", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := owner.PublicID
			metadata := map[string]interface{}{"new": "metadata"}

			payload := map[string]interface{}{
				"name":             clanName,
				"ownerPublicID":    ownerPublicID,
				"metadata":         metadata,
				"allowApplication": !clan.AllowApplication,
				"autoJoin":         !clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, gameID, publicID)
			AssertNotError(g, err)
			g.Assert(dbClan.GameID).Equal(gameID)
			g.Assert(dbClan.PublicID).Equal(publicID)
			g.Assert(dbClan.Name).Equal(clanName)
			g.Assert(dbClan.OwnerID).Equal(owner.ID)
			g.Assert(dbClan.Metadata).Equal(metadata)
			g.Assert(dbClan.AllowApplication).Equal(!clan.AllowApplication)
			g.Assert(dbClan.AutoJoin).Equal(!clan.AutoJoin)
		})

		g.It("Should not update clan if missing parameters", func() {
			route := GetGameRoute("gameID", fmt.Sprintf("/clans/%s", "publicID"))
			a := GetDefaultTestApp()
			res := PutJSON(a, route, t, map[string]interface{}{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("name is required, ownerPublicID is required, metadata is required, allowApplication is required, autoJoin is required")
		})

		g.It("Should not update clan if invalid payload", func() {
			route := GetGameRoute("game-id", fmt.Sprintf("/clans/%s", "random-id"))
			a := GetDefaultTestApp()

			res := PutBody(a, route, t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not update clan if player is not the owner", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			metadata := map[string]interface{}{"x": 1}

			payload := map[string]interface{}{
				"name":             clan.Name,
				"ownerPublicID":    randomdata.FullName(randomdata.RandomGender),
				"metadata":         metadata,
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()

			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})

		g.It("Should not update clan if invalid data", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{
				"name":             strings.Repeat("s", 256),
				"ownerPublicID":    owner.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()

			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result map[string]interface{}
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
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()
			for index, clanObj := range result["clans"].([]interface{}) {
				clan := clanObj.(map[string]interface{}) // Can't be map[string]interface{}
				g.Assert(clan["name"]).Equal(expectedClans[index].Name)
				clanMetadata := clan["metadata"].(map[string]interface{})
				for k, v := range clanMetadata {
					g.Assert(v).Equal(expectedClans[index].Metadata[k])
				}
				g.Assert(clan["publicID"]).Equal(expectedClans[index].PublicID)
				g.Assert(clan["ID"]).Equal(nil)
				g.Assert(clan["autoJoin"]).Equal(expectedClans[index].AutoJoin)
				g.Assert(clan["allowApplication"]).Equal(expectedClans[index].AllowApplication)
			}
		})

		g.It("Should return empty list if invalid game query", func() {
			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute("invalid-query-game-id", "/clans"), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()
			g.Assert(len(result["clans"].([]interface{}))).Equal(0)
		})
	})

	g.Describe("Retrieve Clan Handler", func() {
		g.It("Should get details for clan", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			AssertNotError(g, err)

			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
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
			_, clan, _, _, _, err := models.GetClanWithMemberships(
				testDb, 10, 0, 0, 0, "clan-details-api", "clan-details-api-clan",
			)
			AssertNotError(g, err)

			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			g.Assert(result["roster"] == nil).IsFalse()
		})
	})

	g.Describe("Retrieve Clan Summary Handler", func() {
		g.It("Should get details for clan", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			AssertNotError(g, err)

			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s/summary", clan.PublicID)), t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			g.Assert(int(result["membershipCount"].(float64))).Equal(clan.MembershipCount)
			g.Assert(result["publicID"]).Equal(clan.PublicID)
			g.Assert(result["name"]).Equal(clan.Name)
			g.Assert(result["allowApplication"]).Equal(clan.AllowApplication)
			g.Assert(result["autoJoin"]).Equal(clan.AutoJoin)
			resultMetadata := result["metadata"].(map[string]interface{})
			for k, v := range resultMetadata {
				g.Assert(v).Equal(clan.Metadata[k])
			}
		})

		g.It("Should not get details for clan that does not exist", func() {
			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute("game-id", "/clans/dont-exist/summary"), t)
			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("Clan was not found with id: dont-exist")
		})
	})

	g.Describe("Search Clan Handler", func() {
		g.It("Should search for a clan", func() {
			player, expectedClans, err := models.GetTestClans(
				testDb, "", "clan-apisearch-clan", 10,
			)
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute(player.GameID, "clan-search"), t, map[string]interface{}{
				"term": "APISEARCH",
			})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			clans := result["clans"].([]interface{})
			g.Assert(len(clans)).Equal(10)

			for index, expectedClan := range expectedClans {
				clan := clans[index].(map[string]interface{}) // Can't be map[string]interface{}
				g.Assert(clan["name"]).Equal(expectedClan.Name)
				g.Assert(int(clan["membershipCount"].(float64))).Equal(expectedClan.MembershipCount)
				g.Assert(clan["autoJoin"]).Equal(expectedClan.AutoJoin)
				g.Assert(clan["allowApplication"]).Equal(expectedClan.AllowApplication)
			}
		})

		g.It("Should unicode search for a clan", func() {
			player, expectedClans, err := models.GetTestClans(
				testDb, "clan-apisearch-game", "clan-apisearch-clan", 10,
			)
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()

			res := Get(a, GetGameRoute(player.GameID, "clan-search"), t, map[string]interface{}{
				"term": "💩clán-clan-APISEARCH",
			})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			clans := result["clans"].([]interface{})
			g.Assert(len(clans)).Equal(10)

			for index, expectedClan := range expectedClans {
				clan := clans[index].(map[string]interface{}) // Can't be map[string]interface{}
				g.Assert(clan["name"]).Equal(expectedClan.Name)
				g.Assert(int(clan["membershipCount"].(float64))).Equal(expectedClan.MembershipCount)
				g.Assert(clan["autoJoin"]).Equal(expectedClan.AutoJoin)
				g.Assert(clan["allowApplication"]).Equal(expectedClan.AllowApplication)
			}
		})

	})

	g.Describe("Clan Hooks", func() {
		g.It("Should call create clan hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clancreated",
			}, models.ClanCreatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/clancreated"}, 52525)

			_, player, err := models.CreatePlayerFactory(testDb, hooks[0].GameID, true)
			AssertNotError(g, err)

			clanPublicID := uuid.NewV4().String()
			payload := map[string]interface{}{
				"publicID":         clanPublicID,
				"name":             randomdata.FullName(randomdata.RandomGender),
				"ownerPublicID":    player.PublicID,
				"metadata":         map[string]interface{}{"x": 1},
				"allowApplication": true,
				"autoJoin":         true,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, GetGameRoute(player.GameID, "/clans"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["publicID"]).Equal(clanPublicID)

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			clan := (*responses)[0]["payload"].(map[string]interface{})
			g.Assert(clan["gameID"]).Equal(player.GameID)
			g.Assert(clan["publicID"]).Equal(payload["publicID"])
			g.Assert(clan["name"]).Equal(payload["name"])
			g.Assert(str(clan["membershipCount"])).Equal("1")
			g.Assert(clan["allowApplication"]).Equal(payload["allowApplication"])
			g.Assert(clan["autoJoin"]).Equal(payload["autoJoin"])
			clanMetadata := clan["metadata"].(map[string]interface{})
			metadata := payload["metadata"].(map[string]interface{})
			for k, v := range clanMetadata {
				g.Assert(str(v)).Equal(str(metadata[k]))
			}
		})

		g.It("Should call update clan hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clanupdated",
			}, models.ClanUpdatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/clanupdated"}, 52525)

			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := owner.PublicID
			metadata := map[string]interface{}{"new": "metadata"}

			payload := map[string]interface{}{
				"name":             clanName,
				"ownerPublicID":    ownerPublicID,
				"metadata":         metadata,
				"allowApplication": clan.AllowApplication,
				"autoJoin":         clan.AutoJoin,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			a := GetDefaultTestApp()
			res := PutJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			rClan := (*responses)[0]["payload"].(map[string]interface{})
			g.Assert(rClan["gameID"]).Equal(hooks[0].GameID)
			g.Assert(rClan["publicID"]).Equal(publicID)
			g.Assert(rClan["name"]).Equal(payload["name"])
			g.Assert(str(rClan["membershipCount"])).Equal("1")
			g.Assert(rClan["allowApplication"]).Equal(payload["allowApplication"])
			g.Assert(rClan["autoJoin"]).Equal(payload["autoJoin"])
			clanMetadata := rClan["metadata"].(map[string]interface{})
			metadata = payload["metadata"].(map[string]interface{})
			for k, v := range clanMetadata {
				g.Assert(str(v)).Equal(str(metadata[k]))
			}
		})

		g.It("Should call leave clan hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clanleave",
			}, models.ClanLeftHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/clanleave"}, 52525)

			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s/leave", publicID))
			a := GetDefaultTestApp()
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			rClan := (*responses)[0]["payload"].(map[string]interface{})
			g.Assert(rClan["gameID"]).Equal(hooks[0].GameID)
			g.Assert(int(rClan["type"].(float64))).Equal(5)

			clanDetails := rClan["clan"].(map[string]interface{})
			g.Assert(clanDetails["publicID"]).Equal(clan.PublicID)
			g.Assert(clanDetails["name"]).Equal(clan.Name)
			g.Assert(str(clanDetails["membershipCount"])).Equal("1")
			g.Assert(clanDetails["allowApplication"]).Equal(clan.AllowApplication)
			g.Assert(clanDetails["autoJoin"]).Equal(clan.AutoJoin)

			newOwner := players[0]
			ownerDetails := rClan["newOwner"].(map[string]interface{})
			g.Assert(ownerDetails["publicID"]).Equal(newOwner.PublicID)
			g.Assert(ownerDetails["name"]).Equal(newOwner.Name)
			g.Assert(str(ownerDetails["membershipCount"])).Equal("0")
			g.Assert(str(ownerDetails["ownershipCount"])).Equal("1")
		})

		g.It("Should call transfer ownership hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/clantransfer",
			}, models.ClanOwnershipTransferredHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/clantransfer"}, 52525)

			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID": players[0].PublicID,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s/transfer-ownership", publicID))
			a := GetDefaultTestApp()
			res := PostJSON(a, route, t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			rClan := (*responses)[0]["payload"].(map[string]interface{})
			g.Assert(rClan["gameID"]).Equal(hooks[0].GameID)

			clanDetails := rClan["clan"].(map[string]interface{})
			g.Assert(clanDetails["publicID"]).Equal(clan.PublicID)
			g.Assert(clanDetails["name"]).Equal(clan.Name)
			g.Assert(str(clanDetails["membershipCount"])).Equal("1")
			g.Assert(clanDetails["allowApplication"]).Equal(clan.AllowApplication)
			g.Assert(clanDetails["autoJoin"]).Equal(clan.AutoJoin)

			newOwner := players[0]
			ownerDetails := rClan["newOwner"].(map[string]interface{})
			g.Assert(ownerDetails["publicID"]).Equal(newOwner.PublicID)
			g.Assert(ownerDetails["name"]).Equal(newOwner.Name)
			g.Assert(str(ownerDetails["membershipCount"])).Equal("0")
			g.Assert(str(ownerDetails["ownershipCount"])).Equal("1")
		})
	})
}
