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

	"github.com/bluele/factory-go/factory"
	. "github.com/franela/goblin"
)

var MembershipFactory = factory.NewFactory(
	&Membership{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
})

func TestMembershipModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Membership Model", func() {
		g.It("Should create a new Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := &Membership{
				GameID:   "test",
				ClanID:   clan.ID,
				PlayerID: player.ID,
				Level:    1,
				Approved: false,
				Denied:   false,
			}
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()
			g.Assert(membership.ID != 0).IsTrue()

			dbMembership, err := GetMembershipByID(membership.ID)
			g.Assert(err == nil).IsTrue()

			g.Assert(dbMembership.GameID).Equal(membership.GameID)
			g.Assert(dbMembership.PlayerID).Equal(membership.PlayerID)
			g.Assert(dbMembership.ClanID).Equal(membership.ClanID)
		})

		g.It("Should update a Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"PlayerID": player.ID,
				"ClanID":   clan.ID,
			}).(*Membership)
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()
			dt := membership.UpdatedAt

			membership.Approved = true
			count, err := testDb.Update(membership)
			g.Assert(err == nil).IsTrue()
			g.Assert(int(count)).Equal(1)
			g.Assert(membership.UpdatedAt > dt).IsTrue()
		})

		g.It("Should get existing Membership", func() {
			player := PlayerFactory.MustCreate().(*Player)
			err := testDb.Insert(player)
			g.Assert(err == nil).IsTrue()

			clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
				"OwnerID": player.ID,
			}).(*Clan)
			err = testDb.Insert(clan)
			g.Assert(err == nil).IsTrue()

			membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"PlayerID": player.ID,
				"ClanID":   clan.ID,
			}).(*Membership)
			err = testDb.Insert(membership)
			g.Assert(err == nil).IsTrue()

			dbMembership, err := GetMembershipByID(membership.ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(dbMembership.ID).Equal(membership.ID)
		})

		g.It("Should not get non-existing Membership", func() {
			_, err := GetMembershipByID(-1)
			g.Assert(err != nil).IsTrue()
			g.Assert(err.Error()).Equal("Membership was not found with id: -1")
		})
	})
}
