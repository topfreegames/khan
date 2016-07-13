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
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
)

func TestClanModel(t *testing.T) {
	g := Goblin(t)
	testDb, _err := GetTestDB()
	g.Assert(_err == nil).IsTrue()
	faultyDb := GetFaultyTestDB()

	g.Describe("Clan Model", func() {
		g.Describe("Basic Operations", func() {
			g.It("Should sort clans by name", func() {
				_, clans, err := GetTestClans(testDb, "test", "test-sort-clan", 10)
				g.Assert(err == nil).IsTrue()

				sort.Sort(ClanByName(clans))

				for i := 0; i < 10; i++ {
					g.Assert(clans[i].Name).Equal(fmt.Sprintf("💩clán-test-sort-clan-%d", i))
				}
			})

			g.It("Should create a new Clan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]
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
				time.Sleep(time.Millisecond)

				clan.Metadata = map[string]interface{}{"x": 1}
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
				_, err := GetClanByID(testDb, -1)
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
				_, err := GetClanByPublicID(testDb, "invalid-game", "invalid-clan")
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
				_, err := GetClanByPublicIDAndOwnerPublicID(testDb, "invalid-game", "invalid-clan", "invalid-owner-public-id")
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

			g.Describe("Increment Clan Membership Count", func() {
				g.It("Should work if positive value", func() {
					amount := 1
					_, clans, err := GetTestClans(testDb, "", "", 1)
					g.Assert(err == nil).IsTrue()

					err = IncrementClanMembershipCount(testDb, clans[0].ID, amount)
					g.Assert(err == nil).IsTrue()
					dbClan, err := GetClanByID(testDb, clans[0].ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbClan.MembershipCount).Equal(clans[0].MembershipCount + amount)
				})

				g.It("Should work if negative value", func() {
					amount := -1
					_, clans, err := GetTestClans(testDb, "", "", 1)
					g.Assert(err == nil).IsTrue()

					err = IncrementClanMembershipCount(testDb, clans[0].ID, amount)
					g.Assert(err == nil).IsTrue()
					dbClan, err := GetClanByID(testDb, clans[0].ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbClan.MembershipCount).Equal(clans[0].MembershipCount + amount)
				})

				g.It("Should not work if non-existing Player", func() {
					err := IncrementClanMembershipCount(testDb, -1, 1)
					g.Assert(err != nil).IsTrue()
					g.Assert(err.Error()).Equal("Clan was not found with id: -1")
				})
			})
		})

		g.Describe("Create Clan", func() {
			g.It("Should create a new Clan with CreateClan", func() {
				game, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				clan, err := CreateClan(
					testDb,
					player.GameID,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					player.PublicID,
					map[string]interface{}{},
					true,
					false,
					game.MaxClansPerPlayer,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(clan.ID != 0).IsTrue()

				dbClan, err := GetClanByID(testDb, clan.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbClan.GameID).Equal(clan.GameID)
				g.Assert(dbClan.PublicID).Equal(clan.PublicID)
				g.Assert(dbClan.MembershipCount).Equal(1)

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbPlayer.OwnershipCount).Equal(1)
			})

			g.It("Should not create a new Clan with CreateClan if invalid data", func() {
				game, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				_, err = CreateClan(
					testDb,
					player.GameID,
					strings.Repeat("a", 256),
					"clan-name",
					player.PublicID,
					map[string]interface{}{},
					true,
					false,
					game.MaxClansPerPlayer,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("pq: value too long for type character varying(255)")
			})

			g.It("Should not create a new Clan with CreateClan if reached MaxClansPerPlayer - owner", func() {
				game, _, owner, _, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				_, err = CreateClan(
					testDb,
					owner.GameID,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					owner.PublicID,
					map[string]interface{}{},
					true,
					false,
					game.MaxClansPerPlayer,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s reached max clans", owner.PublicID))
			})

			g.It("Should not create a new Clan with CreateClan if reached MaxClansPerPlayer - member", func() {
				game, _, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				g.Assert(err == nil).IsTrue()

				_, err = CreateClan(
					testDb,
					game.PublicID,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					players[0].PublicID,
					map[string]interface{}{},
					true,
					false,
					game.MaxClansPerPlayer,
				)

				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID))
			})

			g.It("Should not create a new Clan with CreateClan if unexistent player", func() {
				game, _, err := CreatePlayerFactory(testDb, "")
				playerPublicID := randomdata.FullName(randomdata.RandomGender)
				_, err = CreateClan(
					testDb,
					"create-1",
					randomdata.FullName(randomdata.RandomGender),
					"clan-name",
					playerPublicID,
					map[string]interface{}{},
					true,
					false,
					game.MaxClansPerPlayer,
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

				metadata := map[string]interface{}{"x": 1}
				allowApplication := !clan.AllowApplication
				autoJoin := !clan.AutoJoin
				updClan, err := UpdateClan(
					testDb,
					clan.GameID,
					clan.PublicID,
					clan.Name,
					player.PublicID,
					metadata,
					allowApplication,
					autoJoin,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updClan.ID).Equal(clan.ID)

				dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.Metadata).Equal(metadata)
				g.Assert(dbClan.AllowApplication).Equal(allowApplication)
				g.Assert(dbClan.AutoJoin).Equal(autoJoin)
			})

			g.It("Should not update a Clan if player is not the clan owner with UpdateClan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				_, player, err := CreatePlayerFactory(testDb, "")
				g.Assert(err == nil).IsTrue()

				metadata := map[string]interface{}{"x": 1}
				_, err = UpdateClan(
					testDb,
					clan.GameID,
					clan.PublicID,
					clan.Name,
					player.PublicID,
					metadata,
					clan.AllowApplication,
					clan.AutoJoin,
				)

				g.Assert(err == nil).IsFalse()
				g.Assert(err.Error()).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
			})

			g.It("Should not update a Clan with Invalid Data with UpdateClan", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				g.Assert(err == nil).IsTrue()
				clan := clans[0]

				metadata := map[string]interface{}{}
				_, err = UpdateClan(
					testDb,
					clan.GameID,
					clan.PublicID,
					strings.Repeat("a", 256),
					player.PublicID,
					metadata,
					clan.AllowApplication,
					clan.AutoJoin,
				)

				g.Assert(err == nil).IsFalse()
				g.Assert(err.Error()).Equal("pq: value too long for type character varying(255)")
			})
		})

		g.Describe("Leave Clan", func() {
			g.Describe("Should leave a Clan with LeaveClan if clan owner", func() {
				g.It("And clan has memberships", func() {
					_, clan, owner, _, memberships, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()

					err = LeaveClan(testDb, clan.GameID, clan.PublicID)
					g.Assert(err == nil).IsTrue()

					dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbClan.OwnerID).Equal(memberships[0].PlayerID)
					dbDeletedMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbDeletedMembership.DeletedBy).Equal(owner.ID)
					g.Assert(dbDeletedMembership.DeletedAt > time.Now().UnixNano()/1000000-1000).IsTrue()

					dbPlayer, err := GetPlayerByID(testDb, owner.ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(0)

					dbPlayer, err = GetPlayerByID(testDb, memberships[0].PlayerID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(1)
					g.Assert(dbPlayer.MembershipCount).Equal(0)

					dbClan, err = GetClanByID(testDb, clan.ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbClan.MembershipCount).Equal(1)
				})

				g.It("And clan has no memberships", func() {
					_, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()

					err = LeaveClan(testDb, clan.GameID, clan.PublicID)
					g.Assert(err == nil).IsTrue()
					_, err = GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					g.Assert(err != nil).IsTrue()
					g.Assert(err.Error()).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))

					dbPlayer, err := GetPlayerByID(testDb, owner.ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(0)
				})
			})

			g.Describe("Should not leave a Clan with LeaveClan if", func() {
				g.It("Clan does not exist", func() {
					_, clan, _, _, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()

					err = LeaveClan(testDb, clan.GameID, "-1")
					g.Assert(err != nil).IsTrue()
					g.Assert(err.Error()).Equal("Clan was not found with id: -1")
				})
			})
		})

		g.Describe("Transfer Clan Ownership", func() {
			g.Describe("Should transfer the Clan ownership with TransferClanOwnership if clan owner", func() {
				g.It("And first clan owner and next owner memberhip exists", func() {
					game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()
					err = TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						players[0].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					g.Assert(err == nil).IsTrue()

					dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbClan.OwnerID).Equal(players[0].ID)

					oldOwnerMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, owner.PublicID)
					g.Assert(err == nil).IsTrue()
					g.Assert(oldOwnerMembership.CreatedAt).Equal(clan.CreatedAt)
					g.Assert(oldOwnerMembership.Level).Equal("CoLeader")

					newOwnerMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(newOwnerMembership.Banned).IsFalse()
					g.Assert(newOwnerMembership.DeletedBy).Equal(newOwnerMembership.PlayerID)
					g.Assert(newOwnerMembership.DeletedAt > time.Now().UnixNano()/1000000-1000).IsTrue()

					dbPlayer, err := GetPlayerByID(testDb, owner.ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(0)
					g.Assert(dbPlayer.MembershipCount).Equal(1)

					dbPlayer, err = GetPlayerByID(testDb, newOwnerMembership.PlayerID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(1)
					g.Assert(dbPlayer.MembershipCount).Equal(0)
				})

				g.It("And not first clan owner and next owner membership exists", func() {
					game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 2, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()

					err = TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						players[0].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					g.Assert(err == nil).IsTrue()

					err = TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						players[1].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					g.Assert(err == nil).IsTrue()

					dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbClan.OwnerID).Equal(players[1].ID)

					firstOwnerMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, owner.PublicID)
					g.Assert(err == nil).IsTrue()
					g.Assert(firstOwnerMembership.CreatedAt).Equal(clan.CreatedAt)
					g.Assert(firstOwnerMembership.Level).Equal("CoLeader")

					previousOwnerMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(previousOwnerMembership.CreatedAt).Equal(memberships[0].CreatedAt)
					g.Assert(previousOwnerMembership.Level).Equal("CoLeader")

					newOwnerMembership, err := GetMembershipByID(testDb, memberships[1].ID)
					g.Assert(err == nil).IsTrue()
					g.Assert(newOwnerMembership.Banned).IsFalse()
					g.Assert(newOwnerMembership.DeletedBy).Equal(newOwnerMembership.PlayerID)
					g.Assert(newOwnerMembership.DeletedAt > time.Now().UnixNano()/1000000-1000).IsTrue()

					dbPlayer, err := GetPlayerByID(testDb, firstOwnerMembership.PlayerID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(0)
					g.Assert(dbPlayer.MembershipCount).Equal(1)

					dbPlayer, err = GetPlayerByID(testDb, previousOwnerMembership.PlayerID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(0)
					g.Assert(dbPlayer.MembershipCount).Equal(1)

					dbPlayer, err = GetPlayerByID(testDb, newOwnerMembership.PlayerID)
					g.Assert(err == nil).IsTrue()
					g.Assert(dbPlayer.OwnershipCount).Equal(1)
					g.Assert(dbPlayer.MembershipCount).Equal(0)
				})
			})

			g.Describe("Should not transfer the Clan ownership with TransferClanOwnership if", func() {
				g.It("Clan does not exist", func() {
					game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()

					err = TransferClanOwnership(
						testDb,
						clan.GameID,
						"-1",
						players[0].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					g.Assert(err != nil).IsTrue()
					g.Assert(err.Error()).Equal("Clan was not found with id: -1")
				})

				g.It("Membership does not exist", func() {
					game, clan, _, _, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					g.Assert(err == nil).IsTrue()

					err = TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						"some-random-player",
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					g.Assert(err != nil).IsTrue()
					g.Assert(err.Error()).Equal("Membership was not found with id: some-random-player")
				})
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
				_, clan, owner, players, memberships, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, "clan-details", "clan-details-clan",
				)
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanDetails(testDb, clan.GameID, clan.PublicID, 1)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)
				g.Assert(clanData["membershipCount"]).Equal(11)
				g.Assert(clanData["owner"].(map[string]interface{})["publicID"]).Equal(owner.PublicID)

				roster := clanData["roster"].([]map[string]interface{})
				g.Assert(len(roster)).Equal(10)

				pendingApplications := clanData["memberships"].(map[string]interface{})["pendingApplications"].([]map[string]interface{})
				g.Assert(len(pendingApplications)).Equal(0)

				pendingInvites := clanData["memberships"].(map[string]interface{})["pendingInvites"].([]map[string]interface{})
				g.Assert(len(pendingInvites)).Equal(5)

				banned := clanData["memberships"].(map[string]interface{})["banned"].([]map[string]interface{})
				g.Assert(len(banned)).Equal(4)

				denied := clanData["memberships"].(map[string]interface{})["denied"].([]map[string]interface{})
				g.Assert(len(denied)).Equal(3)

				playerDict := map[string]*Player{}
				for i := 0; i < 22; i++ {
					playerDict[players[i].PublicID] = players[i]
				}

				membershipDict := map[int]*Membership{}
				for i := 0; i < 22; i++ {
					membershipDict[memberships[i].PlayerID] = memberships[i]
				}

				for _, playerData := range roster {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					g.Assert(name).Equal(playerDict[pid].Name)
					membershipLevel := playerData["level"]
					g.Assert(membershipLevel).Equal(membershipDict[playerDict[pid].ID].Level)

					//Approval
					approver := player["approver"].(map[string]interface{})
					g.Assert(approver["name"]).Equal(playerDict[pid].Name)
					g.Assert(approver["publicID"]).Equal(playerDict[pid].PublicID)

					g.Assert(player["denier"] == nil).IsTrue()
				}

				for _, playerData := range pendingInvites {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					g.Assert(name).Equal(playerDict[pid].Name)
					membershipLevel := playerData["level"]
					g.Assert(membershipLevel).Equal(membershipDict[playerDict[pid].ID].Level)
				}

				for _, playerData := range banned {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					g.Assert(name).Equal(playerDict[pid].Name)
					g.Assert(playerData["level"]).Equal(nil)
				}

				for _, playerData := range denied {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					g.Assert(name).Equal(playerDict[pid].Name)
					g.Assert(playerData["level"]).Equal(nil)

					//Approval
					denier := player["denier"].(map[string]interface{})
					g.Assert(denier["name"]).Equal(playerDict[pid].Name)
					g.Assert(denier["publicID"]).Equal(playerDict[pid].PublicID)

					g.Assert(player["approver"] == nil).IsTrue()
				}
			})

			g.It("Should not get deleted clan members", func() {
				_, clan, _, players, memberships, err := GetClanWithMemberships(
					testDb, 10, 0, 0, 0, "more-clan-details", "more-clan-details-clan",
				)
				g.Assert(err == nil).IsTrue()

				memberships[9].DeletedAt = time.Now().UnixNano() / 1000000
				memberships[9].DeletedBy = clan.OwnerID
				_, err = testDb.Update(memberships[9])
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanDetails(testDb, clan.GameID, clan.PublicID, 1)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)

				roster := clanData["roster"].([]map[string]interface{})
				g.Assert(len(roster)).Equal(9)

				playerDict := map[string]*Player{}
				for i := 0; i < len(roster); i++ {
					playerDict[players[i].PublicID] = players[i]
				}

				for i := 0; i < len(roster); i++ {
					player := roster[i]["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					g.Assert(name).Equal(playerDict[pid].Name)
				}
			})

			g.It("Should get clan details even if no members", func() {
				_, clan, _, _, _, err := GetClanWithMemberships(
					testDb, 0, 0, 0, 0, "clan-details-2", "clan-details-2-clan",
				)
				g.Assert(err == nil).IsTrue()
				clan.AllowApplication = true
				clan.AutoJoin = true
				_, err = testDb.Update(clan)
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanDetails(testDb, clan.GameID, clan.PublicID, 1)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)
				g.Assert(clanData["allowApplication"]).Equal(clan.AllowApplication)
				g.Assert(clanData["autoJoin"]).Equal(clan.AutoJoin)
				g.Assert(clanData["membershipCount"]).Equal(1)
				roster := clanData["roster"].([]map[string]interface{})
				g.Assert(len(roster)).Equal(0)
			})

			g.It("Should fail if clan does not exist", func() {
				clanData, err := GetClanDetails(testDb, "fake-game-id", "fake-public-id", 1)
				g.Assert(clanData == nil).IsTrue()
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Clan was not found with id: fake-public-id")
			})
		})

		g.Describe("Get Clan Summary", func() {
			g.It("Should get clan members", func() {
				_, clan, _, _, _, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, "clan-summary", "clan-summary-clan",
				)
				g.Assert(err == nil).IsTrue()

				clanData, err := GetClanSummary(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(clanData["membershipCount"]).Equal(clan.MembershipCount)
				g.Assert(clanData["publicID"]).Equal(clan.PublicID)
				g.Assert(clanData["metadata"]).Equal(clan.Metadata)
				g.Assert(clanData["name"]).Equal(clan.Name)
				g.Assert(clanData["allowApplication"]).Equal(clan.AllowApplication)
				g.Assert(clanData["autoJoin"]).Equal(clan.AutoJoin)
				g.Assert(len(clanData)).Equal(6)
			})

			g.It("Should fail if clan does not exist", func() {
				clanData, err := GetClanDetails(testDb, "fake-game-id", "fake-public-id", 1)
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

			g.It("Should return clan by unicode search term", func() {
				player, _, err := GetTestClans(
					testDb, "", "clan-search-clan", 10,
				)
				g.Assert(err == nil).IsTrue()

				clans, err := SearchClan(testDb, player.GameID, "💩clán")
				g.Assert(err == nil).IsTrue()

				g.Assert(len(clans)).Equal(10)
			})

			g.It("Should return empty list if search term is not found", func() {
				player, _, err := GetTestClans(
					testDb, "", "clan-search-clan-2", 10,
				)
				g.Assert(err == nil).IsTrue()

				clans, err := SearchClan(testDb, player.GameID, "qwfjur")
				g.Assert(err == nil).IsTrue()

				g.Assert(len(clans)).Equal(0)
			})

			g.It("Should return invalid response if empty term", func() {
				_, err := SearchClan(testDb, "some-game-id", "")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("A search term was not provided to find a clan.")
			})
		})

		g.Describe("Get Clan and Owner", func() {
			g.It("Should return clan and owner", func() {
				_, clan, owner, _, _, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, "", "",
				)
				g.Assert(err == nil).IsTrue()

				dbClan, dbOwner, err := GetClanAndOwnerByPublicID(testDb, clan.GameID, clan.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbClan.ID).Equal(clan.ID)
				g.Assert(dbOwner.ID).Equal(owner.ID)
			})
		})
	})
}
