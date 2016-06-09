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

var ClanFactory = factory.NewFactory(
	&Clan{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
}).SeqInt("ClanID", func(n int) (interface{}, error) {
	return fmt.Sprintf("clan-%d", n), nil
}).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.FullName(randomdata.RandomGender), nil
}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
	return "{}", nil
})

func TestClanModel(t *testing.T) {
	g := Goblin(t)
	db, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Clan Model", func() {
		g.It("Should create a new Clan", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := db.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := &Clan{
				GameID:   "test",
				ClanID:   "test-clan-2",
				Name:     "clan-name",
				Metadata: "{}",
				OwnerID:  player.ID,
			}
			err = db.Insert(clan)
			g.Assert(err == nil).IsTrue()
			g.Assert(clan.ID != 0).IsTrue()

			dbClan, err := GetClanByID(clan.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbClan.GameID).Equal(clan.GameID)
			g.Assert(dbClan.ClanID).Equal(clan.ClanID)
		})

		g.It("Should update a Clan", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := db.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = db.Insert(clan)
			g.Assert(err == nil).IsTrue()
			dt := clan.UpdatedAt

			clan.Metadata = "{ \"x\": 1 }"
			count, err := db.Update(clan)
			g.Assert(err == nil).IsTrue()
			g.Assert(int(count)).Equal(1)
			g.Assert(clan.UpdatedAt > dt).IsTrue()
		})

		g.It("Should get existing Clan", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := db.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = db.Insert(clan)
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
