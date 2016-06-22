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
				player := &Player{
					GameID:   "test",
					PublicID: uuid.NewV4().String(),
					Name:     uuid.NewV4().String(),
					Metadata: "{}",
				}
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()
				g.Assert(player.ID != 0).IsTrue()

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbPlayer.GameID).Equal(player.GameID)
				g.Assert(dbPlayer.PublicID).Equal(player.PublicID)
			})

			g.It("Should update a new Player", func() {
				player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()
				dt := player.UpdatedAt

				time.Sleep(time.Millisecond)

				player.Metadata = "{ \"x\": 1 }"
				count, err := testDb.Update(player)
				g.Assert(err == nil).IsTrue()
				g.Assert(int(count)).Equal(1)
				g.Assert(player.UpdatedAt > dt).IsTrue()
			})
		})

		g.Describe("Get Player By ID", func() {
			g.It("Should get existing Player", func() {
				player, err := CreatePlayerFactory(testDb, "")
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
				player, err := CreatePlayerFactory(testDb, "")
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
					"{}",
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
				player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				metadata := "{\"x\": 1}"
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

			g.It("Should not update a Player with Invalid Data with UpdatePlayer", func() {
				_, err := UpdatePlayer(
					testDb,
					"-1",
					"qwe",
					"some player name",
					"{}",
				)

				g.Assert(err == nil).IsFalse()
			})
		})
	})
}
