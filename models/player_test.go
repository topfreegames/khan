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
	db := GetTestDB()

	g.Describe("Player Model", func() {
		g.It("Should create a new Player", func() {
			player := &Player{
				GameID:   "test",
				PlayerID: "test-player",
				Name:     "user-name",
				Metadata: "{}",
			}
			err := db.Insert(player)
			g.Assert(err == nil).IsTrue()
			g.Assert(player.ID != 0).IsTrue()

			dbPlayer, err := GetPlayerByID(player.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbPlayer.GameID).Equal(player.GameID)
			g.Assert(dbPlayer.PlayerID).Equal(player.PlayerID)
		})
	})
}
