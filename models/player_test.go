// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	"github.com/satori/go.uuid"
)

func TestPlayerModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()

	g.Assert(err == nil).IsTrue()

	g.Describe("Player Model", func() {

		g.Describe("Model Basic Tests", func() {
			g.It("Should create a new Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()
				g.Assert(player.ID != 0).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.GameID).Equal(player.GameID)
				g.Assert(dbPlayer.PublicID).Equal(player.PublicID)
			})

			g.It("Should update a new Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()
				dt := player.UpdatedAt

				time.Sleep(time.Millisecond)

				player.Metadata = map[string]interface{}{"x": 1}
				count, err := testDb.Update(player)
				g.Assert(err == nil).IsTrue()
				g.Assert(int(count)).Equal(1)
				g.Assert(player.UpdatedAt > dt).IsTrue()
			})
		})

		g.Describe("Get Player By ID", func() {
			g.It("Should get existing Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.ID).Equal(player.ID)
			})

			g.It("Should not get non-existing Player", func() {
				_, err := GetPlayerByID(testDb, -1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: -1")
			})
		})

		g.Describe("Get Player By Public ID", func() {
			g.It("Should get existing Player by Game and Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				dbPlayer, err := GetPlayerByPublicID(testDb, player.GameID, player.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.ID).Equal(player.ID)
			})

			g.It("Should not get non-existing Player by Game and Player", func() {
				_, err := GetPlayerByPublicID(testDb, "invalid-game", "invalid-player")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: invalid-player")
			})
		})

		g.Describe("Create Player", func() {
			g.It("Should create a new Player with CreatePlayer", func() {
				player, err := CreatePlayer(
					testDb,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					"player-name",
					map[string]interface{}{},
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(player.ID != 0).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.GameID).Equal(player.GameID)
				g.Assert(dbPlayer.PublicID).Equal(player.PublicID)
			})
		})

		g.Describe("Update Player", func() {
			g.It("Should update a Player with UpdatePlayer", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				metadata := map[string]interface{}{"x": 1}
				updPlayer, err := UpdatePlayer(
					testDb,
					player.GameID,
					player.PublicID,
					player.Name,
					metadata,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updPlayer.ID).Equal(player.ID)

				dbPlayer, err := GetPlayerByPublicID(testDb, player.GameID, player.PublicID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.Metadata).Equal(metadata)
			})

			g.It("Should create Player with UpdatePlayer if player does not exist", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()

				gameID := game.PublicID
				publicID := uuid.NewV4().String()

				metadata := map[string]interface{}{"x": "1"}
				updPlayer, err := UpdatePlayer(
					testDb,
					gameID,
					publicID,
					publicID,
					metadata,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updPlayer.ID > 0).IsTrue()

				dbPlayer, err := GetPlayerByPublicID(testDb, updPlayer.GameID, updPlayer.PublicID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.Metadata).Equal(metadata)
			})

			g.It("Should not update a Player with Invalid Data with UpdatePlayer", func() {
				_, err := UpdatePlayer(
					testDb,
					"-1",
					"qwe",
					"some player name",
					map[string]interface{}{},
				)

				g.Assert(err == nil).IsFalse()
			})
		})

		g.Describe("Get Player Details", func() {
			g.It("Should get Player Details", func() {
				gameID := "player-details"
				player, err := GetTestPlayerWithMemberships(testDb, gameID, 5, 2, 3, 8)
				g.Assert(err == nil).IsTrue()

				playerDetails, err := GetPlayerDetails(
					testDb,
					player.GameID,
					player.PublicID,
				)

				g.Assert(err == nil).IsTrue()

				// Player Details
				g.Assert(playerDetails["publicID"]).Equal(player.PublicID)
				g.Assert(playerDetails["name"]).Equal(player.Name)
				g.Assert(playerDetails["metadata"]).Equal(player.Metadata)
				g.Assert(playerDetails["createdAt"]).Equal(player.CreatedAt)
				g.Assert(playerDetails["updatedAt"]).Equal(player.UpdatedAt)

				//Memberships
				g.Assert(len(playerDetails["memberships"].([]map[string]interface{}))).Equal(18)

				clans := playerDetails["clans"].(map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				g.Assert(len(approved)).Equal(5)
				g.Assert(len(denied)).Equal(2)
				g.Assert(len(banned)).Equal(3)
				g.Assert(len(pendingApplications)).Equal(0)
				g.Assert(len(pendingInvites)).Equal(8)

				approvedMembership := playerDetails["memberships"].([]map[string]interface{})[0]

				g.Assert(approvedMembership["approver"] != nil).IsTrue()
				approver := approvedMembership["approver"].(map[string]interface{})
				g.Assert(approver["name"]).Equal(player.Name)
				g.Assert(approver["publicID"]).Equal(player.PublicID)

				g.Assert(approvedMembership["approvedAt"] != nil).IsTrue()
				g.Assert(approvedMembership["approvedAt"].(int64) > 0).IsTrue()
			})

			g.It("Should get Player Details without memberships that were deleted by the player", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				err = DeleteMembership(
					testDb,
					game,
					game.PublicID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
				)
				g.Assert(err == nil).IsTrue()

				playerDetails, err := GetPlayerDetails(
					testDb,
					players[0].GameID,
					players[0].PublicID,
				)

				g.Assert(err == nil).IsTrue()

				// Player Details
				g.Assert(playerDetails["publicID"]).Equal(players[0].PublicID)
				g.Assert(playerDetails["name"]).Equal(players[0].Name)
				g.Assert(playerDetails["metadata"]).Equal(players[0].Metadata)
				g.Assert(playerDetails["createdAt"]).Equal(players[0].CreatedAt)
				g.Assert(playerDetails["updatedAt"]).Equal(players[0].UpdatedAt)

				//Memberships
				g.Assert(len(playerDetails["memberships"].([]map[string]interface{}))).Equal(0)

				clans := playerDetails["clans"].(map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				g.Assert(len(approved)).Equal(0)
				g.Assert(len(denied)).Equal(0)
				g.Assert(len(banned)).Equal(0)
				g.Assert(len(pendingApplications)).Equal(0)
				g.Assert(len(pendingInvites)).Equal(0)
			})

			g.It("Should get Player Details including owned clans", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				game.MaxClansPerPlayer = 2
				_, err = testDb.Update(game)
				g.Assert(err == nil).IsTrue()

				ownedClan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":          players[0].GameID,
					"PublicID":        uuid.NewV4().String(),
					"OwnerID":         players[0].ID,
					"Metadata":        map[string]interface{}{"x": "a"},
					"MembershipCount": 1,
				}).(*Clan)
				err = testDb.Insert(ownedClan)
				g.Assert(err == nil).IsTrue()

				playerDetails, err := GetPlayerDetails(
					testDb,
					players[0].GameID,
					players[0].PublicID,
				)

				g.Assert(err == nil).IsTrue()

				// Player Details
				g.Assert(playerDetails["publicID"]).Equal(players[0].PublicID)
				g.Assert(playerDetails["name"]).Equal(players[0].Name)
				g.Assert(playerDetails["metadata"]).Equal(players[0].Metadata)
				g.Assert(playerDetails["createdAt"]).Equal(players[0].CreatedAt)
				g.Assert(playerDetails["updatedAt"]).Equal(players[0].UpdatedAt)

				//Memberships
				g.Assert(len(playerDetails["memberships"].([]map[string]interface{}))).Equal(2)
				g.Assert(playerDetails["memberships"].([]map[string]interface{})[0]["level"]).Equal("Member")
				g.Assert(playerDetails["memberships"].([]map[string]interface{})[0]["clan"].(map[string]interface{})["publicID"]).Equal(clan.PublicID)
				g.Assert(playerDetails["memberships"].([]map[string]interface{})[1]["level"]).Equal("owner")
				g.Assert(playerDetails["memberships"].([]map[string]interface{})[1]["approved"]).IsTrue()
				g.Assert(playerDetails["memberships"].([]map[string]interface{})[1]["clan"].(map[string]interface{})["publicID"]).Equal(ownedClan.PublicID)

				clans := playerDetails["clans"].(map[string]interface{})
				owned := clans["owned"].([]map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				g.Assert(len(owned)).Equal(1)
				g.Assert(len(approved)).Equal(1)
				g.Assert(len(denied)).Equal(0)
				g.Assert(len(banned)).Equal(0)
				g.Assert(len(pendingApplications)).Equal(0)
				g.Assert(len(pendingInvites)).Equal(0)

				g.Assert(approved[0]["publicID"]).Equal(clan.PublicID)
				g.Assert(approved[0]["name"]).Equal(clan.Name)

				g.Assert(owned[0]["publicID"]).Equal(ownedClan.PublicID)
				g.Assert(owned[0]["name"]).Equal(ownedClan.Name)

				g.Assert(int(playerDetails["memberships"].([]map[string]interface{})[0]["clan"].(map[string]interface{})["membershipCount"].(int64))).Equal(clan.MembershipCount)
			})

			g.It("Should get Player Details when player has no affiliations", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				playerDetails, err := GetPlayerDetails(
					testDb,
					player.GameID,
					player.PublicID,
				)

				g.Assert(err == nil).IsTrue()

				// Player Details
				g.Assert(playerDetails["publicID"]).Equal(player.PublicID)
				g.Assert(playerDetails["name"]).Equal(player.Name)
				g.Assert(playerDetails["metadata"]).Equal(player.Metadata)
				g.Assert(playerDetails["createdAt"]).Equal(player.CreatedAt)
				g.Assert(playerDetails["updatedAt"]).Equal(player.UpdatedAt)

				//Memberships
				g.Assert(len(playerDetails["memberships"].([]map[string]interface{}))).Equal(0)

				clans := playerDetails["clans"].(map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				g.Assert(len(approved)).Equal(0)
				g.Assert(len(denied)).Equal(0)
				g.Assert(len(banned)).Equal(0)
				g.Assert(len(pendingApplications)).Equal(0)
				g.Assert(len(pendingInvites)).Equal(0)
			})

			g.It("Should return error if Player does not exist", func() {
				playerDetails, err := GetPlayerDetails(
					testDb,
					"game-id",
					"invalid-player-id",
				)

				g.Assert(playerDetails == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: invalid-player-id")
			})
		})

		g.Describe("Increment Player Membership Count", func() {
			g.It("Should work if positive value", func() {
				amount := 1
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				err = IncrementPlayerMembershipCount(testDb, player.ID, amount)
				g.Assert(err == nil).IsTrue()
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(player.MembershipCount + amount)
			})

			g.It("Should work if negative value", func() {
				amount := -1
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				err = IncrementPlayerMembershipCount(testDb, player.ID, amount)
				g.Assert(err == nil).IsTrue()
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.MembershipCount).Equal(player.MembershipCount + amount)
			})

			g.It("Should not work if non-existing Player", func() {
				err := IncrementPlayerMembershipCount(testDb, -1, 1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: -1")
			})
		})

		g.Describe("Increment Player Ownership Count", func() {
			g.It("Should work if positive value", func() {
				amount := 1
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				err = IncrementPlayerOwnershipCount(testDb, player.ID, amount)
				g.Assert(err == nil).IsTrue()
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.OwnershipCount).Equal(player.OwnershipCount + amount)
			})

			g.It("Should work if negative value", func() {
				amount := -1
				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				err = IncrementPlayerOwnershipCount(testDb, player.ID, amount)
				g.Assert(err == nil).IsTrue()
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.OwnershipCount).Equal(player.OwnershipCount + amount)
			})

			g.It("Should not work if non-existing Player", func() {
				err := IncrementPlayerOwnershipCount(testDb, -1, 1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: -1")
			})
		})
	})
}
