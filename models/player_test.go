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
)

func TestPlayerModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()

	g.Assert(err == nil).IsTrue()

	g.Describe("Player Model", func() {

		g.Describe("Model Basic Tests", func() {
			g.It("Should create a new Player", func() {
				player := &Player{
					GameID:   "test",
					PublicID: "test-player",
					Name:     "user-name",
					Metadata: "{}",
				}
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()
				g.Assert(player.ID != 0).IsTrue()

				dbPlayer, err := GetPlayerByID(player.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.GameID).Equal(player.GameID)
				g.Assert(dbPlayer.PublicID).Equal(player.PublicID)
			})

			g.It("Should update a new Player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()
				dt := player.UpdatedAt

				player.Metadata = "{ \"x\": 1 }"
				count, err := testDb.Update(player)
				g.Assert(err == nil).IsTrue()
				g.Assert(int(count)).Equal(1)
				g.Assert(player.UpdatedAt > dt).IsTrue()
			})
		})

		g.Describe("Get Player By ID", func() {
			g.It("Should get existing Player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				dbPlayer, err := GetPlayerByID(player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.ID).Equal(player.ID)
			})

			g.It("Should not get non-existing Player", func() {
				_, err := GetPlayerByID(-1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: -1")
			})

			g.It("Should not get deleted Player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				player.DeletedAt = time.Now().UnixNano()
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				dbPlayer, err := GetPlayerByID(player.ID)
				g.Assert(dbPlayer == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player was not found with id: %d", player.ID))
			})
		})

		g.Describe("Get Player By Public ID", func() {
			g.It("Should get existing Player by Game and Player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := db.Insert(player)
				g.Assert(err == nil).IsTrue()

				dbPlayer, err := GetPlayerByPublicID(player.GameID, player.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.ID).Equal(player.ID)
			})

			g.It("Should not get non-existing Player by Game and Player", func() {
				_, err := GetPlayerByPublicID("invalid-game", "invalid-player")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: invalid-player")
			})

			g.It("Should not get deleted Player by Game and Player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				player.DeletedAt = time.Now().UnixNano()
				err := db.Insert(player)
				g.Assert(err == nil).IsTrue()

				dbPlayer, err := GetPlayerByPublicID(player.GameID, player.PublicID)
				g.Assert(dbPlayer == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player was not found with id: %s", player.PublicID))
			})
		})

		g.Describe("Create Player", func() {
			g.It("Should create a new Player with CreatePlayer", func() {
				player, err := CreatePlayer(
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					"player-name",
					"{}",
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(player.ID != 0).IsTrue()

				dbPlayer, err := GetPlayerByID(player.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.GameID).Equal(player.GameID)
				g.Assert(dbPlayer.PublicID).Equal(player.PublicID)
			})
		})

		g.Describe("Update Player", func() {
			g.It("Should update a Player with UpdatePlayer", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				metadata := "{\"x\": 1}"
				updPlayer, err := UpdatePlayer(
					player.GameID,
					player.PublicID,
					player.Name,
					metadata,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updPlayer.ID).Equal(player.ID)

				dbPlayer, err := GetPlayerByPublicID(player.GameID, player.PublicID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.Metadata).Equal(metadata)
			})

			g.It("Should not update a deleted player", func() {
				player := PlayerFactory.MustCreate().(*Player)
				player.DeletedAt = time.Now().UnixNano()
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				metadata := "{\"x\": 1}"
				updPlayer, err := UpdatePlayer(
					player.GameID,
					player.PublicID,
					player.Name,
					metadata,
				)

				g.Assert(updPlayer == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player was not found with id: %s", player.PublicID))
			})

			g.It("Should not update a Player with Invalid Data with UpdatePlayer", func() {
				_, err := UpdatePlayer(
					"-1",
					"qwe",
					"some player name",
					"{}",
				)

				g.Assert(err == nil).IsFalse()
			})
		})

		g.Describe("Delete Player", func() {
			g.It("Should delete a Player with DeletePlayer", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				deleted, err := DeletePlayer(
					"-1",
					"-1",
				)

				g.Assert(deleted == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Player was not found with id: -1")
			})
		})
	})
}
