// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	"fmt"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/extensions/mongo/interfaces"
	. "github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/testing"
	"github.com/topfreegames/khan/util"

	"github.com/Pallinder/go-randomdata"
)

var _ = Describe("Clan Model", func() {
	var testDb DB
	var testMongo interfaces.MongoDB
	var faultyDb DB

	BeforeEach(func() {
		var err error

		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
		testMongo, err = GetTestMongo()
		Expect(err).NotTo(HaveOccurred())

		ConfigureAndStartGoWorkers()

		faultyDb = GetFaultyTestDB()
	})

	Describe("Clan Model", func() {
		Describe("Basic Operations", func() {
			It("Should sort clans by name", func() {
				gameID := uuid.NewV4().String()
				_, clans, err := GetTestClans(testDb, gameID, "test-sort-clan", 10)
				Expect(err).NotTo(HaveOccurred())

				sort.Sort(ClanByName(clans))

				for i := 0; i < 10; i++ {
					Expect(clans[i].Name).To(Equal(fmt.Sprintf("💩clán-test-sort-clan-%d", i)))
				}
			})

			It("Should create a new Clan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]
				Expect(clan.ID).NotTo(BeEquivalentTo(0))

				dbClan, err := GetClanByID(testDb, clan.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbClan.GameID).To(Equal(clan.GameID))
				Expect(dbClan.PublicID).To(Equal(clan.PublicID))
			})

			It("Should update a Clan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				dt := clan.UpdatedAt
				time.Sleep(time.Millisecond)

				clan.Metadata = map[string]interface{}{"x": "1"}
				count, err := testDb.Update(clan)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(BeEquivalentTo(1))
				Expect(clan.UpdatedAt).To(BeNumerically(">", dt))
			})
		})

		Describe("Get By Id", func() {
			It("Should get existing Clan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				dbClan, err := GetClanByID(testDb, clan.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbClan.ID).To(Equal(clan.ID))
			})

			It("Should not get non-existing Clan", func() {
				_, err := GetClanByID(testDb, -1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan was not found with id: -1"))
			})
		})

		Describe("Get By Public Id", func() {
			It("Should get an existing Clan by Game and PublicID", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbClan.ID).To(Equal(clan.ID))
			})

			It("Should not get a non-existing Clan by Game and PublicID", func() {
				_, err := GetClanByPublicID(testDb, "invalid-game", "invalid-clan")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan was not found with id: invalid-clan"))
			})
		})

		Describe("Get By Public Id", func() {
			It("Should get an existing Clan by Game and PublicID prefix", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				dbClan, err := GetClanByShortPublicID(testDb, clan.GameID, clan.PublicID[0:8])
				Expect(err).NotTo(HaveOccurred())
				Expect(dbClan.ID).To(Equal(clan.ID))
			})

			It("Should not get a non-existing Clan by Game and PublicID", func() {
				_, err := GetClanByPublicID(testDb, "invalid-game", "invalid-clan")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan was not found with id: invalid-clan"))
			})
		})

		Describe("Get By Public Ids", func() {
			It("Should get existing Clans by Game and PublicIDs", func() {
				gameID := uuid.NewV4().String()
				_, clan1, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())
				_, clan2, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())
				_, clan3, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())

				clans := map[string]*Clan{
					clan1.PublicID: clan1,
					clan2.PublicID: clan2,
					clan3.PublicID: clan3,
				}
				clanIDs := []string{clan1.PublicID, clan2.PublicID, clan3.PublicID}

				dbClans, err := GetClansByPublicIDs(testDb, clan1.GameID, clanIDs)
				Expect(len(dbClans)).To(Equal(3))
				Expect(err).NotTo(HaveOccurred())
				for _, dbClan := range dbClans {
					Expect(dbClan.ID).To(Equal(clans[dbClan.PublicID].ID))
				}
			})

			It("Should get only existing Clans by Game and PublicIDs, unexistent ID", func() {
				gameID := uuid.NewV4().String()
				_, clan1, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())
				_, clan2, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())
				_, clan3, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())

				clanIDs := []string{"invalid_clan", clan1.PublicID, clan2.PublicID, clan3.PublicID}

				dbClans, err := GetClansByPublicIDs(testDb, clan1.GameID, clanIDs)
				Expect(err).To(HaveOccurred())
				Expect(len(dbClans)).To(Equal(3))
				Expect(err.Error()).To(Equal(fmt.Sprintf(
					"Could not find all requested clans or the given game. GameId: %s, Missing clans: invalid_clan",
					gameID,
				)))
			})

			It("Should get only existing Clans by Game and PublicIDs, unexistent Game", func() {
				gameID := uuid.NewV4().String()
				_, clan1, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())

				clanIDs := []string{clan1.PublicID}

				dbClans, err := GetClansByPublicIDs(testDb, "invalid_game", clanIDs)
				Expect(err).To(HaveOccurred())
				Expect(len(dbClans)).To(Equal(0))
				Expect(err.Error()).To(Equal(fmt.Sprintf(
					"Could not find all requested clans or the given game. GameId: invalid_game, Missing clans: %s",
					strings.Join(clanIDs, ","),
				)))
			})
		})

		Describe("Get By Public Id and OwnerPublicID", func() {
			It("Should get an existing Clan by Game, PublicID and OwnerPublicID", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				dbClan, err := GetClanByPublicIDAndOwnerPublicID(testDb, clan.GameID, clan.PublicID, player.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbClan.ID).To(Equal(clan.ID))
				Expect(dbClan.GameID).To(Equal(clan.GameID))
				Expect(dbClan.PublicID).To(Equal(clan.PublicID))
				Expect(dbClan.Name).To(Equal(clan.Name))
				Expect(dbClan.OwnerID).To(Equal(clan.OwnerID))
			})

			It("Should not get a non-existing Clan by Game, PublicID and OwnerPublicID", func() {
				_, err := GetClanByPublicIDAndOwnerPublicID(testDb, "invalid-game", "invalid-clan", "invalid-owner-public-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan was not found with id: invalid-clan"))
			})

			It("Should not get a existing Clan by Game, PublicID and OwnerPublicID if not Clan owner", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				_, err = GetClanByPublicIDAndOwnerPublicID(testDb, clan.GameID, clan.PublicID, "invalid-owner-public-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: invalid-owner-public-id"))
			})

			Describe("Update Clan Membership Count", func() {
				It("Should work if membership is created", func() {
					previousAmount := 5
					_, clan, _, _, _, err := GetClanWithMemberships(testDb, previousAmount-1, 2, 3, 4, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, player, err := CreatePlayerFactory(testDb, clan.GameID, true)
					Expect(err).NotTo(HaveOccurred())

					membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":      player.GameID,
						"PlayerID":    player.ID,
						"ClanID":      clan.ID,
						"RequestorID": player.ID,
						"Metadata":    map[string]interface{}{"x": "a"},
						"Approved":    true,
					}).(*Membership)
					err = testDb.Insert(membership)
					Expect(err).NotTo(HaveOccurred())

					err = UpdateClanMembershipCount(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(previousAmount + 1))
				})

				It("Should work if membership is deleted", func() {
					previousAmount := 5
					_, clan, _, _, memberships, err := GetClanWithMemberships(testDb, previousAmount-1, 2, 3, 4, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = testDb.Delete(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					err = UpdateClanMembershipCount(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(previousAmount - 1))
				})

				It("Should not work if non-existing Player", func() {
					err := UpdateClanMembershipCount(testDb, -1)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Clan was not found with id: -1"))
				})
			})
		})

		Describe("Create Clan", func() {
			It("Should create a new Clan with CreateClan", func() {
				game, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

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

				Expect(err).NotTo(HaveOccurred())
				Expect(clan.ID).NotTo(BeEquivalentTo(0))

				dbClan, err := GetClanByID(testDb, clan.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbClan.GameID).To(Equal(clan.GameID))
				Expect(dbClan.PublicID).To(Equal(clan.PublicID))
				Expect(dbClan.MembershipCount).To(Equal(1))

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.OwnershipCount).To(Equal(1))
			})

			It("Should not create a new Clan with CreateClan if invalid data", func() {
				game, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

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

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("pq: value too long for type character varying(255)"))
			})

			It("Should not create a new Clan with CreateClan if reached MaxClansPerPlayer - owner", func() {
				game, _, owner, _, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

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

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s reached max clans", owner.PublicID)))
			})

			It("Should not create a new Clan with CreateClan if reached MaxClansPerPlayer - member", func() {
				game, _, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

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

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID)))
			})

			It("Should not create a new Clan with CreateClan if unexistent player", func() {
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

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID)))
			})
		})

		Describe("Update Clan", func() {
			It("Should update a Clan with UpdateClan", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				metadata := map[string]interface{}{"x": "1"}
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

				Expect(err).NotTo(HaveOccurred())
				Expect(updClan.ID).To(Equal(clan.ID))
				Expect(updClan.OwnerID).To(Equal(clan.OwnerID))

				dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbClan.Metadata["x"]).To(BeEquivalentTo(metadata["x"]))
				Expect(dbClan.AllowApplication).To(Equal(allowApplication))
				Expect(dbClan.AutoJoin).To(Equal(autoJoin))
				Expect(dbClan.OwnerID).To(Equal(clan.OwnerID))
			})

			It("Should not update a Clan if player is not the clan owner with UpdateClan", func() {
				_, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				_, player, err := CreatePlayerFactory(testDb, clan.GameID, true)
				Expect(err).NotTo(HaveOccurred())

				metadata := map[string]interface{}{"x": "1"}
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

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(
					"Player %s doesn't own clan %s. GameId: %s",
					player.PublicID,
					clan.PublicID,
					clan.GameID,
				)))
			})

			It("Should not update a Clan with Invalid Data with UpdateClan", func() {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
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

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("pq: value too long for type character varying(255)"))
			})

			Measure("Should update a Clan with UpdateClan", func(b Benchmarker) {
				player, clans, err := GetTestClans(testDb, "", "", 1)
				Expect(err).NotTo(HaveOccurred())
				clan := clans[0]

				metadata := map[string]interface{}{"x": "1"}
				allowApplication := !clan.AllowApplication
				autoJoin := !clan.AutoJoin

				runtime := b.Time("runtime", func() {
					UpdateClan(
						testDb,
						clan.GameID,
						clan.PublicID,
						clan.Name,
						player.PublicID,
						metadata,
						allowApplication,
						autoJoin,
					)
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.2), "Operation shouldn't take this long")
			}, 200)
		})

		Describe("Leave Clan", func() {
			Describe("Should leave a Clan with LeaveClan if clan owner", func() {
				It("And clan has memberships", func() {
					_, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					clan, previousOwner, newOwner, err := LeaveClan(testDb, clan.GameID, clan.PublicID)
					Expect(err).NotTo(HaveOccurred())

					Expect(previousOwner.ID).To(Equal(owner.ID))
					Expect(newOwner.ID).To(Equal(players[0].ID))

					dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.OwnerID).To(Equal(memberships[0].PlayerID))
					dbDeletedMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbDeletedMembership.DeletedBy).To(Equal(memberships[0].PlayerID))
					Expect(dbDeletedMembership.DeletedAt).To(BeNumerically(">", util.NowMilli()-1000))

					dbPlayer, err := GetPlayerByID(testDb, owner.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(0))

					dbPlayer, err = GetPlayerByID(testDb, memberships[0].PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(1))
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err = GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("And clan has no memberships", func() {
					_, clan, owner, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					clan, previousOwner, newOwner, err := LeaveClan(testDb, clan.GameID, clan.PublicID)
					Expect(err).NotTo(HaveOccurred())
					Expect(previousOwner.ID).To(Equal(owner.ID))
					Expect(newOwner).To(BeNil())

					_, err = GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID)))

					dbPlayer, err := GetPlayerByID(testDb, owner.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(0))
				})

			})

			Describe("Should not leave a Clan with LeaveClan if", func() {
				It("Clan does not exist", func() {
					_, clan, _, _, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, _, _, err = LeaveClan(testDb, clan.GameID, "-1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Clan was not found with id: -1"))
				})
			})
		})

		Describe("Transfer Clan Ownership", func() {
			Describe("Should transfer the Clan ownership with TransferClanOwnership if clan owner", func() {
				It("And first clan owner and next owner memberhip exists", func() {
					game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())
					clan, previousOwner, newOwner, err := TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						players[0].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					Expect(err).NotTo(HaveOccurred())

					Expect(previousOwner.ID).To(Equal(owner.ID))
					Expect(newOwner.ID).To(Equal(players[0].ID))

					dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.OwnerID).To(Equal(players[0].ID))

					oldOwnerMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, owner.PublicID)
					Expect(err).NotTo(HaveOccurred())
					Expect(oldOwnerMembership.CreatedAt).To(Equal(clan.CreatedAt))
					Expect(oldOwnerMembership.Level).To(Equal("CoLeader"))

					newOwnerMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(newOwnerMembership.Banned).To(BeFalse())
					Expect(newOwnerMembership.DeletedBy).To(Equal(newOwnerMembership.PlayerID))
					Expect(newOwnerMembership.DeletedAt).To(BeNumerically(">", util.NowMilli()-1000))

					dbPlayer, err := GetPlayerByID(testDb, owner.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(0))
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbPlayer, err = GetPlayerByID(testDb, newOwnerMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(1))
					Expect(dbPlayer.MembershipCount).To(Equal(0))
				})

				It("And not first clan owner and next owner membership exists", func() {
					game, clan, owner, players, memberships, err := GetClanWithMemberships(testDb, 2, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					clan, previousOwner, newOwner, err := TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						players[0].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(previousOwner.ID).To(Equal(owner.ID))
					Expect(newOwner.ID).To(Equal(players[0].ID))

					clan, previousOwner, newOwner, err = TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						players[1].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(previousOwner.ID).To(Equal(players[0].ID))
					Expect(newOwner.ID).To(Equal(players[1].ID))

					dbClan, err := GetClanByPublicID(testDb, clan.GameID, clan.PublicID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.OwnerID).To(Equal(players[1].ID))

					firstOwnerMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, owner.PublicID)
					Expect(err).NotTo(HaveOccurred())
					Expect(firstOwnerMembership.CreatedAt).To(Equal(clan.CreatedAt))
					Expect(firstOwnerMembership.Level).To(Equal("CoLeader"))

					previousOwnerMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(previousOwnerMembership.CreatedAt).To(Equal(memberships[0].CreatedAt))
					Expect(previousOwnerMembership.Level).To(Equal("CoLeader"))

					newOwnerMembership, err := GetMembershipByID(testDb, memberships[1].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(newOwnerMembership.Banned).To(BeFalse())
					Expect(newOwnerMembership.DeletedBy).To(Equal(newOwnerMembership.PlayerID))
					Expect(newOwnerMembership.DeletedAt).To(BeNumerically(">", util.NowMilli()-1000))

					dbPlayer, err := GetPlayerByID(testDb, firstOwnerMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(0))
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbPlayer, err = GetPlayerByID(testDb, previousOwnerMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(0))
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbPlayer, err = GetPlayerByID(testDb, newOwnerMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.OwnershipCount).To(Equal(1))
					Expect(dbPlayer.MembershipCount).To(Equal(0))
				})
			})

			Describe("Should not transfer the Clan ownership with TransferClanOwnership if", func() {
				It("Clan does not exist", func() {
					game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, _, _, err = TransferClanOwnership(
						testDb,
						clan.GameID,
						"-1",
						players[0].PublicID,
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Clan was not found with id: -1"))
				})

				It("Membership does not exist", func() {
					game, clan, _, _, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, _, _, err = TransferClanOwnership(
						testDb,
						clan.GameID,
						clan.PublicID,
						"some-random-player",
						game.MembershipLevels,
						game.MaxMembershipLevel,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Membership was not found with id: some-random-player"))
				})
			})
		})

		Describe("Get List of Clans", func() {
			It("Should get all clans", func() {
				player, _, err := GetTestClans(testDb, "", "", 10)
				Expect(err).NotTo(HaveOccurred())

				clans, err := GetAllClans(testDb, player.GameID)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clans)).To(Equal(10))
			})

			It("Should fail when game id is empty", func() {
				clans, err := GetAllClans(testDb, "")
				Expect(clans).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Game ID is required to retrieve Clan!"))
			})

			It("Should fail when connection fails", func() {
				clans, err := GetAllClans(faultyDb, "game-id")
				Expect(clans).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("pq: role \"khan_tet\" does not exist"))
			})
		})

		Describe("Get Clan Members", func() {
			It("Should get clan player ids", func() {
				gameID := uuid.NewV4().String()
				_, clan, _, _, _, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, gameID, uuid.NewV4().String(),
				)
				clanData, err := GetClanDetails(testDb, clan.GameID, clan, 1)
				Expect(err).NotTo(HaveOccurred())

				clanPlayers, err := GetClanMembers(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				roster := clanData["roster"].([]map[string]interface{})
				// roster + owner
				Expect(len(clanPlayers["members"].([]string))).To(Equal(len(roster) + 1))
				for _, p := range roster {
					player := p["player"].(map[string]interface{})
					Expect(clanPlayers["members"]).To(ContainElement(player["publicID"]))
				}
			})

			It("Should return empty list if no memberships", func() {
				gameID := uuid.NewV4().String()
				_, clan, _, _, _, err := GetClanWithMemberships(
					testDb, 0, 3, 4, 5, gameID, uuid.NewV4().String(),
				)
				clanData, err := GetClanDetails(testDb, clan.GameID, clan, 1)
				Expect(err).NotTo(HaveOccurred())

				clanPlayers, err := GetClanMembers(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				roster := clanData["roster"].([]map[string]interface{})
				Expect(len(roster)).To(Equal(0))
				Expect(len(clanPlayers["members"].([]string))).To(Equal(len(roster) + 1))
			})
		})

		Describe("Get Clan Details", func() {
			It("Should get clan members", func() {
				gameID := uuid.NewV4().String()
				_, clan, owner, players, memberships, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, gameID, uuid.NewV4().String(),
				)
				Expect(err).NotTo(HaveOccurred())

				clanData, err := GetClanDetails(testDb, clan.GameID, clan, 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(clanData["name"]).To(Equal(clan.Name))
				Expect(clanData["metadata"]).To(Equal(clan.Metadata))
				Expect(clanData["membershipCount"]).To(Equal(11))
				Expect(clanData["owner"].(map[string]interface{})["publicID"]).To(Equal(owner.PublicID))

				roster := clanData["roster"].([]map[string]interface{})
				Expect(len(roster)).To(Equal(10))

				pendingApplications := clanData["memberships"].(map[string]interface{})["pendingApplications"].([]map[string]interface{})
				Expect(len(pendingApplications)).To(Equal(0))

				//We do not return pending invites or applications anymore
				pendingInvites := clanData["memberships"].(map[string]interface{})["pendingInvites"].([]map[string]interface{})
				Expect(len(pendingInvites)).To(Equal(0))

				banned := clanData["memberships"].(map[string]interface{})["banned"].([]map[string]interface{})
				Expect(len(banned)).To(Equal(4))

				denied := clanData["memberships"].(map[string]interface{})["denied"].([]map[string]interface{})
				Expect(len(denied)).To(Equal(3))

				playerDict := map[string]*Player{}
				for i := 0; i < 22; i++ {
					playerDict[players[i].PublicID] = players[i]
				}

				membershipDict := map[int64]*Membership{}
				for i := 0; i < 22; i++ {
					membershipDict[memberships[i].PlayerID] = memberships[i]
				}

				for _, playerData := range roster {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					Expect(name).To(Equal(playerDict[pid].Name))
					membershipLevel := playerData["level"]
					Expect(membershipLevel).To(Equal(membershipDict[playerDict[pid].ID].Level))

					//Approval
					approver := player["approver"].(map[string]interface{})
					Expect(approver["name"]).To(Equal(playerDict[pid].Name))
					Expect(approver["publicID"]).To(Equal(playerDict[pid].PublicID))

					Expect(player["denier"]).To(BeNil())
				}

				for _, playerData := range pendingInvites {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					Expect(name).To(Equal(playerDict[pid].Name))
					membershipLevel := playerData["level"]
					Expect(membershipLevel).To(Equal(membershipDict[playerDict[pid].ID].Level))
				}

				for _, playerData := range pendingApplications {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					message := player["message"].(string)
					Expect(name).To(Equal(playerDict[pid].Name))
					membershipLevel := playerData["level"]
					Expect(membershipLevel).To(Equal(membershipDict[playerDict[pid].ID].Level))
					Expect(message).To(Equal("Accept me"))
				}

				for _, playerData := range banned {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					Expect(name).To(Equal(playerDict[pid].Name))
					Expect(playerData["level"]).To(BeNil())
				}

				for _, playerData := range denied {
					player := playerData["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					Expect(name).To(Equal(playerDict[pid].Name))
					Expect(playerData["level"]).To(BeNil())

					//Approval
					denier := player["denier"].(map[string]interface{})
					Expect(denier["name"]).To(Equal(playerDict[pid].Name))
					Expect(denier["publicID"]).To(Equal(playerDict[pid].PublicID))

					Expect(player["approver"]).To(BeNil())
				}
			})

			It("Should not get deleted clan members", func() {
				gameID := uuid.NewV4().String()
				_, clan, _, players, memberships, err := GetClanWithMemberships(
					testDb, 10, 0, 0, 0, gameID, uuid.NewV4().String(),
				)
				Expect(err).NotTo(HaveOccurred())

				memberships[9].DeletedAt = util.NowMilli()
				memberships[9].DeletedBy = clan.OwnerID
				_, err = testDb.Update(memberships[9])
				Expect(err).NotTo(HaveOccurred())

				clanData, err := GetClanDetails(testDb, clan.GameID, clan, 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(clanData["name"]).To(Equal(clan.Name))
				Expect(clanData["metadata"]).To(Equal(clan.Metadata))

				roster := clanData["roster"].([]map[string]interface{})
				Expect(len(roster)).To(Equal(9))

				playerDict := map[string]*Player{}
				for i := 0; i < len(roster); i++ {
					playerDict[players[i].PublicID] = players[i]
				}

				for i := 0; i < len(roster); i++ {
					player := roster[i]["player"].(map[string]interface{})
					pid := player["publicID"].(string)
					name := player["name"].(string)
					Expect(name).To(Equal(playerDict[pid].Name))
				}
			})

			It("Should get clan details even if no members", func() {
				gameID := uuid.NewV4().String()
				_, clan, _, _, _, err := GetClanWithMemberships(
					testDb, 0, 0, 0, 0, gameID, uuid.NewV4().String(),
				)
				Expect(err).NotTo(HaveOccurred())
				clan.AllowApplication = true
				clan.AutoJoin = true
				_, err = testDb.Update(clan)
				Expect(err).NotTo(HaveOccurred())

				clanData, err := GetClanDetails(testDb, clan.GameID, clan, 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(clanData["name"]).To(Equal(clan.Name))
				Expect(clanData["metadata"]).To(Equal(clan.Metadata))
				Expect(clanData["allowApplication"]).To(Equal(clan.AllowApplication))
				Expect(clanData["autoJoin"]).To(Equal(clan.AutoJoin))
				Expect(clanData["membershipCount"]).To(Equal(1))
				roster := clanData["roster"].([]map[string]interface{})
				Expect(len(roster)).To(Equal(0))
			})

			It("Should fail if clan does not exist", func() {
				clanData, err := GetClanDetails(testDb, "fake-game-id", &Clan{PublicID: "fake-public-id"}, 1)
				Expect(clanData).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan was not found with id: fake-public-id"))
			})
		})

		Describe("Get Clan Summary", func() {
			It("Should get clan members", func() {
				gameID := uuid.NewV4().String()
				_, clan, _, _, _, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, gameID, uuid.NewV4().String(),
				)
				Expect(err).NotTo(HaveOccurred())

				clanData, err := GetClanSummary(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(clanData["membershipCount"]).To(Equal(clan.MembershipCount))
				Expect(clanData["publicID"]).To(Equal(clan.PublicID))
				Expect(clanData["metadata"]).To(Equal(clan.Metadata))
				Expect(clanData["name"]).To(Equal(clan.Name))
				Expect(clanData["allowApplication"]).To(Equal(clan.AllowApplication))
				Expect(clanData["autoJoin"]).To(Equal(clan.AutoJoin))
				Expect(len(clanData)).To(Equal(6))
			})

			It("Should fail if clan does not exist", func() {
				clanData, err := GetClanDetails(testDb, "fake-game-id", &Clan{PublicID: "fake-public-id"}, 1)
				Expect(clanData).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan was not found with id: fake-public-id"))
			})

		})

		Describe("Get Clans Summaries", func() {
			It("Should get clan members", func() {
				gameID := uuid.NewV4().String()

				_, clan1, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())
				_, clan2, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())
				_, clan3, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())

				clans := map[string]*Clan{
					clan1.PublicID: clan1,
					clan2.PublicID: clan2,
					clan3.PublicID: clan3,
				}
				clanIDs := []string{clan1.PublicID, clan2.PublicID, clan3.PublicID}

				clansSummaries, err := GetClansSummaries(testDb, clan1.GameID, clanIDs)
				Expect(err).NotTo(HaveOccurred())

				clansSummariesArr := clansSummaries
				Expect(len(clansSummariesArr)).To(Equal(3))
				for _, clanSummary := range clansSummariesArr {
					clan := clans[clanSummary["publicID"].(string)]
					Expect(clanSummary["membershipCount"]).To(Equal(clan.MembershipCount))
					Expect(clanSummary["metadata"]).To(Equal(clan.Metadata))
					Expect(clanSummary["name"]).To(Equal(clan.Name))
					Expect(clanSummary["allowApplication"]).To(Equal(clan.AllowApplication))
					Expect(clanSummary["autoJoin"]).To(Equal(clan.AutoJoin))
					Expect(len(clanSummary)).To(Equal(6))
				}
			})

			It("Should retrieve only existent clans", func() {
				gameID := uuid.NewV4().String()

				_, clan1, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String())
				Expect(err).NotTo(HaveOccurred())
				_, clan2, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())
				_, clan3, _, _, _, err := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID,
					uuid.NewV4().String(), true)
				Expect(err).NotTo(HaveOccurred())

				clanIDs := []string{clan1.PublicID, clan2.PublicID, clan3.PublicID, "unexistent_clan"}

				clansSummaries, err := GetClansSummaries(testDb, clan1.GameID, clanIDs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(
					"Could not find all requested clans or the given game. GameId: %s, Missing clans: unexistent_clan",
					gameID,
				)))
				Expect(len(clansSummaries)).To(Equal(3))
				for _, clanSummary := range clansSummaries {
					clanSummaryObj := clanSummary
					Expect(len(clanSummaryObj)).To(Equal(6))
				}
			})

			It("Should fail if game does not exist", func() {
				gameID := uuid.NewV4().String()
				clanID := uuid.NewV4().String()
				_, clan1, _, _, _, err1 := GetClanWithMemberships(testDb, 0, 0, 0, 0, gameID, clanID)
				Expect(err1).To(BeNil())

				clanIDs := []string{clan1.PublicID}

				clansSummaries, err := GetClansSummaries(testDb, "unexistent_game", clanIDs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(
					"Could not find all requested clans or the given game. GameId: unexistent_game, Missing clans: %s",
					strings.Join(clanIDs, ","),
				)))

				Expect(len(clansSummaries)).To(Equal(0))
			})
		})

		Describe("Clan Search", func() {
			var player *Player
			var realClans []*Clan

			BeforeEach(func() {
				var err error
				player, realClans, err = GetTestClans(
					testDb, "", "clan-search-clan", 10,
				)
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(500 * time.Millisecond)
			})

			It("Should return clan by search term", func() {
				err := testing.CreateClanNameTextIndexInMongo(GetTestMongo, player.GameID)
				Expect(err).NotTo(HaveOccurred())
				clans, err := SearchClan(testDb, testMongo, player.GameID, "SEARCH", 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clans)).To(Equal(10))
			})

			It("Should return clan by unicode search term", func() {
				err := testing.CreateClanNameTextIndexInMongo(GetTestMongo, player.GameID)
				Expect(err).NotTo(HaveOccurred())
				clans, err := SearchClan(testDb, testMongo, player.GameID, "💩clán", 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clans)).To(Equal(10))
			})

			It("Should return clan by full public ID as search term", func() {
				searchClanID := realClans[0].PublicID
				clans, err := SearchClan(testDb, testMongo, player.GameID, searchClanID, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clans)).To(Equal(1))
			})

			It("Should return clan by short public ID as search term", func() {
				dbClan, err := GetTestClanWithRandomPublicIDAndName(testDb, player.GameID, player.ID)
				Expect(err).NotTo(HaveOccurred())
				searchClanID := dbClan.PublicID[:8]
				clans, err := SearchClan(testDb, testMongo, player.GameID, searchClanID, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clans)).To(Equal(1))
			})

			It("Should return empty list if search term is not found", func() {
				err := testing.CreateClanNameTextIndexInMongo(GetTestMongo, player.GameID)
				Expect(err).NotTo(HaveOccurred())
				clans, err := SearchClan(testDb, testMongo, player.GameID, "qwfjur", 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clans)).To(Equal(0))
			})

			It("Should return invalid response if empty term", func() {
				_, err := SearchClan(testDb, testMongo, "some-game-id", "", 10)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("A search term was not provided to find a clan."))
			})
		})

		Describe("Get Clan and Owner", func() {
			It("Should return clan and owner", func() {
				_, clan, owner, _, _, err := GetClanWithMemberships(
					testDb, 10, 3, 4, 5, "", "",
				)
				Expect(err).NotTo(HaveOccurred())

				dbClan, dbOwner, err := GetClanAndOwnerByPublicID(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbClan.ID).To(Equal(clan.ID))
				Expect(dbOwner.ID).To(Equal(owner.ID))
			})
		})
	})
})
