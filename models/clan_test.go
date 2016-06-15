// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
)

func TestClanModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()
	g.Assert(err == nil).IsTrue()
	faultyDb := GetFaultyTestDB()

	g.Describe("Clan Model", func() {
		g.Describe("Basic Operations", func() {
			g.It("Should sort clans by name", func() {
				_, clans, err := GetTestClans(testDb, "test", "test-sort-clan", 10)
				g.Assert(err == nil).IsTrue()

				sort.Sort(ClanByName(clans))

				for i := 0; i < 10; i++ {
					g.Assert(clans[i].Name).Equal(fmt.Sprintf("test-sort-clan-%d", i))
				}
			})

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

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbClan.GameID).Equal(clan.GameID)
				g.Assert(dbClan.PublicID).Equal(clan.PublicID)
			})

			g.It("Should update a Clan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				dt := clan.UpdatedAt

				clan.Metadata = "{ \"x\": 1 }"
				count, err := testDb.Update(clan)
				g.Assert(err == nil).IsTrue()
				g.Assert(int(count)).Equal(1)
				g.Assert(clan.UpdatedAt > dt).IsTrue()
			})
		})

		g.Describe("Get By Id", func() {
			g.It("Should get existing Clan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.ID).Equal(clan.ID)
			})

			g.It("Should not get non-existing Clan", func() {
				_, err = GetClanByID(testDb, -1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Clan was not found with id: -1")
			})
		})

		g.Describe("Get By Public Id", func() {
			g.It("Should get an existing Clan by Game and PublicID", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.ID).Equal(clan.ID)
			})

			g.It("Should not get a non-existing Clan by Game and PublicID", func() {
				_, err = GetClanByPublicID(testDb, "invalid-game", "invalid-clan")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Clan was not found with id: invalid-clan")
			})
		})

		g.Describe("Get By Public Id and OwnerPublicID", func() {
			g.It("Should get an existing Clan by Game, PublicID and OwnerPublicID", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				dbClan, err := GetClanByPublicIDAndOwnerPublicID(testDb, clan.GameID, clan.PublicID, player.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.ID).Equal(clan.ID)
				g.Assert(dbClan.GameID).Equal(clan.GameID)
				g.Assert(dbClan.PublicID).Equal(clan.PublicID)
				g.Assert(dbClan.Name).Equal(clan.Name)
				g.Assert(dbClan.OwnerID).Equal(clan.OwnerID)
			})

			g.It("Should not get a non-existing Clan by Game, PublicID and OwnerPublicID", func() {
				_, err = GetClanByPublicIDAndOwnerPublicID(testDb, "invalid-game", "invalid-clan", "invalid-owner-public-id")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Clan was not found with id: invalid-clan")
			})

			g.It("Should not get a existing Clan by Game, PublicID and OwnerPublicID if not Clan owner", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				_, err = GetClanByPublicIDAndOwnerPublicID(testDb, clan.GameID, clan.PublicID, "invalid-owner-public-id")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
			})
		})

		g.Describe("Create Clan", func() {
			g.It("Should create a new Clan with CreateClan", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				clan, err := CreateClan(
					testDb,
					player.GameID,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					player.PublicID,
					"{}",
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(clan.ID != 0).IsTrue()

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbClan.GameID).Equal(clan.GameID)
				g.Assert(dbClan.PublicID).Equal(clan.PublicID)
			})

			g.It("Should not create a new Clan with CreateClan if invalid data", func() {
				player := PlayerFactory.MustCreate().(*Player)
				err := testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				_, err = CreateClan(
					testDb,
					player.GameID,
					randomdata.FullName(randomdata.RandomGender),
					"clan-name",
					player.PublicID,
					"it-will-fail-because-metadata-is-not-a-json",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("pq: invalid input syntax for type json")
			})

			g.It("Should not create a new Clan with CreateClan if unexistent player", func() {
				playerPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateClan(
					testDb,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					"clan-name",
					playerPublicID,
					"{}",
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID))
			})
		})

		g.Describe("Update Clan", func() {
			g.It("Should update a Clan with UpdateClan", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				metadata := "{\"x\": 1}"
				updClan, err := UpdateClan(
					testDb,
					clan.GameID,
					clan.PublicID,
					clan.Name,
					player.PublicID,
					metadata,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updClan.ID).Equal(clan.ID)

				dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbClan.Metadata).Equal(metadata)
			})

			g.It("Should not update a Clan if player is not the clan owner with UpdateClan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				player := PlayerFactory.MustCreate().(*Player)
				err = testDb.Insert(player)
				g.Assert(err == nil).IsTrue()

				metadata := "{\"x\": 1}"
				_, err = UpdateClan(
					testDb,
					clan.GameID,
					clan.PublicID,
					clan.Name,
					player.PublicID,
					metadata,
				)

				g.Assert(err == nil).IsFalse()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
			})

			g.It("Should not update a Clan with Invalid Data with UpdateClan", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				metadata := "it will not work because i am not a json"
				_, err = UpdateClan(
					testDb,
					clan.GameID,
					clan.PublicID,
					clan.Name,
					player.PublicID,
					metadata,
				)

				g.Assert(err == nil).IsFalse()
				g.Assert(err.Error()).Equal("pq: invalid input syntax for type json")
			})
		})

		g.Describe("Get List of Clans", func() {
			g.It("Should get all clans", func() {
				player, _, err := GetTestClans(testDb, "", "", 10)
				g.Assert(err == nil).IsTrue()

				clans, err := GetAllClans(testDb, player.GameID)
				g.Assert(err == nil).IsTrue()
				g.Assert(len(clans)).Equal(10)
			})

			g.It("Should fail when game id is empty", func() {
				clans, err := GetAllClans(testDb, "")
				g.Assert(clans == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Game ID is required to retrieve Clan!")
			})

			g.It("Should fail when connection fails", func() {
				clans, err := GetAllClans(faultyDb, "game-id")
				g.Assert(clans == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("pq: role \"khan_tet\" does not exist")
			})
		})

		g.Describe("Get Clan Details", func() {
			g.It("Should get clan members", func() {
				clan, _, players, _, err := GetClanWithMemberships(
					testDb, 10, "clan-details", "clan-details-clan",
				)
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanDetails(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)
				members := clanData["members"].([]map[string]interface{})
				g.Assert(len(members)).Equal(10)

				for i := 0; i < 10; i++ {
					g.Assert(members[i]["playerName"]).Equal(players[i].Name)
				}
			})

			g.It("Should not get deleted clan members", func() {
				clan, _, players, memberships, err := GetClanWithMemberships(
					testDb, 10, "more-clan-details", "more-clan-details-clan",
				)
				g.Assert(err == nil).IsTrue()

				memberships[9].DeletedAt = time.Now().UnixNano()
				memberships[9].DeletedBy = clan.OwnerID
				_, err = testDb.Update(memberships[9])
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanDetails(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)
				members := clanData["members"].([]map[string]interface{})
				g.Assert(len(members)).Equal(9)

				for i := 0; i < 9; i++ {
					g.Assert(members[i]["playerName"]).Equal(players[i].Name)
				}
			})

			g.It("Should get clan details even if no members", func() {
				clan, _, _, _, err := GetClanWithMemberships(
					testDb, 0, "clan-details-2", "clan-details-2-clan",
				)
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanDetails(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)
				members := clanData["members"].([]map[string]interface{})
				g.Assert(len(members)).Equal(0)
			})

			g.It("Should fail if clan does not exist", func() {
				clanData, err := GetClanDetails(testDb, "fake-game-id", "fake-public-id")
				g.Assert(clanData == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Clan was not found with id: fake-public-id")
			})

		})

		g.Describe("Clan Search", func() {
			g.It("Should return clan by search term", func() {
				player, _, err := GetTestClans(
					testDb, "", "clan-search-clan", 10,
				)
				g.Assert(err == nil).IsTrue()

				clans, err := SearchClan(testDb, player.GameID, "SEARCH")
				g.Assert(err == nil).IsTrue()

				g.Assert(len(clans)).Equal(10)
			})
		})
	})
}
