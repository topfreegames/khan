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

func TestMembershipModel(t *testing.T) {
	g := Goblin(t)
	db, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Membership Model", func() {
		g.It("Should create a new Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := db.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = db.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := &Membership{
				GameID:   "test",
				ClanID:   clan.ID,
				PlayerID: player.ID,
				Level:    1,
				Approved: false,
				Denied:   false,
			}
			err = db.Insert(membership)
			g.Assert(err == nil).IsTrue()
			g.Assert(membership.ID != 0).IsTrue()

			dbMembership, err := GetMembershipByID(membership.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbMembership.GameID).Equal(membership.GameID)
			g.Assert(dbMembership.PlayerID).Equal(membership.PlayerID)
			g.Assert(dbMembership.ClanID).Equal(membership.ClanID)
		})
	})
}
