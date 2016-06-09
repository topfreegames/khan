// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestClanModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Clan Model", func() {
		g.It("Should create a new Clan", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := &Clan{
				GameID:   "test",
				PublicID: "test-clan-2",
				Name:     "clan-name",
				Metadata: "{}",
				OwnerID:  player.ID,
			}
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()
			g.Assert(clan.ID != 0).IsTrue()

			dbClan, err := GetClanByID(clan.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbClan.GameID).Equal(clan.GameID)
			g.Assert(dbClan.PublicID).Equal(clan.PublicID)
		})

		g.It("Should update a Clan", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()
			dt := clan.UpdatedAt

			clan.Metadata = "{ \"x\": 1 }"
			count, err := testDb.Update(clan)
			g.Assert(err == nil).IsTrue()
			g.Assert(int(count)).Equal(1)
			g.Assert(clan.UpdatedAt > dt).IsTrue()
		})

		g.It("Should get existing Clan", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			dbClan, err := GetClanByID(clan.ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbClan.ID).Equal(clan.ID)
		})

		g.It("Should not get non-existing Clan", func() {
			_, err := GetClanByID(-1)
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal("Clan was not found with id: -1")
		})
	})
}
