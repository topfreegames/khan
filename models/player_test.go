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
	"github.com/bluele/factory-go/factory"
	. "github.com/franela/goblin"
)

var PlayerFactory = factory.NewFactory(
	&Player{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
}).SeqInt("PlayerID", func(n int) (interface{}, error) {
	return fmt.Sprintf("player-%d", n), nil
}).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.FullName(randomdata.RandomGender), nil
}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
	return "{}", nil
})

func TestPlayerModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Player Model", func() {
		g.It("Should create a new Player", func() {
			player := &Player{
				GameID:   "test",
				PlayerID: "test-player",
				Name:     "user-name",
				Metadata: "{}",
			}
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()
			g.Assert(player.ID != 0).IsTrue()

			dbPlayer, err := GetPlayerByID(player.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbPlayer.GameID).Equal(player.GameID)
			g.Assert(dbPlayer.PlayerID).Equal(player.PlayerID)
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

		g.It("Should get existing Player by Game and Player", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := db.Insert(player)
			g.Assert(err == nil).IsTrue()

			dbPlayer, err := GetPlayerByPlayerID(player.GameID, player.PlayerID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbPlayer.ID).Equal(player.ID)
		})

		g.It("Should not get non-existing Player by Game and Player", func() {
			_, err := GetPlayerByPlayerID("invalid-game", "invalid-player")
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal("Player was not found with id: invalid-player")
		})

		g.It("Should create a new Player with CreatePlayer", func() {
			player, err := CreatePlayer(
				"create-1",
				randomdata.FullName(randomdata.RandomGender),
				"player-name",
				"{}",
			)
			fmt.Println(err)
			g.Assert(err == nil).IsTrue()
			g.Assert(player.ID != 0).IsTrue()

			dbPlayer, err := GetPlayerByID(player.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbPlayer.GameID).Equal(player.GameID)
			g.Assert(dbPlayer.PlayerID).Equal(player.PlayerID)
		})

	})
}
