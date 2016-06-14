// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
)

func TestMembershipModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Membership Model", func() {
		g.It("Should create a new Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := &Membership{
				GameID:      "test",
				ClanID:      clan.ID,
				PlayerID:    player.ID,
				RequestorID: player.ID,
				Level:       1,
				Approved:    false,
				Denied:      false,
			}
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()
			g.Assert(membership.ID != 0).IsTrue()

			dbMembership, err := GetMembershipByID(testDb, membership.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbMembership.GameID).Equal(membership.GameID)
			g.Assert(dbMembership.PlayerID).Equal(membership.PlayerID)
			g.Assert(dbMembership.ClanID).Equal(membership.ClanID)
		})

		g.It("Should update a Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"PlayerID":    player.ID,
				"ClanID":      clan.ID,
				"RequestorID": clan.OwnerID,
			}).(*Membership)
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()
			dt := membership.UpdatedAt

			membership.Approved = true
			count, err := testDb.Update(membership)
			g.Assert(err == nil).IsTrue()
			g.Assert(int(count)).Equal(1)
			g.Assert(membership.UpdatedAt > dt).IsTrue()
		})

		g.It("Should get existing Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"PlayerID":    player.ID,
				"ClanID":      clan.ID,
				"RequestorID": player.ID,
			}).(*Membership)
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByID(testDb, membership.ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(membership.ID)
		})

		g.It("Should not get non-existing Membership", func() {
			_, err = GetMembershipByID(testDb, -1)
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal("Membership was not found with id: -1")
		})

		g.It("Should get an existing Membership by the player public ID", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      player.GameID,
				"PlayerID":    player.ID,
				"ClanID":      clan.ID,
				"RequestorID": player.ID,
			}).(*Membership)
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, player.GameID, clan.PublicID, player.PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(membership.ID)
			g.Assert(dbMembership.PlayerID).Equal(player.ID)
		})

		g.It("Should not get non-existing Membership by the player public ID", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, player.GameID, clan.PublicID, player.PublicID)
			g.Assert(err != nil).IsTrue()
			g.Assert(dbMembership == nil).IsTrue()
		})

		g.Describe("Should create a new Membership with CreateMembership", func() {
			g.It("If requestor is the player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clan.PublicID,
					player.PublicID,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(player.ID)
				g.Assert(dbMembership.RequestorID).Equal(player.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
			})

			g.It("If requestor is the clan owner", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(player.ID)
				g.Assert(dbMembership.RequestorID).Equal(owner.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
			})

			g.It("If requestor is a member of the clan with level greater than the min level", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": owner.ID,
					"Level":       5,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(player.ID)
				g.Assert(dbMembership.RequestorID).Equal(requestor.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
			})
		})

		g.Describe("Should not create a new Membership with CreateMembership if", func() {
			g.It("Unexistent player", func() {
				owner := PlayerFactory.MustCreate().(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateMembership(
					testDb,
					owner.GameID,
					1,
					playerPublicID,
					clan.PublicID,
					owner.PublicID,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
			})

			g.It("Unexistent clan", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				clanPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clanPublicID,
					player.PublicID,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan was not found with id: %s", clanPublicID))

			})

			g.It("Unexistent requestor", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestorPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clan.PublicID,
					requestorPublicID,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", requestorPublicID, clan.PublicID))
			})

			g.It("Requestor's level is less than min level", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", requestor.PublicID, clan.PublicID))
			})

			g.It("Membership already exists", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				previousMembership := &Membership{
					GameID:      player.GameID,
					ClanID:      clan.ID,
					PlayerID:    player.ID,
					RequestorID: player.ID,
					Level:       0,
					Approved:    true,
					Denied:      false,
				}
				err = testDb.Insert(previousMembership)

				membership, err := CreateMembership(
					testDb,
					player.GameID,
					1,
					player.PublicID,
					clan.PublicID,
					player.PublicID,
				)

				g.Assert(membership == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("pq: duplicate key value violates unique constraint \"playerid_clanid\"")
			})
		})

		g.Describe("Should approve a Membership invitation with AcceptOrDenyMembershipInvitation if", func() {
			g.It("Player is not the membership requestor", func() {
				action := "approve"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
				g.Assert(updatedMembership.Approved).Equal(true)
				g.Assert(updatedMembership.Denied).Equal(false)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(playerMembership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(playerMembership.PlayerID)
				g.Assert(dbMembership.RequestorID).Equal(playerMembership.RequestorID)
				g.Assert(dbMembership.Level).Equal(playerMembership.Level)
				g.Assert(dbMembership.Approved).Equal(true)
				g.Assert(dbMembership.Denied).Equal(false)
			})
		})

		g.Describe("Should not accept a Membership invitation with AcceptOrDenyMembershipInvitation if", func() {
			g.It("Player is the membership requestor", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", player.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("Membership does not exist", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID))
			})

			g.It("Membership is already approved", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})

			g.It("Membership is already denied", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      true,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})
		})

		g.Describe("Should deny a Membership invitation with AcceptOrDenyMembershipInvitation if", func() {
			g.It("Player is not the membership requestor", func() {
				action := "deny"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
				g.Assert(updatedMembership.Approved).Equal(false)
				g.Assert(updatedMembership.Denied).Equal(true)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(playerMembership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(playerMembership.PlayerID)
				g.Assert(dbMembership.RequestorID).Equal(playerMembership.RequestorID)
				g.Assert(dbMembership.Level).Equal(playerMembership.Level)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(true)
			})
		})

		g.Describe("Should not AcceptOrDenyMembershipInvitation if", func() {
			g.It("Invalid action", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					"invalid-action",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("invalid-action a membership is not a valid action.")
			})

			g.Describe("Should approve a Membership application with ApproveOrDenyMembershipApplication if", func() {
				g.It("Owner", func() {
					action := "approve"

					player := PlayerFactory.MustCreate().(*Player)
					err := testDb.Insert(player)
					g.Assert(err == nil).IsTrue()

					owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": player.GameID,
					}).(*Player)
					err = testDb.Insert(owner)
					g.Assert(err == nil).IsTrue()

					clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":  owner.GameID,
						"OwnerID": owner.ID,
					}).(*Clan)
					err = testDb.Insert(clan)
					g.Assert(err == nil).IsTrue()

					playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":      player.GameID,
						"ClanID":      clan.ID,
						"PlayerID":    player.ID,
						"RequestorID": player.ID,
						"Level":       0,
						"Approved":    false,
						"Denied":      false,
					}).(*Membership)
					err = testDb.Insert(playerMembership)
					g.Assert(err == nil).IsTrue()

					updatedMembership, err := ApproveOrDenyMembershipApplication(
						testDb,
						player.GameID,
						player.PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					g.Assert(err == nil).IsTrue()
					g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
					g.Assert(updatedMembership.Approved).Equal(true)
					g.Assert(updatedMembership.Denied).Equal(false)

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					g.Assert(err == nil).IsTrue()

					g.Assert(dbMembership.GameID).Equal(playerMembership.GameID)
					g.Assert(dbMembership.PlayerID).Equal(playerMembership.PlayerID)
					g.Assert(dbMembership.RequestorID).Equal(playerMembership.RequestorID)
					g.Assert(dbMembership.Level).Equal(playerMembership.Level)
					g.Assert(dbMembership.Approved).Equal(true)
					g.Assert(dbMembership.Denied).Equal(false)
				})

				g.It("Requestor is member of the clan with level > minLevel", func() {
					action := "approve"

					player := PlayerFactory.MustCreate().(*Player)
					err := testDb.Insert(player)
					g.Assert(err == nil).IsTrue()

					owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": player.GameID,
					}).(*Player)
					err = testDb.Insert(owner)
					g.Assert(err == nil).IsTrue()

					clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":  owner.GameID,
						"OwnerID": owner.ID,
					}).(*Clan)
					err = testDb.Insert(clan)
					g.Assert(err == nil).IsTrue()

					requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": player.GameID,
					}).(*Player)
					err = testDb.Insert(requestor)
					g.Assert(err == nil).IsTrue()

					requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":      player.GameID,
						"ClanID":      clan.ID,
						"PlayerID":    requestor.ID,
						"RequestorID": requestor.ID,
						"Level":       5,
						"Approved":    true,
						"Denied":      false,
					}).(*Membership)
					err = testDb.Insert(requestorMembership)
					g.Assert(err == nil).IsTrue()

					playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":      player.GameID,
						"ClanID":      clan.ID,
						"PlayerID":    player.ID,
						"RequestorID": player.ID,
						"Level":       0,
						"Approved":    false,
						"Denied":      false,
					}).(*Membership)
					err = testDb.Insert(playerMembership)
					g.Assert(err == nil).IsTrue()

					updatedMembership, err := ApproveOrDenyMembershipApplication(
						testDb,
						player.GameID,
						player.PublicID,
						clan.PublicID,
						requestor.PublicID,
						action,
					)

					g.Assert(err == nil).IsTrue()
					g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
					g.Assert(updatedMembership.Approved).Equal(true)
					g.Assert(updatedMembership.Denied).Equal(false)

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					g.Assert(err == nil).IsTrue()

					g.Assert(dbMembership.GameID).Equal(playerMembership.GameID)
					g.Assert(dbMembership.PlayerID).Equal(playerMembership.PlayerID)
					g.Assert(dbMembership.RequestorID).Equal(playerMembership.RequestorID)
					g.Assert(dbMembership.Level).Equal(playerMembership.Level)
					g.Assert(dbMembership.Approved).Equal(true)
					g.Assert(dbMembership.Denied).Equal(false)
				})
			})
		})

		g.Describe("Should not accept a Membership application with ApproveOrDenyMembershipApplication if", func() {
			g.It("Requestor is member of the clan with level < minLevel", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": requestor.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("Requestor is not approved member of the clan", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": requestor.ID,
					"Level":       5,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("Requestor is not member of the clan", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("Requestor is the player of the membership", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", player.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("Player was not the membership requestor", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", owner.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("Membership does not exist", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID))
			})

			g.It("Membership is already approved", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})

			g.It("Membership is already denied", func() {
				action := "approve"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})
		})

		g.Describe("Should deny a Membership application with ApproveOrDenyMembershipApplication if", func() {
			g.It("Owner", func() {
				action := "deny"

				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
				g.Assert(updatedMembership.Approved).Equal(false)
				g.Assert(updatedMembership.Denied).Equal(true)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(playerMembership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(playerMembership.PlayerID)
				g.Assert(dbMembership.RequestorID).Equal(playerMembership.RequestorID)
				g.Assert(dbMembership.Level).Equal(playerMembership.Level)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(true)
			})
		})

		g.Describe("Should not AcceptOrDenyMembershipInvitation if", func() {
			g.It("Invalid action", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": player.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"invalid-action",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("invalid-action a membership is not a valid action.")
			})
		})

		g.Describe("Should promote a member with PromoteOrDemoteMember", func() {
			g.It("If requestor is the owner", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
				g.Assert(updatedMembership.Level).Equal(playerMembership.Level + 1)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Level).Equal(playerMembership.Level + 1)
			})

			g.It("If requestor has enough level", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      requestor.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": owner.ID,
					"Level":       10,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
				g.Assert(updatedMembership.Level).Equal(playerMembership.Level + 1)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Level).Equal(playerMembership.Level + 1)
			})
		})

		g.Describe("Should not promote a member with PromoteOrDemoteMember", func() {
			g.It("If requestor is the player", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s member %s in clan %s", player.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("If requestor does not have enough level", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      requestor.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": owner.ID,
					"Level":       1,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s member %s in clan %s", requestor.PublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("If requestor is not a clan member", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				requestorPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestorPublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s member %s in clan %s", requestorPublicID, action, player.PublicID, clan.PublicID))
			})

			g.It("If player is not a clan member", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestorPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestorPublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID))
			})

			g.It("If player membership is not approved", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that is denied or not yet approved", action))
			})

			g.It("If player membership is denied", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    false,
					"Denied":      true,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that is denied or not yet approved", action))
			})

			g.It("If requestor membership is not approved", func() {
				action := "promote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				requestorMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      requestor.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    requestor.ID,
					"RequestorID": owner.ID,
					"Level":       1,
					"Approved":    false,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(requestorMembership)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s member %s in clan %s", requestor.PublicID, action, player.PublicID, clan.PublicID))
			})
		})

		g.Describe("Should demote a member with PromoteOrDemoteMember", func() {
			g.It("If requestor is the owner", func() {
				action := "demote"
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(playerMembership.ID)
				g.Assert(updatedMembership.Level).Equal(playerMembership.Level - 1)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Level).Equal(playerMembership.Level - 1)
			})
		})

		g.Describe("Should not PromoteOrDemoteMember if invalid action", func() {
			g.It("If requestor is the owner", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": player.GameID,
				}).(*Player)
				err = testDb.Insert(owner)
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  owner.GameID,
					"OwnerID": owner.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				playerMembership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"ClanID":      clan.ID,
					"PlayerID":    player.ID,
					"RequestorID": owner.ID,
					"Level":       0,
					"Approved":    true,
					"Denied":      false,
				}).(*Membership)
				err = testDb.Insert(playerMembership)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"invalid-action",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("invalid-action a membership is not a valid action.")
			})
		})
	})
}
