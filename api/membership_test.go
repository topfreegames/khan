// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var _ = Describe("Membership API Handler", func() {
	var testDb, db models.DB
	var a *api.App

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())

		a = GetDefaultTestApp()
		db = a.Db(nil)
		fixtures.ConfigureAndStartGoWorkers()
	})

	Describe("Apply For Membership Handler", func() {
		It("Should create membership application", func() {
			_, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			clan.AutoJoin = true
			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":          level,
				"playerPublicID": player.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["approved"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(player.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.ClanID).To(Equal(clan.ID))
			Expect(dbMembership.RequestorID).To(Equal(player.ID))
			Expect(dbMembership.Denied).To(Equal(false))
		})

		It("Should create membership application sending a message", func() {
			_, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			clan.AutoJoin = false
			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":          level,
				"playerPublicID": player.PublicID,
				"message":        "Please accept me, I am nice",
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["approved"]).To(BeFalse())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(player.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.ClanID).To(Equal(clan.ID))
			Expect(dbMembership.RequestorID).To(Equal(player.ID))
			Expect(dbMembership.Denied).To(Equal(false))
			Expect(dbMembership.Message).To(Equal(payload["message"]))
		})

		It("Should conflict when player is already a member", func() {
			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.Approved).To(Equal(true))
			Expect(dbMembership.Denied).To(Equal(false))

			payload = map[string]interface{}{
				"level":          "Member",
				"playerPublicID": players[0].PublicID,
				"message":        "Please accept me, I am nice",
			}
			status, body = PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusConflict))
		})

		It("Should create membership application sending a message after player left clan", func() {
			_, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			clanPublicID := clan.PublicID
			player := players[0]
			var result map[string]interface{}

			// Leave clan
			payload := map[string]interface{}{
				"playerPublicID":    player.PublicID,
				"requestorPublicID": player.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), payload)

			Expect(status).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			// Left clan, trying to apply again

			// Apply to clan
			level := "Member"
			payload = map[string]interface{}{
				"level":          level,
				"playerPublicID": player.PublicID,
				"message":        "Please accept me, I am nice",
			}
			status, body = PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(player.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.ClanID).To(Equal(clan.ID))
			Expect(dbMembership.RequestorID).To(Equal(player.ID))
			Expect(dbMembership.Denied).To(Equal(false))
			Expect(dbMembership.Message).To(Equal(payload["message"]))
		})

		It("Should delete member", func() {
			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", players[0].PublicID)))
		})

		It("Should not create membership application if missing parameters", func() {
			status, body := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "application"), map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("level is required, playerPublicID is required"))
		})

		It("Should not create membership application if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			status, body := Post(a, CreateMembershipRoute(gameID, clanPublicID, "application"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not create membership application if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":          level,
				"playerPublicID": playerPublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID)))
		})

		It("Should not create membership application if invalid data", func() {
			_, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"level":          1,
				"playerPublicID": player.PublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("expected string near offset"))
		})
	})

	Describe("Invite For Membership Handler", func() {
		It("Should create membership invitation if clan owner", func() {
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(player.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.ClanID).To(Equal(clan.ID))
			Expect(dbMembership.RequestorID).To(Equal(owner.ID))
			Expect(dbMembership.Denied).To(Equal(false))
		})

		It("Should create membership invitation sending a message", func() {
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
				"message":           "Please accept me, I am nice",
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(player.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.ClanID).To(Equal(clan.ID))
			Expect(dbMembership.RequestorID).To(Equal(owner.ID))
			Expect(dbMembership.Denied).To(Equal(false))
			Expect(dbMembership.Message).To(Equal(payload["message"]))
		})

		It("Should create membership invitation if requestor has level greater than min level", func() {
			_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].Level = "CoLeader"
			memberships[0].Approved = true
			_, err = testDb.Update(memberships[0])
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": players[0].PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, player.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(player.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.ClanID).To(Equal(clan.ID))
			Expect(dbMembership.Level).To(Equal(level))
			Expect(dbMembership.RequestorID).To(Equal(players[0].ID))
			Expect(dbMembership.Denied).To(Equal(false))
		})

		It("Should not create membership invitation if missing parameters", func() {
			status, body := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "invitation"), map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("level is required, playerPublicID is required, requestorPublicID is required"))
		})

		It("Should not create membership invitation if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			status, body := Post(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not create membership invitation if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := owner.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID)))
		})

		It("Should not create membership invitation if invalid data", func() {
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"level":             1,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), payload)

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("expected string near offset"))
		})
	})

	Describe("Approve Or Deny Membership Invitation Handler", func() {
		It("Should approve membership invitation", func() {
			_, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID": players[0].PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
			Expect(dbMembership.Approved).To(Equal(true))
			Expect(dbMembership.Denied).To(Equal(false))
		})

		It("Should deny membership invitation", func() {
			_, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID": players[0].PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/deny"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.GameID).To(Equal(gameID))
			Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
			Expect(dbMembership.Approved).To(Equal(false))
			Expect(dbMembership.Denied).To(Equal(true))
		})

		It("Should not approve membership invitation if missing parameters", func() {
			status, body := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "invitation/approve"), map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("playerPublicID is required"))
		})

		It("Should not approve membership invitation if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			status, body := Post(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not approve membership invitation if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID": playerPublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID)))
		})
	})

	Describe("Approve Or Deny Membership Application Handler", func() {
		It("Should approve membership application", func() {
			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.Approved).To(Equal(true))
			Expect(dbMembership.Denied).To(Equal(false))
		})

		It("Should conflict when player is already a member", func() {
			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.Approved).To(Equal(true))
			Expect(dbMembership.Denied).To(Equal(false))

			status, body = PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), payload)
			Expect(status).To(Equal(http.StatusConflict))
		})

		It("Should deny membership application", func() {
			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].RequestorID = memberships[0].PlayerID
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/deny"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.Approved).To(Equal(false))
			Expect(dbMembership.Denied).To(Equal(true))
		})

		It("Should not approve membership application if missing parameters", func() {
			status, body := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "application/approve"), map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("playerPublicID is required, requestorPublicID is required"))
		})

		It("Should not approve membership application if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			status, body := Post(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not approve membership application if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID)))
		})
	})

	Describe("Promote Or Demote Member Handler", func() {
		It("Should promote member", func() {
			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].Approved = true
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["level"]).To(Equal("Elder"))

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.Level).To(Equal("Elder"))
		})

		It("Should demote member", func() {
			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			memberships[0].Approved = true
			memberships[0].Level = "CoLeader"
			_, err = testDb.Update(memberships[0])

			gameID := players[0].GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "demote"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["level"]).To(Equal("Elder"))

			dbMembership, err := models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbMembership.Level).To(Equal("Elder"))
		})

		It("Should not promote member if missing parameters", func() {
			status, body := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "promote"), map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("playerPublicID is required, requestorPublicID is required"))
		})

		It("Should not promote member if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			status, body := Post(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not promote member if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID)))
		})
	})

	Describe("Delete Member Handler", func() {
		It("Should delete member", func() {
			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := clan.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = models.GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, players[0].PublicID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", players[0].PublicID)))
		})

		It("Should not delete member if missing parameters", func() {
			status, body := PostJSON(a, CreateMembershipRoute("gameID", "clanPublicID", "delete"), map[string]interface{}{})

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("playerPublicID is required, requestorPublicID is required"))
		})

		It("Should not delete member if invalid payload", func() {
			gameID := "gameID"
			clanPublicID := randomdata.FullName(randomdata.RandomGender)

			status, body := Post(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), "invalid")

			Expect(status).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring(InvalidJSONError))
		})

		It("Should not delete member if player does not exist", func() {
			playerPublicID := randomdata.FullName(randomdata.RandomGender)
			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
			Expect(err).NotTo(HaveOccurred())

			gameID := owner.GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    playerPublicID,
				"requestorPublicID": owner.PublicID,
			}

			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), payload)

			Expect(status).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Membership was not found with id: %s", playerPublicID)))
		})
	})

	Describe("Membership Hooks", func() {
		It("Apply should call membership application created hook with non empty message", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipapply",
			}, models.MembershipApplicationCreatedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipapply"}, 52525)

			_, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":          level,
				"playerPublicID": player.PublicID,
				"message":        "Please accept me, I am nice",
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, player, player)
			Expect(response["message"]).To(Equal(payload["message"]))
		})

		It("Invite should call membership application created hook", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipinvite",
			}, models.MembershipApplicationCreatedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipinvite"}, 52525)

			_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID": clan.GameID,
			}).(*models.Player)
			err = testDb.Insert(player)
			Expect(err).NotTo(HaveOccurred())

			gameID := player.GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    player.PublicID,
				"requestorPublicID": owner.PublicID,
				"message":           "Join my clan",
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, player, owner)
			Expect(response["message"]).To(Equal(payload["message"]))
		})

		It("should call membership approved hook on application", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipapplicationapproved",
			}, models.MembershipApprovedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipapplicationapproved"}, 52525)

			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true, false)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/approve"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], owner)
			validateApproveDenyMembershipHookResponse(response, players[0])
		})

		It("should call membership approved hook on invitation", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipinvitationapproved",
			}, models.MembershipApprovedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipinvitationapproved"}, 52525)

			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": players[0].PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/approve"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], players[0])
			validateApproveDenyMembershipHookResponse(response, owner)
		})

		It("should call membership denied hook on application", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipapplicationdenied",
			}, models.MembershipDeniedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipapplicationdenied"}, 52525)

			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true, false)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "application/deny"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], owner)
			validateApproveDenyMembershipHookResponse(response, players[0])
		})

		It("should call membership denied hook on invitation", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipinvitationdenied",
			}, models.MembershipDeniedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipinvitationdenied"}, 52525)

			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID
			level := "Member"

			payload := map[string]interface{}{
				"level":             level,
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": players[0].PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "invitation/deny"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], players[0])
			validateApproveDenyMembershipHookResponse(response, owner)
		})

		It("should call membership promoted hook", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershippromoted",
			}, models.MembershipPromotedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershippromoted"}, 52525)

			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "promote"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], owner)
		})

		It("should call membership demoted hook", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipdemoted",
			}, models.MembershipDemotedHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipdemoted"}, 52525)

			_, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			memberships[0].Level = "Elder"
			_, err = testDb.Update(memberships[0])
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "demote"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], owner)
		})

		It("should call membership deleted hook", func() {
			hooks, err := fixtures.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/membershipdeleted",
			}, models.MembershipLeftHook)
			Expect(err).NotTo(HaveOccurred())
			responses := startRouteHandler([]string{"/membershipdeleted"}, 52525)

			_, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, hooks[0].GameID, "", true)
			Expect(err).NotTo(HaveOccurred())

			clan.AllowApplication = true
			_, err = testDb.Update(clan)
			Expect(err).NotTo(HaveOccurred())

			gameID := hooks[0].GameID
			clanPublicID := clan.PublicID

			payload := map[string]interface{}{
				"playerPublicID":    players[0].PublicID,
				"requestorPublicID": owner.PublicID,
			}
			status, body := PostJSON(a, CreateMembershipRoute(gameID, clanPublicID, "delete"), payload)

			Expect(status).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			Eventually(func() int {
				return len(*responses)
			}).Should(Equal(1))

			response := (*responses)[0]["payload"].(map[string]interface{})
			validateMembershipHookResponse(response, gameID, clan, players[0], owner)
		})
	})
})
