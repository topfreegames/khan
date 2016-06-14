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

func TestMembershipHandler(t *testing.T) {
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	g.Assert(err == nil).IsTrue()

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Apply For Membership Handler", func() {
		g.It("Should create membership", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := 1

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":          level,
				"playerPublicID": player.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.RequestorID).Equal(player.ID)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should not create membership if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not create membership if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)

			owner := models.PlayerFactory.MustCreate().(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID
			level := 1

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":          level,
				"playerPublicID": playerPublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
		})

		g.It("Should not create membership if invalid data", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":          "it-will-fail-beacause-level-is-not-an-int",
				"playerPublicID": player.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("\n[IRIS]  Error: While trying to read [JSON json: cannot unmarshal string into Go value of type int] from the request body. Trace %!!(MISSING)s(MISSING)")
		})
	})

	g.Describe("Invite For Membership Handler", func() {
		g.It("Should create membership if clan owner", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := 1

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.RequestorID).Equal(owner.ID)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should create membership if requestor has level greater than min level", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			requestor := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(requestor)
			g.Assert(err == nil).IsTrue()

			requestorMembership := models.MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      player.GameID,
				"ClanID":      clan.ID,
				"PlayerID":    requestor.ID,
				"RequestorID": owner.ID,
				"Level":       5,
				"Approved":    true,
				"Denied":      false,
			}).(*models.Membership)
			err = testDb.Insert(requestorMembership)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := 1

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": requestor.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.Level).Equal(level)
			g.Assert(dbMembership.RequestorID).Equal(requestor.ID)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should not create membership if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not create membership if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)

			owner := models.PlayerFactory.MustCreate().(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID
			level := 1

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
		})

		g.It("Should not create membership if invalid data", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":             "it-will-fail-beacause-level-is-not-an-int",
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("\n[IRIS]  Error: While trying to read [JSON json: cannot unmarshal string into Go value of type int] from the request body. Trace %!!(MISSING)s(MISSING)")
		})
	})

	g.Describe("Approve Or Deny Membership Invitation Handler", func() {
		g.It("Should approve membership", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			playerMembership := models.MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      player.GameID,
				"ClanID":      clan.ID,
				"PlayerID":    player.ID,
				"RequestorID": owner.ID,
				"Level":       0,
				"Approved":    false,
				"Denied":      false,
			}).(*models.Membership)
			err = testDb.Insert(playerMembership)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID": player.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(playerMembership.Level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.RequestorID).Equal(owner.ID)
			g.Assert(dbMembership.Approved).Equal(true)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should deny membership", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			playerMembership := models.MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      player.GameID,
				"ClanID":      clan.ID,
				"PlayerID":    player.ID,
				"RequestorID": owner.ID,
				"Level":       0,
				"Approved":    false,
				"Denied":      false,
			}).(*models.Membership)
			err = testDb.Insert(playerMembership)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID": player.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/deny"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(playerMembership.Level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.RequestorID).Equal(owner.ID)
			g.Assert(dbMembership.Approved).Equal(false)
			g.Assert(dbMembership.Denied).Equal(true)
		})

		g.It("Should not approve membership if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not approve membership if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)

			owner := models.PlayerFactory.MustCreate().(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID": playerPublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})

	g.Describe("Approve Or Deny Membership Application Handler", func() {
		g.It("Should approve membership", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			playerMembership := models.MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      player.GameID,
				"ClanID":      clan.ID,
				"PlayerID":    player.ID,
				"RequestorID": player.ID,
				"Level":       0,
				"Approved":    false,
				"Denied":      false,
			}).(*models.Membership)
			err = testDb.Insert(playerMembership)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(playerMembership.Level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.RequestorID).Equal(player.ID)
			g.Assert(dbMembership.Approved).Equal(true)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should deny membership", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			owner := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": player.GameID,
			}).(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			playerMembership := models.MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      player.GameID,
				"ClanID":      clan.ID,
				"PlayerID":    player.ID,
				"RequestorID": player.ID,
				"Level":       0,
				"Approved":    false,
				"Denied":      false,
			}).(*models.Membership)
			err = testDb.Insert(playerMembership)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/deny"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
			g.Assert(dbMembership.Level).Equal(playerMembership.Level)
			g.Assert(dbMembership.ClanID).Equal(clan.ID)
			g.Assert(dbMembership.RequestorID).Equal(player.ID)
			g.Assert(dbMembership.Approved).Equal(false)
			g.Assert(dbMembership.Denied).Equal(true)
		})

		g.It("Should not approve membership if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(
				"\n[IRIS]  Error: While trying to read [JSON invalid character 'i' looking for beginning of value] from the request body. Trace %!!(MISSING)s(MISSING)",
			)
		})

		g.It("Should not approve membership if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)

			owner := models.PlayerFactory.MustCreate().(*models.Player)
			err = testDb.Insert(owner)
			g.Assert(err == nil).IsTrue()

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  owner.GameID,
				"OwnerID": owner.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})
}
