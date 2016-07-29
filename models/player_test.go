// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/khan/models"

	"github.com/satori/go.uuid"
)

var _ = Describe("Player Model", func() {
	var testDb DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})
	Describe("Player Model", func() {

		Describe("Model Basic Tests", func() {
			It("Should create a new Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(player.ID).NotTo(Equal(0))

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.GameID).To(Equal(player.GameID))
				Expect(dbPlayer.PublicID).To(Equal(player.PublicID))
			})

			It("Should update a new Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())
				dt := player.UpdatedAt

				time.Sleep(time.Millisecond)

				player.Metadata = map[string]interface{}{"x": 1}
				count, err := testDb.Update(player)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(count)).To(Equal(1))
				Expect(player.UpdatedAt).To(BeNumerically(">", dt))
			})
		})

		Describe("Get Player By ID", func() {
			It("Should get existing Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.ID).To(Equal(player.ID))
			})

			It("Should not get non-existing Player", func() {
				_, err := GetPlayerByID(testDb, -1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: -1"))
			})
		})

		Describe("Get Player By Public ID", func() {
			It("Should get existing Player by Game and Player", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				dbPlayer, err := GetPlayerByPublicID(testDb, player.GameID, player.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.ID).To(Equal(player.ID))
			})

			It("Should not get non-existing Player by Game and Player", func() {
				_, err := GetPlayerByPublicID(testDb, "invalid-game", "invalid-player")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: invalid-player"))
			})
		})

		Describe("Create Player", func() {
			It("Should create a new Player with CreatePlayer", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerID := uuid.NewV4().String()
				player, err := CreatePlayer(
					testDb,
					game.PublicID,
					playerID,
					"player-name",
					map[string]interface{}{},
					false,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(player.ID).NotTo(BeEquivalentTo(0))

				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.GameID).To(Equal(player.GameID))
				Expect(dbPlayer.PublicID).To(Equal(player.PublicID))
			})
		})

		Describe("Update Player", func() {
			It("Should update a Player with UpdatePlayer", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				metadata := map[string]interface{}{"x": 1}
				updPlayer, err := UpdatePlayer(
					testDb,
					player.GameID,
					player.PublicID,
					player.Name,
					metadata,
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(updPlayer.ID).To(Equal(player.ID))

				dbPlayer, err := GetPlayerByPublicID(testDb, player.GameID, player.PublicID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Metadata["x"]).To(BeEquivalentTo(metadata["x"]))
			})

			It("Should create Player with UpdatePlayer if player does not exist", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				gameID := game.PublicID
				publicID := uuid.NewV4().String()

				metadata := map[string]interface{}{"x": "1"}
				updPlayer, err := UpdatePlayer(
					testDb,
					gameID,
					publicID,
					publicID,
					metadata,
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(updPlayer.ID).To(BeNumerically(">", 0))

				dbPlayer, err := GetPlayerByPublicID(testDb, updPlayer.GameID, updPlayer.PublicID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Metadata).To(Equal(metadata))
			})

			It("Should not update a Player with Invalid Data with UpdatePlayer", func() {
				_, err := UpdatePlayer(
					testDb,
					"-1",
					"qwe",
					"some player name",
					map[string]interface{}{},
				)

				Expect(err).To(HaveOccurred())
			})
		})

		Describe("Get Player Details", func() {
			It("Should get Player Details", func() {
				gameID := uuid.NewV4().String()
				player, err := GetTestPlayerWithMemberships(testDb, gameID, 5, 2, 3, 8)
				Expect(err).NotTo(HaveOccurred())

				playerDetails, err := GetPlayerDetails(
					testDb,
					player.GameID,
					player.PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				// Player Details
				Expect(playerDetails["publicID"]).To(Equal(player.PublicID))
				Expect(playerDetails["name"]).To(Equal(player.Name))
				Expect(playerDetails["metadata"]).To(Equal(player.Metadata))
				Expect(playerDetails["createdAt"]).To(Equal(player.CreatedAt))
				Expect(playerDetails["updatedAt"]).To(Equal(player.UpdatedAt))

				//Memberships
				Expect(len(playerDetails["memberships"].([]map[string]interface{}))).To(Equal(18))

				clans := playerDetails["clans"].(map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				Expect(len(approved)).To(Equal(5))
				Expect(len(denied)).To(Equal(2))
				Expect(len(banned)).To(Equal(3))
				Expect(len(pendingApplications)).To(Equal(0))
				Expect(len(pendingInvites)).To(Equal(8))

				approvedMembership := playerDetails["memberships"].([]map[string]interface{})[0]

				Expect(approvedMembership["approver"]).NotTo(BeEquivalentTo(nil))
				approver := approvedMembership["approver"].(map[string]interface{})
				Expect(approver["name"]).To(Equal(player.Name))
				Expect(approver["publicID"]).To(Equal(player.PublicID))
				Expect(approvedMembership["denier"]).To(BeNil())

				Expect(approvedMembership["approvedAt"]).NotTo(BeEquivalentTo(nil))
				Expect(approvedMembership["approvedAt"].(int64)).To(BeNumerically(">", 0))
				Expect(approvedMembership["message"]).To(Equal(""))

				deniedMembership := playerDetails["memberships"].([]map[string]interface{})[6]
				Expect(deniedMembership["denier"]).NotTo(BeEquivalentTo(nil))
				denier := deniedMembership["denier"].(map[string]interface{})
				Expect(denier["name"]).To(Equal(player.Name))
				Expect(denier["publicID"]).To(Equal(player.PublicID))
				Expect(deniedMembership["approver"]).To(BeNil())

				Expect(deniedMembership["deniedAt"]).NotTo(BeEquivalentTo(nil))
				Expect(deniedMembership["deniedAt"].(int64)).To(BeNumerically(">", 0))
				Expect(deniedMembership["message"]).To(Equal(""))
			})

			It("Should get Player Details without memberships that were deleted by the player", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				err = DeleteMembership(
					testDb,
					game,
					game.PublicID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
				)
				Expect(err).NotTo(HaveOccurred())

				playerDetails, err := GetPlayerDetails(
					testDb,
					players[0].GameID,
					players[0].PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				// Player Details
				Expect(playerDetails["publicID"]).To(Equal(players[0].PublicID))
				Expect(playerDetails["name"]).To(Equal(players[0].Name))
				Expect(playerDetails["metadata"]).To(Equal(players[0].Metadata))
				Expect(playerDetails["createdAt"]).To(Equal(players[0].CreatedAt))
				Expect(playerDetails["updatedAt"]).To(Equal(players[0].UpdatedAt))

				//Memberships
				Expect(len(playerDetails["memberships"].([]map[string]interface{}))).To(Equal(0))

				clans := playerDetails["clans"].(map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				Expect(len(approved)).To(Equal(0))
				Expect(len(denied)).To(Equal(0))
				Expect(len(banned)).To(Equal(0))
				Expect(len(pendingApplications)).To(Equal(0))
				Expect(len(pendingInvites)).To(Equal(0))
			})

			It("Should get owned clans as not deleted if there are deleted memberships of other clans (a.k.a fix John's bug)", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				err = DeleteMembership(
					testDb,
					game,
					game.PublicID,
					players[0].PublicID,
					clan.PublicID,
					players[0].PublicID,
				)
				Expect(err).NotTo(HaveOccurred())

				c, err := CreateClan(
					testDb,
					game.PublicID,
					"johns-bug-clan",
					"johns-bug-clan",
					players[0].PublicID,
					map[string]interface{}{"one": "one"},
					false,
					false,
					1,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(c.OwnerID).To(Equal(players[0].ID))

				playerDetails, err := GetPlayerDetails(
					testDb,
					players[0].GameID,
					players[0].PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				// Player Details
				Expect(playerDetails["publicID"]).To(Equal(players[0].PublicID))
				Expect(playerDetails["name"]).To(Equal(players[0].Name))
				Expect(playerDetails["metadata"]).To(Equal(players[0].Metadata))
				Expect(playerDetails["createdAt"]).To(Equal(players[0].CreatedAt))
				Expect(playerDetails["updatedAt"]).To(Equal(players[0].UpdatedAt))

				//Memberships
				Expect(len(playerDetails["memberships"].([]map[string]interface{}))).To(Equal(1))

				clans := playerDetails["clans"].(map[string]interface{})
				owned := clans["owned"].([]map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				Expect(len(owned)).To(Equal(1))
				Expect(len(approved)).To(Equal(0))
				Expect(len(denied)).To(Equal(0))
				Expect(len(banned)).To(Equal(0))
				Expect(len(pendingApplications)).To(Equal(0))
				Expect(len(pendingInvites)).To(Equal(0))
			})

			It("Should get Player Details including owned clans", func() {
				game, clan, _, players, _, err := GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.MaxClansPerPlayer = 2
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				ownedClan := ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":          players[0].GameID,
					"PublicID":        uuid.NewV4().String(),
					"OwnerID":         players[0].ID,
					"Metadata":        map[string]interface{}{"x": "a"},
					"MembershipCount": 1,
				}).(*Clan)
				err = testDb.Insert(ownedClan)
				Expect(err).NotTo(HaveOccurred())

				playerDetails, err := GetPlayerDetails(
					testDb,
					players[0].GameID,
					players[0].PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				// Player Details
				Expect(playerDetails["publicID"]).To(Equal(players[0].PublicID))
				Expect(playerDetails["name"]).To(Equal(players[0].Name))
				Expect(playerDetails["metadata"]).To(Equal(players[0].Metadata))
				Expect(playerDetails["createdAt"]).To(Equal(players[0].CreatedAt))
				Expect(playerDetails["updatedAt"]).To(Equal(players[0].UpdatedAt))

				//Memberships
				Expect(len(playerDetails["memberships"].([]map[string]interface{}))).To(Equal(2))
				Expect(playerDetails["memberships"].([]map[string]interface{})[0]["level"]).To(Equal("Member"))
				Expect(playerDetails["memberships"].([]map[string]interface{})[0]["clan"].(map[string]interface{})["publicID"]).To(Equal(clan.PublicID))
				Expect(playerDetails["memberships"].([]map[string]interface{})[1]["level"]).To(Equal("owner"))
				Expect(playerDetails["memberships"].([]map[string]interface{})[1]["approved"]).To(BeTrue())
				Expect(playerDetails["memberships"].([]map[string]interface{})[1]["clan"].(map[string]interface{})["publicID"]).To(Equal(ownedClan.PublicID))

				clans := playerDetails["clans"].(map[string]interface{})
				owned := clans["owned"].([]map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				Expect(len(owned)).To(Equal(1))
				Expect(len(approved)).To(Equal(1))
				Expect(len(denied)).To(Equal(0))
				Expect(len(banned)).To(Equal(0))
				Expect(len(pendingApplications)).To(Equal(0))
				Expect(len(pendingInvites)).To(Equal(0))

				Expect(approved[0]["publicID"]).To(Equal(clan.PublicID))
				Expect(approved[0]["name"]).To(Equal(clan.Name))

				Expect(owned[0]["publicID"]).To(Equal(ownedClan.PublicID))
				Expect(owned[0]["name"]).To(Equal(ownedClan.Name))

				Expect(int(playerDetails["memberships"].([]map[string]interface{})[0]["clan"].(map[string]interface{})["membershipCount"].(int64))).To(Equal(clan.MembershipCount))
			})

			It("Should get Player Details when player has no affiliations", func() {
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playerDetails, err := GetPlayerDetails(
					testDb,
					player.GameID,
					player.PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				// Player Details
				Expect(playerDetails["publicID"]).To(Equal(player.PublicID))
				Expect(playerDetails["name"]).To(Equal(player.Name))
				Expect(playerDetails["metadata"]).To(Equal(player.Metadata))
				Expect(playerDetails["createdAt"]).To(Equal(player.CreatedAt))
				Expect(playerDetails["updatedAt"]).To(Equal(player.UpdatedAt))

				//Memberships
				Expect(len(playerDetails["memberships"].([]map[string]interface{}))).To(Equal(0))

				clans := playerDetails["clans"].(map[string]interface{})
				approved := clans["approved"].([]map[string]interface{})
				denied := clans["denied"].([]map[string]interface{})
				banned := clans["banned"].([]map[string]interface{})
				pendingApplications := clans["pendingApplications"].([]map[string]interface{})
				pendingInvites := clans["pendingInvites"].([]map[string]interface{})

				Expect(len(approved)).To(Equal(0))
				Expect(len(denied)).To(Equal(0))
				Expect(len(banned)).To(Equal(0))
				Expect(len(pendingApplications)).To(Equal(0))
				Expect(len(pendingInvites)).To(Equal(0))
			})

			It("Should return error if Player does not exist", func() {
				playerDetails, err := GetPlayerDetails(
					testDb,
					"game-id",
					"invalid-player-id",
				)

				Expect(playerDetails).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: invalid-player-id"))
			})
		})

		Describe("Increment Player Membership Count", func() {
			It("Should work if positive value", func() {
				amount := 1
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				err = IncrementPlayerMembershipCount(testDb, player.ID, amount)
				Expect(err).NotTo(HaveOccurred())
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.MembershipCount).To(Equal(player.MembershipCount + amount))
			})

			It("Should work if negative value", func() {
				amount := -1
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				err = IncrementPlayerMembershipCount(testDb, player.ID, amount)
				Expect(err).NotTo(HaveOccurred())
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.MembershipCount).To(Equal(player.MembershipCount + amount))
			})

			It("Should not work if non-existing Player", func() {
				err := IncrementPlayerMembershipCount(testDb, -1, 1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: -1"))
			})
		})

		Describe("Increment Player Ownership Count", func() {
			It("Should work if positive value", func() {
				amount := 1
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				err = IncrementPlayerOwnershipCount(testDb, player.ID, amount)
				Expect(err).NotTo(HaveOccurred())
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.OwnershipCount).To(Equal(player.OwnershipCount + amount))
			})

			It("Should work if negative value", func() {
				amount := -1
				_, player, err := CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				err = IncrementPlayerOwnershipCount(testDb, player.ID, amount)
				Expect(err).NotTo(HaveOccurred())
				dbPlayer, err := GetPlayerByID(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.OwnershipCount).To(Equal(player.OwnershipCount + amount))
			})

			It("Should not work if non-existing Player", func() {
				err := IncrementPlayerOwnershipCount(testDb, -1, 1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: -1"))
			})
		})
	})
})
