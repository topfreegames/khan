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
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

func TestMembershipHandler(t *testing.T) {
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Apply For Membership Handler", func() {
		g.It("Should create membership application", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":          level,
				"playerPublicID": player.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
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

		g.It("Should not create membership application if missing parameters", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "application"), t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("level is required, playerPublicID is required")
		})

		g.It("Should not create membership application if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create membership application if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":          level,
				"playerPublicID": playerPublicID,
			}

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
		})

		g.It("Should not create membership application if invalid data", func() {
			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"level":          1,
				"playerPublicID": player.PublicID,
			}

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})
	})

	g.Describe("Invite For Membership Handler", func() {
		g.It("Should create membership invitation if clan owner", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
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
			_, clan, _, players, memberships, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].Level = "CoLeader"
			memberships[0].Approved = true
			_, err = testDb.Update(memberships[0])
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": players[0].PublicID,
			}
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
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

		g.It("Should not create membership invitation if missing parameters", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "invitation"), t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("level is required, playerPublicID is required, requestorPublicID is required")
		})

		g.It("Should not create membership invitation if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create membership invitation if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
		})

		g.It("Should not create membership invitation if invalid data", func() {
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"level":             1,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})
	})

	g.Describe("Approve Or Deny Membership Invitation Handler", func() {
		g.It("Should approve membership invitation", func() {
			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID": players[0].PublicID,
			}
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
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
			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID": players[0].PublicID,
			}
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/deny"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.GameID).Equal(gameID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
			g.Assert(dbMembership.Approved).Equal(false)
			g.Assert(dbMembership.Denied).Equal(true)
		})

		g.It("Should not approve membership invitation if missing parameters", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "invitation/approve"), t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("playerPublicID is required")
		})

		g.It("Should not approve membership invitation if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not approve membership invitation if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID": playerPublicID,
			}

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})

	g.Describe("Approve Or Deny Membership Application Handler", func() {
		g.It("Should approve membership application", func() {
			_, clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Approved).Equal(true)
			g.Assert(dbMembership.Denied).Equal(false)
		})

		g.It("Should deny membership application", func() {
			_, clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/deny"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Approved).Equal(false)
			g.Assert(dbMembership.Denied).Equal(true)
		})

		g.It("Should not approve membership application if missing parameters", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "application/approve"), t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("playerPublicID is required, requestorPublicID is required")
		})

		g.It("Should not approve membership application if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not approve membership application if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})

	g.Describe("Promote Or Demote Member Handler", func() {
		g.It("Should promote member", func() {
			_, clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].Approved = true
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["level"]).Equal("Elder")

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Level).Equal("Elder")
		})

		g.It("Should demote member", func() {
			_, clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].Approved = true
			memberships[0].Level = "CoLeader"
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "demote"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()
			g.Assert(result["level"]).Equal("Elder")

			dbMembership, err := models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.Level).Equal("Elder")
		})

		g.It("Should not promote member if missing parameters", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "promote"), t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("playerPublicID is required, requestorPublicID is required")
		})

		g.It("Should not promote member if invalid payload", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not promote member if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			payload := util.JSON{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})

	g.Describe("Delete Member Handler", func() {
		g.It("Should delete member", func() {
			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			_, err = models.GetMembershipByClanAndPlayerPublicID(a.Db, gameID, clanPublicID, players[0].PublicID)
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", players[0].PublicID))
		})

		g.It("Should not delete member if missing parameters", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			res := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "delete"), t, util.JSON{})

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("playerPublicID is required, requestorPublicID is required")
		})

		g.It("Should not delete member if invalid payload", func() {
			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			res := PostBody(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), t, "invalid")

			g.Assert(res.Raw().StatusCode).Equal(http.StatusBadRequest)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not delete member if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			a := GetDefaultTestApp()
			//a.Games = LoadGames(a)
			payload := util.JSON{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusInternalServerError)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID))
		})
	})

	g.Describe("Membership Hooks", func() {
		g.It("Apply should call membership application created hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipapply",
			}, models.MembershipApplicationCreatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipapply"}, 52525)

			_, clan, _, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":          level,
				"playerPublicID": player.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, player, player)
		})

		g.It("Invite should call membership application created hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipinvite",
			}, models.MembershipApplicationCreatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipinvite"}, 52525)

			_, clan, owner, _, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, player, owner)
		})

		g.It("should call membership approved hook on application", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipapplicationapproved",
			}, models.MembershipApprovedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipapplicationapproved"}, 52525)

			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true, false)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, players[0], owner)
		})

		g.It("should call membership approved hook on invitation", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipinvitationapproved",
			}, models.MembershipApprovedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipinvitationapproved"}, 52525)

			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": players[0].PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, players[0], players[0])
		})

		g.It("should call membership denied hook on application", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipapplicationdenied",
			}, models.MembershipDeniedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipapplicationdenied"}, 52525)

			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true, false)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/deny"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, players[0], owner)
		})

		g.It("should call membership denied hook on invitation", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipinvitationdenied",
			}, models.MembershipDeniedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipinvitationdenied"}, 52525)

			_, clan, _, players, _, err := models.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := util.JSON{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": players[0].PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/deny"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, players[0], players[0])
		})

		g.It("should call membership promoted hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershippromoted",
			}, models.MembershipPromotedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershippromoted"}, 52525)

			_, clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, players[0], owner)
		})

		g.It("should call membership demoted hook", func() {
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipdemoted",
			}, models.MembershipDemotedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/membershipdemoted"}, 52525)

			_, clan, owner, players, memberships, err := models.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			g.Assert(err == nil).IsTrue()

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			g.Assert(err == nil).IsTrue()

			memberships[0].Level = "Elder"
			_, err = testDb.Update(memberships[0])
			g.Assert(err == nil).IsTrue()

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID

			payload := util.JSON{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			a := GetDefaultTestApp()
			res := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "demote"), t, payload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			a.Dispatcher.Wait()

			g.Assert(len(*responses)).Equal(1)

			apply := (*responses)[0]
			validateMembershipHookResponse(g, apply, gameID, clan, players[0], owner)
		})

	})

}
