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
	"github.com/topfreegames/khan/util"
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

				player.Metadata = util.JSON{"x": 1}
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
					util.JSON{},
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

				metadata := util.JSON{"x": 1}
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

				metadata := util.JSON{"x": "1"}
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
					util.JSON{},
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
				g.Assert(len(playerDetails["memberships"].([]util.JSON))).Equal(18)

				clans := playerDetails["clans"].(util.JSON)
				approved := clans["approved"].([]util.JSON)
				denied := clans["denied"].([]util.JSON)
				banned := clans["banned"].([]util.JSON)
				pendingApplications := clans["pendingApplications"].([]util.JSON)
				pendingInvites := clans["pendingInvites"].([]util.JSON)

				g.Assert(len(approved)).Equal(5)
				g.Assert(len(denied)).Equal(2)
				g.Assert(len(banned)).Equal(3)
				g.Assert(len(pendingApplications)).Equal(0)
				g.Assert(len(pendingInvites)).Equal(8)
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
				g.Assert(len(playerDetails["memberships"].([]util.JSON))).Equal(0)

				clans := playerDetails["clans"].(util.JSON)
				approved := clans["approved"].([]util.JSON)
				denied := clans["denied"].([]util.JSON)
				banned := clans["banned"].([]util.JSON)
				pending := clans["pending"].([]util.JSON)

				g.Assert(len(approved)).Equal(0)
				g.Assert(len(denied)).Equal(0)
				g.Assert(len(banned)).Equal(0)
				g.Assert(len(pending)).Equal(0)
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
