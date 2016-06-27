// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package main

import (
	"fmt"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

func createTestData(db models.DB, games, clansPerGame, playersPerClan, playersWithoutClan, pendingMembershipsPerClan int, progress func() bool) error {
	for gameIndex := 0; gameIndex < games; gameIndex++ {
		game := models.GameFactory.MustCreateWithOption(util.JSON{
			"PublicID": uuid.NewV4().String(),
		}).(*models.Game)
		err := db.Insert(game)
		if err != nil {
			panic(err.Error())
		}
		progress()

		for playerIndex := 0; playerIndex < playersWithoutClan; playerIndex++ {
			player := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID":   game.PublicID,
				"PublicID": uuid.NewV4().String(),
			}).(*models.Player)
			err = db.Insert(player)
			if err != nil {
				panic(err.Error())
			}
			progress()
		}

		for clanIndex := 0; clanIndex < clansPerGame; clanIndex++ {
			owner := models.PlayerFactory.MustCreateWithOption(util.JSON{
				"GameID":   game.PublicID,
				"PublicID": uuid.NewV4().String(),
			}).(*models.Player)
			err = db.Insert(owner)
			if err != nil {
				panic(err.Error())
			}

			clan := models.ClanFactory.MustCreateWithOption(util.JSON{
				"GameID":   game.PublicID,
				"PublicID": uuid.NewV4().String(),
				"OwnerID":  owner.ID,
			}).(*models.Clan)
			err = db.Insert(clan)
			if err != nil {
				panic(err.Error())
			}
			progress()

			for playerIndex := 0; playerIndex < playersPerClan; playerIndex++ {
				player := models.PlayerFactory.MustCreateWithOption(util.JSON{
					"GameID":   game.PublicID,
					"PublicID": uuid.NewV4().String(),
				}).(*models.Player)
				err = db.Insert(player)
				if err != nil {
					panic(err.Error())
				}

				membership := models.MembershipFactory.MustCreateWithOption(util.JSON{
					"GameID":      game.PublicID,
					"PlayerID":    player.ID,
					"ClanID":      clan.ID,
					"RequestorID": owner.ID,
					"Metadata":    util.JSON{"x": "a"},
					"Accepted":    true,
				}).(*models.Membership)

				err = db.Insert(membership)
				if err != nil {
					panic(err.Error())
				}
				progress()
			}

			for membershipIndex := 0; membershipIndex < pendingMembershipsPerClan; membershipIndex++ {
				player := models.PlayerFactory.MustCreateWithOption(util.JSON{
					"GameID":   game.PublicID,
					"PublicID": uuid.NewV4().String(),
				}).(*models.Player)
				err = db.Insert(player)
				if err != nil {
					panic(err.Error())
				}

				membership := models.MembershipFactory.MustCreateWithOption(util.JSON{
					"GameID":      game.PublicID,
					"PlayerID":    player.ID,
					"ClanID":      clan.ID,
					"RequestorID": owner.ID,
					"Metadata":    util.JSON{"x": "a"},
				}).(*models.Membership)

				err = db.Insert(membership)
				if err != nil {
					panic(err.Error())
				}
				progress()
			}
		}
	}

	return nil
}

func main() {
	start := time.Now()

	games := 20
	clansPerGame := 1000
	playersPerClan := 50
	playersWithoutClan := 500000
	pendingMembershipsPerClan := 250

	totalClans := games * clansPerGame
	totalPlayers := totalClans*playersPerClan + games*playersWithoutClan
	totalPending := totalClans * pendingMembershipsPerClan
	totalOps := games + totalClans + totalPlayers + totalPending

	uiprogress.Start()                     // start rendering
	bar := uiprogress.AddBar(totalOps - 1) // Add a new bar
	bar.AppendCompleted()
	bar.PrependElapsed()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		ellapsed := time.Now().Sub(start)
		itemsPerSec := float64(b.Current()+1) / ellapsed.Seconds()
		timeToComplete := float64(totalOps) / itemsPerSec / 60.0 / 60.0
		text := fmt.Sprintf("%d/%d (%.2fhs to complete)", b.Current()+1, totalOps, timeToComplete)
		return strutil.Resize(text, uint(len(text)))
	})

	testDb, err := models.GetPerfDB()
	if err != nil {
		panic(err.Error())
	}

	createTestData(testDb, games, clansPerGame, playersPerClan, playersWithoutClan, pendingMembershipsPerClan, bar.Incr)
}
