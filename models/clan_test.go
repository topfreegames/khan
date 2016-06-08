package models

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestClanModel(t *testing.T) {
	g := Goblin(t)
	db := GetTestDB()

	g.Describe("Clan Model", func() {
		g.It("Should create a new Clan", func() {
			clan := &Clan{
				GameID:   "test",
				ClanID:   "test-clan",
				Name:     "user-name",
				Metadata: "{}",
			}
			db.Create(clan)
			g.Assert(db.NewRecord(clan)).IsFalse()
			g.Assert(clan.ID != 0).IsTrue()

			var dbClan Clan
			db.First(&dbClan, clan.ID)

			g.Assert(dbClan.GameID).Equal(clan.GameID)
			g.Assert(dbClan.ClanID).Equal(clan.ClanID)
		})
	})
}
