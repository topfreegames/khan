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

func TestMembershipHandler(t *testing.T) {
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	g.Assert(err == nil).IsTrue()

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Apply For Membership Handler", func() {
		g.It("Should create membership application", func() {
			clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
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

		g.It("Should not create membership application if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create membership application if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
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

		g.It("Should not create membership application if invalid data", func() {
			clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
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
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})
	})

	g.Describe("Invite For Membership Handler", func() {
		g.It("Should create membership invitation if clan owner", func() {
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
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

		g.It("Should create membership invitation if requestor has level greater than min level", func() {
			clan, _, players, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].Level = 10
			memberships[0].Approved = true
			_, err = testDb.Update(memberships[0])
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := 1

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": players[0].PublicID,
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
			g.Assert(dbMembership.RequestorID).Equal(players[0].ID)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should not create membership invitation if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create membership invitation if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
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

		g.It("Should not create membership invitation if invalid data", func() {
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
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
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})
	})

	g.Describe("Approve Or Deny Membership Invitation Handler", func() {
		g.It("Should approve membership invitation", func() {
			clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID": players[0].PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
			g.Assert(dbMembership.Approved).Equal(true)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should deny membership invitation", func() {
			clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID": players[0].PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/deny"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
			g.Assert(dbMembership.Approved).Equal(false)
			g.Assert(dbMembership.Denied).Equal(true)
		})

		g.It("Should not approve membership invitation if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not approve membership invitation if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
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
		g.It("Should approve membership application", func() {
			clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Approved).Equal(true)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should deny membership application", func() {
			clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/deny"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Approved).Equal(false)
			g.Assert(dbMembership.Denied).Equal(true)
		})

		g.It("Should not approve membership application if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not approve membership application if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
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

	g.Describe("Promote Or Demote Member Handler", func() {
		g.It("Should promote member", func() {
			clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].Approved = true
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(int(result["level"].(float64))).Equal(memberships[0].Level + 1)

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Level).Equal(memberships[0].Level + 1)
		})

		g.It("Should demote member", func() {
			clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].Approved = true
			memberships[0].Level = 5
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "demote"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(int(result["level"].(float64))).Equal(memberships[0].Level - 1)

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Level).Equal(memberships[0].Level - 1)
		})

		g.It("Should not promote member if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not promote member if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})

	g.Describe("Delete Member Handler", func() {
		g.It("Should delete member", func() {
			clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			_, err = models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", players[0].PublicID))
		})

		g.It("Should not delete member if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not delete member if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})
}
