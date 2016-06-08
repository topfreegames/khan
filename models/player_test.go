package models

import (
	"testing"

	. "github.com/franela/goblin"
)

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
			db.Create(player)
			g.Assert(db.NewRecord(player)).IsFalse()
			g.Assert(player.ID != 0).IsTrue()

			var dbPlayer Player
			db.First(&dbPlayer, player.ID)

			g.Assert(dbPlayer.GameID).Equal(player.GameID)
			g.Assert(dbPlayer.PlayerID).Equal(player.PlayerID)
		})
	})
}
