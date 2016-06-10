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

func TestClanHandler(t *testing.T) {
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	g.Assert(err == nil).IsTrue()

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Create Clan Handler", func() {
		g.It("Should create clan", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			publicID := randomdata.FullName(randomdata.RandomGender)
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PostJSON(a, "/clans", t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(gameID, publicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbClan.GameID).Equal(gameID)
			g.Assert(dbClan.PublicID).Equal(publicID)
			g.Assert(dbClan.Name).Equal(clanName)
			g.Assert(dbClan.OwnerID).Equal(player.ID)
			g.Assert(dbClan.Metadata).Equal(metadata)
		})

		g.It("Should not create clan if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PostBody(a, "/clans", t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not create clan if owner does not exist", func() {
			gameID := randomdata.FullName(randomdata.RandomGender)
			publicID := randomdata.FullName(randomdata.RandomGender)
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PostJSON(a, "/clans", t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", ownerPublicID))
		})

		g.It("Should not create clan if invalid data", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			publicID := randomdata.FullName(randomdata.RandomGender)
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "it-will-fail-beacause-metada-is-not-a-json"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"PublicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PostJSON(a, "/clans", t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})

	g.Describe("Update Clan Handler", func() {
		g.It("Should update clan", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "{\"new\": \"metadata\"}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PutJSON(a, "/clans", t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(gameID, publicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbClan.GameID).Equal(gameID)
			g.Assert(dbClan.PublicID).Equal(publicID)
			g.Assert(dbClan.Name).Equal(clanName)
			g.Assert(dbClan.OwnerID).Equal(player.ID)
			g.Assert(dbClan.Metadata).Equal(metadata)
		})

		g.It("Should not update clan if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PutBody(a, "/clans", t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not update clan if player is not the owner", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PutJSON(a, "/clans", t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})

		g.It("Should not update clan if invalid data", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "it-will-fail-beacause-metada-is-not-a-json"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PutJSON(a, "/clans", t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})
}
