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
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/util"
)

func TestMembershipModel(t *testing.T) {
	g := Goblin(t)
	testDb, _error := GetTestDB()
	g.Assert(_error == nil).IsTrue()

	g.Describe("Membership Model", func() {
		g.Describe("Get Number of Pending Invites", func() {
			g.It("Should get number of pending invites", func() {
				player, err := GetTestPlayerWithMemberships(testDb, uuid.NewV4().String(), 0, 0, 0, 20)
				g.Assert(err == nil).IsTrue()
				g.Assert(player != nil).IsTrue()

				totalInvites, err := GetNumberOfPendingInvites(testDb, player)
				fmt.Println(err)
				g.Assert(err == nil).IsTrue()
				g.Assert(totalInvites).Equal(20)
			})
		})
		g.Describe("Create Membership", func() {
			g.It("Should create a new Membership", func() {
				_, _, _, _, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				membership := memberships[0]

				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(membership.PlayerID)
				g.Assert(dbMembership.ClanID).Equal(membership.ClanID)
			})

			g.It("Should not create a new membership for a member with max number of invitations", func() {
				gameID := "max-invitations-game"
				player, err := GetTestPlayerWithMemberships(testDb, gameID, 0, 0, 0, 20)
				g.Assert(err == nil).IsTrue()
				g.Assert(player != nil).IsTrue()

				game, err := GetGameByPublicID(testDb, gameID)
				g.Assert(err == nil).IsTrue()
				g.Assert(game != nil).IsTrue()

				_, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, gameID, "", true)
				g.Assert(err == nil).IsTrue()
				g.Assert(clan != nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)

				g.Assert(err != nil).IsTrue()

				expected := fmt.Sprintf(
					"Player %s reached max number of pending invites",
					player.PublicID,
				)
				g.Assert(err.Error()).Equal(expected)
			})

			g.It("Should create a new membership for a member when game MaxPendingInvites is -1", func() {
				gameID := "max-invitations-game-2"
				player, err := GetTestPlayerWithMemberships(testDb, gameID, 0, 0, 0, 20)
				g.Assert(err == nil).IsTrue()
				g.Assert(player != nil).IsTrue()

				game, err := GetGameByPublicID(testDb, gameID)
				g.Assert(err == nil).IsTrue()
				g.Assert(game != nil).IsTrue()
				game.MaxPendingInvites = -1
				_, err = testDb.Update(game)
				g.Assert(err == nil).IsTrue()

				_, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, gameID, "", true)
				g.Assert(err == nil).IsTrue()
				g.Assert(clan != nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.GameID).Equal(game.PublicID)
				g.Assert(membership.PlayerID).Equal(player.ID)
				g.Assert(membership.ClanID).Equal(clan.ID)
			})
		})

		g.It("Should update a Membership", func() {
			_, _, _, _, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()
			dt := memberships[0].UpdatedAt

			time.Sleep(time.Millisecond)
			memberships[0].Approved = true
			count, err := testDb.Update(memberships[0])
			g.Assert(err == nil).IsTrue()
			g.Assert(int(count)).Equal(1)
			g.Assert(memberships[0].UpdatedAt > dt).IsTrue()
		})

		g.It("Should get existing Membership", func() {
			_, _, _, _, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(memberships[0].ID)
		})

		g.It("Should not get non-existing Membership", func() {
			_, err := GetMembershipByID(testDb, -1)
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal("Membership was not found with id: -1")
		})

		g.It("Should get an existing Membership by the player public ID", func() {
			_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, players[0].GameID, clan.PublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(memberships[0].ID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
		})

		g.Describe("Should not get Membership by the player public ID", func() {
			g.It("If non-existing Membership", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  player.GameID,
					"OwnerID": player.ID,
				}).(*Clan)
				err = testDb.Insert(clan)
				g.Assert(err == nil).IsTrue()

				dbMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, player.GameID, clan.PublicID, player.PublicID)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID))
				g.Assert(dbMembership == nil).IsTrue()
			})

			g.It("If Membership was deleted", func() {
				_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = players[0].ID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				dbMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, players[0].PublicID)
				g.Assert(err != nil).IsTrue()
				g.Assert(dbMembership == nil).IsTrue()
			})
		})

		g.It("Should get a deleted Membership by the player private ID using GetDeletedMembershipByPlayerID", func() {
			_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].DeletedAt = util.NowMilli()
			memberships[0].DeletedBy = players[0].ID
			_, err = testDb.Update(memberships[0])
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetDeletedMembershipByPlayerID(testDb, clan.GameID, players[0].ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(memberships[0].ID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
			g.Assert(dbMembership.DeletedBy).Equal(players[0].ID)
		})

		g.It("Should get a deleted Membership by the player public ID using GetMembershipByClanAndPlayerPublicID", func() {
			_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
			g.Assert(err == nil).IsTrue()

			memberships[0].DeletedAt = util.NowMilli()
			memberships[0].DeletedBy = players[0].ID
			_, err = testDb.Update(memberships[0])
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(memberships[0].ID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
			g.Assert(dbMembership.DeletedBy).Equal(players[0].ID)
		})

		g.It("Should get a denied Membership by the player public ID using GetMembershipByClanAndPlayerPublicID", func() {
			_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, players[0].PublicID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(memberships[0].ID)
			g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
		})
		g.Describe("GetOldestMemberWithHighestLevel", func() {
			g.It("Should get the member with the highest level", func() {
				_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Level = "CoLeader"
				_, err = testDb.Update(memberships[0])

				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.ID).Equal(memberships[0].ID)
				g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
			})

			g.It("Should get the oldest member", func() {
				_, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].CreatedAt -= 100000
				_, err = testDb.Update(memberships[0])

				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.ID).Equal(memberships[0].ID)
				g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
				g.Assert(dbMembership.CreatedAt < memberships[1].CreatedAt).IsTrue()
			})

			g.It("Should return an error if clan has no members", func() {
				_, clan, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, clan.GameID, clan.PublicID)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan %v has no members", clan.PublicID))
				g.Assert(dbMembership == nil).IsTrue()
			})

			g.It("Should return an error if clan does not exist", func() {
				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, "abc", "def")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Clan def has no members")
				g.Assert(dbMembership == nil).IsTrue()
			})
		})

		g.Describe("Should create a new Membership with CreateMembership", func() {
			g.It("If requestor is the player and clan.AllowApplication = true", func() {
				game, clan, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				clan.AllowApplication = true
				_, err = testDb.Update(clan)
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"Please accept me",
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
				g.Assert(dbMembership.Message).Equal("Please accept me")

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})

			g.It("If requestor is the player and clan.AllowApplication = true and previous deleted membership", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				clan.AllowApplication = true
				_, err = testDb.Update(clan)
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
				)

				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					players[0].GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
					"Please accept me",
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
				g.Assert(dbMembership.RequestorID).Equal(players[0].ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.DeletedAt).Equal(int64(0))
				g.Assert(dbMembership.Message).Equal("Please accept me")

				dbPlayer, err := GetPlayerByID(testDb, players[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})

			g.It("And approve it automatically if requestor is the player, clan.AllowApplication=true and clan.AutoJoin=true", func() {
				game, clan, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				clan.AllowApplication = true
				clan.AutoJoin = true
				_, err = testDb.Update(clan)
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"Please accept me",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(player.ID)
				g.Assert(dbMembership.RequestorID).Equal(player.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(true)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.Message).Equal("Please accept me")

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(1)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(2)
			})

			g.It("If requestor is the clan owner", func() {
				game, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(player.ID)
				g.Assert(dbMembership.RequestorID).Equal(clan.OwnerID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.Message).Equal("Please accept me")
			})

			g.It("If requestor is a member of the clan with level greater than the min level", func() {
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				memberships[0].Level = "CoLeader"
				memberships[0].Approved = true
				memberships[0].Denied = false
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					players[0].PublicID,
					"Please accept me",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(player.ID)
				g.Assert(dbMembership.RequestorID).Equal(players[0].ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.Message).Equal("Please accept me")
			})

			g.It("If deleted previous membership", func() {
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = memberships[0].PlayerID
				memberships[0].Approved = false
				memberships[0].Denied = false
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					players[0].GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
				g.Assert(dbMembership.RequestorID).Equal(owner.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.Message).Equal("Please accept me")

				dbPlayer, err := GetPlayerByID(testDb, dbMembership.PlayerID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})

			g.It("If deleted previous membership, after waiting cooldown seconds", func() {
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				game.CooldownAfterDelete = 1
				_, err = testDb.Update(game)
				g.Assert(err == nil).IsTrue()

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = memberships[0].PlayerID
				memberships[0].Approved = false
				memberships[0].Denied = false
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				time.Sleep(time.Second)
				membership, err := CreateMembership(
					testDb,
					game,
					players[0].GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
				g.Assert(dbMembership.RequestorID).Equal(owner.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.Message).Equal("Please accept me")

				dbPlayer, err := GetPlayerByID(testDb, dbMembership.PlayerID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})

			g.It("If denied previous membership, after waiting cooldown seconds", func() {
				game, clan, owner, players, _, err := GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				game.CooldownAfterDeny = 1
				_, err = testDb.Update(game)
				g.Assert(err == nil).IsTrue()

				time.Sleep(time.Second)
				membership, err := CreateMembership(
					testDb,
					game,
					players[0].GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(membership.ID != 0).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.GameID).Equal(membership.GameID)
				g.Assert(dbMembership.PlayerID).Equal(players[0].ID)
				g.Assert(dbMembership.RequestorID).Equal(owner.ID)
				g.Assert(dbMembership.ClanID).Equal(clan.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.Message).Equal("Please accept me")

				dbPlayer, err := GetPlayerByID(testDb, dbMembership.PlayerID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})
		})

		g.Describe("Should not create a new Membership with CreateMembership if", func() {
			g.It("If deleted previous membership, before waiting cooldown seconds", func() {
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				game.CooldownAfterDelete = 10
				_, err = testDb.Update(game)
				g.Assert(err == nil).IsTrue()

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = memberships[0].PlayerID
				memberships[0].Approved = false
				memberships[0].Denied = false
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					players[0].GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s must wait 10 seconds before creating a membership in clan %s.", players[0].PublicID, clan.PublicID))
			})

			g.It("If denied previous membership, before waiting cooldown seconds", func() {
				game, clan, owner, players, _, err := GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				game.CooldownAfterDeny = 10
				_, err = testDb.Update(game)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					players[0].GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s must wait 10 seconds before creating a membership in clan %s.", players[0].PublicID, clan.PublicID))
			})

			g.It("If clan reached the game's MaxMembers", func() {
				game, clan, owner, _, _, err := GetClanReachedMaxMemberships(testDb)
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan %s reached max members", clan.PublicID))
			})

			g.It("If player reached the game's MaxClansPerPlayer (member of another clans)", func() {
				game, _, owner, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				anotherClan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":   owner.GameID,
					"PublicID": uuid.NewV4().String(),
					"OwnerID":  owner.ID,
					"Metadata": map[string]interface{}{"x": "a"},
				}).(*Clan)
				err = testDb.Insert(anotherClan)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					game.PublicID,
					"Member",
					players[0].PublicID,
					anotherClan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID))
			})

			g.It("If player reached the game's MaxClansPerPlayer (owner of another clan)", func() {
				game, _, owner, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				anotherClan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":   players[0].GameID,
					"PublicID": uuid.NewV4().String(),
					"OwnerID":  players[0].ID,
					"Metadata": map[string]interface{}{"x": "a"},
				}).(*Clan)
				err = testDb.Insert(anotherClan)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					game.PublicID,
					"Member",
					players[0].PublicID,
					anotherClan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID))
			})

			g.It("If requestor is the player and clan.AllowApplication = false", func() {
				game, clan, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", player.PublicID, clan.PublicID))

			})

			g.It("Unexistent player", func() {
				game, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				playerPublicID := randomdata.FullName(randomdata.RandomGender)

				_, err = CreateMembership(
					testDb,
					game,
					owner.GameID,
					"Member",
					playerPublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
			})

			g.It("Unexistent clan", func() {
				game, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				clanPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clanPublicID,
					player.PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan was not found with id: %s", clanPublicID))

			})

			g.It("Unexistent requestor", func() {
				game, clan, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				requestorPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					requestorPublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", requestorPublicID, clan.PublicID))
			})

			g.It("Requestor's level is less than min level", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = CreateMembership(
					testDb,
					game,
					player.GameID,
					"Member",
					player.PublicID,
					clan.PublicID,
					players[0].PublicID,
					"Please accept me",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", players[0].PublicID, clan.PublicID))
			})

			g.It("Membership already exists", func() {
				game, clan, owner, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				membership, err := CreateMembership(
					testDb,
					game,
					clan.GameID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"Please accept me",
				)

				g.Assert(membership == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s already has a valid membership in clan %s.", players[0].PublicID, clan.PublicID))
			})
		})

		g.Describe("Should approve a Membership invitation with ApproveOrDenyMembershipInvitation if", func() {
			g.It("Player is not the membership requestor", func() {
				action := "approve"
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.PlayerID).Equal(players[0].ID)
				g.Assert(updatedMembership.Approved).Equal(true)
				g.Assert(updatedMembership.Denied).Equal(false)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Approved).Equal(true)
				g.Assert(dbMembership.Denied).Equal(false)

				g.Assert(dbMembership.ApproverID.Valid).IsTrue()
				g.Assert(dbMembership.ApproverID.Int64).Equal(int64(players[0].ID))

				dbPlayer, err := GetPlayerByID(testDb, players[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(1)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(2)
			})
		})

		g.Describe("Should not approve a Membership invitation with ApproveOrDenyMembershipInvitation if", func() {
			g.It("If clan reached the game's MaxMembers", func() {
				action := "approve"
				game, clan, _, players, _, err := GetClanReachedMaxMemberships(testDb)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[1].GameID,
					players[1].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan %s reached max members", clan.PublicID))
			})

			g.It("If player reached the game's MaxClansPerPlayer", func() {
				action := "approve"
				game, _, owner, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				anotherClan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":   owner.GameID,
					"PublicID": uuid.NewV4().String(),
					"OwnerID":  owner.ID,
					"Metadata": map[string]interface{}{"x": "a"},
				}).(*Clan)
				err = testDb.Insert(anotherClan)
				g.Assert(err == nil).IsTrue()

				membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      game.PublicID,
					"PlayerID":    players[0].ID,
					"ClanID":      anotherClan.ID,
					"RequestorID": owner.ID,
					"Metadata":    map[string]interface{}{"x": "a"},
					"Level":       "Member",
				}).(*Membership)
				err = testDb.Insert(membership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					anotherClan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID))
			})

			g.It("Player is the membership requestor", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = players[0].ID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[0].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Membership does not exist", func() {
				action := "approve"
				game, clan, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					player.GameID,
					player.PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID))
			})

			g.It("Membership is deleted", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = players[0].ID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", players[0].PublicID))
			})

			g.It("Membership is already approved", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})

			g.It("Membership is already denied", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Denied = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})
		})

		g.Describe("Should deny a Membership invitation with ApproveOrDenyMembershipInvitation if", func() {
			g.It("Player is not the membership requestor", func() {
				action := "deny"
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.PlayerID).Equal(players[0].ID)
				g.Assert(updatedMembership.Approved).Equal(false)
				g.Assert(updatedMembership.Denied).Equal(true)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(true)
				g.Assert(dbMembership.DenierID.Valid).IsTrue()
				g.Assert(dbMembership.DenierID.Int64).Equal(int64(players[0].ID))
				g.Assert(dbMembership.DeniedAt > util.NowMilli()-1000).IsTrue()
			})
		})

		g.Describe("Should not ApproveOrDenyMembershipInvitation if", func() {
			g.It("Invalid action", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipInvitation(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					"invalid-action",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("invalid-action a membership is not a valid action.")
			})
		})

		g.Describe("Should approve a Membership application with ApproveOrDenyMembershipApplication if", func() {
			g.It("Owner", func() {
				action := "approve"
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])

				updatedMembership, err := ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.Approved).Equal(true)
				g.Assert(updatedMembership.Denied).Equal(false)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.Approved).Equal(true)
				g.Assert(dbMembership.Denied).Equal(false)

				g.Assert(dbMembership.ApproverID.Valid).IsTrue()
				g.Assert(dbMembership.ApproverID.Int64).Equal(int64(owner.ID))

				dbPlayer, err := GetPlayerByID(testDb, players[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(1)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(2)
			})

			g.It("Requestor is member of the clan with level > minLevel", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "CoLeader"
				memberships[1].Approved = true
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(memberships[0].ID)
				g.Assert(updatedMembership.Approved).Equal(true)
				g.Assert(updatedMembership.Denied).Equal(false)

				g.Assert(updatedMembership.ApproverID.Valid).IsTrue()
				g.Assert(updatedMembership.ApproverID.Int64).Equal(int64(players[1].ID))

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Approved).Equal(true)
				g.Assert(dbMembership.Denied).Equal(false)

				dbPlayer, err := GetPlayerByID(testDb, players[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(1)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(2)
			})
		})

		g.Describe("Should not approve a Membership application with ApproveOrDenyMembershipApplication if", func() {
			g.It("If clan reached the game's MaxMembers", func() {
				action := "approve"
				game, clan, owner, players, memberships, err := GetClanReachedMaxMemberships(testDb)
				g.Assert(err == nil).IsTrue()

				memberships[1].RequestorID = memberships[1].PlayerID
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[1].GameID,
					players[1].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan %s reached max members", clan.PublicID))
			})

			g.It("If player reached the game's MaxClansPerPlayer", func() {
				action := "approve"
				game, _, owner, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				anotherClan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":   owner.GameID,
					"PublicID": uuid.NewV4().String(),
					"OwnerID":  owner.ID,
					"Metadata": map[string]interface{}{"x": "a"},
				}).(*Clan)
				err = testDb.Insert(anotherClan)
				g.Assert(err == nil).IsTrue()

				membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      game.PublicID,
					"PlayerID":    players[0].ID,
					"ClanID":      anotherClan.ID,
					"RequestorID": players[0].ID,
					"Metadata":    map[string]interface{}{"x": "a"},
					"Level":       "Member",
				}).(*Membership)
				err = testDb.Insert(membership)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					anotherClan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID))
			})

			g.It("Requestor is member of the clan with level < minLevel", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "Member"
				memberships[1].Approved = true
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Requestor is not approved member of the clan", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "CoLeader"
				memberships[1].Approved = false
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Requestor is not member of the clan", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					requestor.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Requestor membership is deleted", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				memberships[1].DeletedAt = util.NowMilli()
				memberships[1].DeletedBy = players[1].ID
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Requestor is the player of the membership", func() {
				action := "approve"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[0].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Player was not the membership requestor", func() {
				action := "approve"
				game, clan, owner, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", owner.PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Membership does not exist", func() {
				action := "approve"
				game, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
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
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action))
			})

			g.It("Membership is already denied", func() {
				action := "approve"
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				memberships[0].Denied = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
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
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.Approved).Equal(false)
				g.Assert(updatedMembership.Denied).Equal(true)

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(true)
				g.Assert(dbMembership.DenierID.Valid).IsTrue()
				g.Assert(dbMembership.DenierID.Int64).Equal(int64(owner.ID))
				g.Assert(dbMembership.DeniedAt > util.NowMilli()-1000).IsTrue()

			})
		})

		g.Describe("Should not ApproveOrDenyMembershipApplication if", func() {
			g.It("Invalid action", func() {
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].RequestorID = memberships[0].PlayerID
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = ApproveOrDenyMembershipApplication(
					testDb,
					game,
					players[0].GameID,
					players[0].PublicID,
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
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(memberships[0].ID)
				g.Assert(updatedMembership.Level).Equal("Elder")

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Level).Equal("Elder")
			})

			g.It("If requestor has enough level", func() {
				action := "promote"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "CoLeader"
				memberships[1].Approved = true
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				updatedMembership, err := PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(memberships[0].ID)
				g.Assert(updatedMembership.Level).Equal("Elder")

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Level).Equal("Elder")
			})
		})

		g.Describe("Should not promote a member with PromoteOrDemoteMember", func() {
			g.It("If requestor is the player", func() {
				action := "promote"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[0].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("If requestor does not have enough level", func() {
				action := "promote"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "Member"
				memberships[1].Approved = true
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("If requestor is not a clan member", func() {
				action := "promote"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				requestorPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					requestorPublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestorPublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Requestor membership is deleted", func() {
				action := "promote"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])

				memberships[1].DeletedAt = util.NowMilli()
				memberships[1].DeletedBy = clan.OwnerID
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("If player is not a clan member", func() {
				action := "promote"
				game, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID))
			})

			g.It("If player membership is not approved", func() {
				action := "promote"
				game, clan, owner, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that is denied or not yet approved", action))
			})

			g.It("If player membership is denied", func() {
				action := "promote"
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Denied = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s membership that is denied or not yet approved", action))
			})

			g.It("If requestor membership is not approved", func() {
				action := "promote"
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID))
			})

			g.It("Player is already max level", func() {
				action := "promote"
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				memberships[0].Level = "CoLeader"
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s member that is already level %d", action, 3))
			})
		})

		g.Describe("Should demote a member with PromoteOrDemoteMember", func() {
			g.It("If requestor is the owner", func() {
				action := "demote"
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Level = "CoLeader"
				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])

				updatedMembership, err := PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updatedMembership.ID).Equal(memberships[0].ID)
				g.Assert(updatedMembership.Level).Equal("Elder")

				dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.Level).Equal("Elder")
			})
		})

		g.Describe("Should not demote a member with PromoteOrDemoteMember if", func() {
			g.It("Player is already min level", func() {
				action := "demote"
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				memberships[0].Level = "Member"
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					action,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Cannot %s member that is already level %d", action, 1))
			})
		})

		g.Describe("Should not PromoteOrDemoteMember", func() {
			g.It("Invalid action", func() {
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[0].Approved = true
				_, err = testDb.Update(memberships[0])
				g.Assert(err == nil).IsTrue()

				_, err = PromoteOrDemoteMember(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"invalid-action",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("invalid-action a membership is not a valid action.")
			})
		})

		g.Describe("Should delete a membership with DeleteMembership", func() {
			g.It("If requestor is the owner", func() {
				game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
				)

				g.Assert(err == nil).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.DeletedBy).Equal(owner.ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.DeletedAt > util.NowMilli()-1000).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, memberships[0].PlayerID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})

			g.It("If requestor is the player", func() {
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
				)

				g.Assert(err == nil).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.DeletedBy).Equal(players[0].ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.DeletedAt > util.NowMilli()-1000).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, memberships[0].PlayerID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(1)
			})

			g.It("If requestor has enough level and offset", func() {
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 2, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "CoLeader"
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
				)
				g.Assert(err == nil).IsTrue()

				dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbMembership.DeletedBy).Equal(players[1].ID)
				g.Assert(dbMembership.Approved).Equal(false)
				g.Assert(dbMembership.Denied).Equal(false)
				g.Assert(dbMembership.DeletedAt > util.NowMilli()-1000).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, memberships[0].PlayerID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(0)

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.MembershipCount).Equal(2)
			})
		})

		g.Describe("Should not delete a membership with DeleteMembership", func() {
			g.It("If requestor does not have enough level", func() {
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "Member"
				memberships[1].Approved = true
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
				)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID))
			})

			g.It("If requestor has enough level but not enough offset", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
				)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID))
			})

			g.It("If requestor is not a clan member", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				g.Assert(err == nil).IsTrue()

				requestor := PlayerFactory.MustCreateWithOption(map[string]interface{}{
					"GameID": clan.GameID,
				}).(*Player)
				err = testDb.Insert(requestor)
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					requestor.PublicID,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, "delete", players[0].PublicID, clan.PublicID))
			})

			g.It("If requestor membership is denied", func() {
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "CoLeader"
				memberships[1].Denied = true
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
				)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID))
			})

			g.It("If requestor membership is not approved", func() {
				game, clan, _, players, memberships, err := GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				g.Assert(err == nil).IsTrue()

				memberships[1].Level = "CoLeader"
				memberships[1].Approved = false
				_, err = testDb.Update(memberships[1])
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					clan.GameID,
					players[0].PublicID,
					clan.PublicID,
					players[1].PublicID,
				)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID))
			})
		})
	})
}
