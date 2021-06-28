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
	egorp "github.com/topfreegames/extensions/v9/gorp/interfaces"
	. "github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
	"github.com/topfreegames/khan/testing"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"

	uuid "github.com/satori/go.uuid"
)

var _ = Describe("Player Model", func() {
	var testDb egorp.Database
	var logger zap.Logger

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		logger = testing.NewMockLogger()
		Expect(err).NotTo(HaveOccurred())
	})
	Describe("Player Model", func() {

		Describe("Model Basic Tests", func() {
			It("Should set CreatedAt and UpdatedAt as same value on creating a new player", func() {
				gameID := uuid.NewV4().String()
				game := fixtures.GameFactory.MustCreateWithOption(map[string]interface{}{
					"PublicID": gameID,
				}).(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				player := &Player{
					GameID:   game.PublicID,
					PublicID: "publicID",
				}
				err = testDb.Insert(player)
				Expect(err).NotTo(HaveOccurred())
				Expect(player.ID).NotTo(Equal(0))

				Expect(player.CreatedAt).Should(BeNumerically(">", 0))
				Expect(player.UpdatedAt).To(Equal(player.CreatedAt))
			})

			It("Should update a new Player", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())
				dt := player.UpdatedAt

				time.Sleep(time.Millisecond)

				player.Metadata = map[string]interface{}{"x": "a"}
				count, err := testDb.Update(player)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(count)).To(Equal(1))
				Expect(player.UpdatedAt).To(BeNumerically(">", dt))
			})
		})

		Describe("Get Player By ID", func() {
			It("Should get existing Player", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.ID).To(Equal(player.ID))
				Expect(dbPlayer.Name).To(Equal(player.Name))
			})

			It("Should not get non-existing Player", func() {
				_, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), -1)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: -1"))
			})

			It("Should decrypt Player.Name", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				encryptedName, err := util.EncryptData(player.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				name := player.Name
				player.Name = encryptedName
				count, err := testDb.Update(player)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(count)).To(Equal(1))

				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.ID).To(Equal(player.ID))
				Expect(dbPlayer.Name).To(Equal(name))
			})
		})

		Describe("Get Player By Public ID", func() {
			It("Should get existing Player by Game and Player", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				dbPlayer, err := GetPlayerByPublicID(testDb, fixtures.GetEncryptionKey(), player.GameID, player.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.ID).To(Equal(player.ID))
			})

			It("Should not get non-existing Player by Game and Player", func() {
				_, err := GetPlayerByPublicID(testDb, fixtures.GetEncryptionKey(), "invalid-game", "invalid-player")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: invalid-player"))
			})

			It("Should decrypt Player.Name", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				encryptedName, err := util.EncryptData(player.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				name := player.Name
				player.Name = encryptedName
				count, err := testDb.Update(player)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(count)).To(Equal(1))

				dbPlayer, err := GetPlayerByPublicID(testDb, fixtures.GetEncryptionKey(), player.GameID, player.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.ID).To(Equal(player.ID))
				Expect(dbPlayer.Name).To(Equal(name))
			})
		})

		Describe("Create Player", func() {
			It("Should create a new Player with CreatePlayer", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerID := uuid.NewV4().String()
				player, err := CreatePlayer(
					testDb,
					logger,
					[]byte(""),
					game.PublicID,
					playerID,
					"player-name",
					map[string]interface{}{},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(player.ID).NotTo(BeEquivalentTo(0))

				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.GameID).To(Equal(player.GameID))
				Expect(dbPlayer.PublicID).To(Equal(player.PublicID))
			})

			It("Should create a new Player encrypting the Player.Name", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerName := uuid.NewV4().String()
				playerPublicID := uuid.NewV4().String()
				player, err := CreatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					game.PublicID,
					playerPublicID,
					playerName,
					map[string]interface{}{},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(player.Name).To(Equal(playerName))

				var dbPlayer *Player
				err = testDb.SelectOne(&dbPlayer, "select * from players where public_id = $1", player.PublicID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Name).NotTo(Equal(playerName))

				decryptedName, err := util.DecryptData(dbPlayer.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				Expect(decryptedName).To(Equal(player.Name))
			})

			It("Should create a EncryptedPlayer to trace players encryption process when have valid encryption key", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerPublicID := uuid.NewV4().String()
				playerName := uuid.NewV4().String()
				player, err := CreatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					game.PublicID,
					playerPublicID,
					playerName,
					map[string]interface{}{},
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(encryptedPlayer.PlayerID).To(Equal(player.ID))

			})

			It("Should not create a EncryptedPlayer to trace players encryption process when have invalid encryption key", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerPublicID := uuid.NewV4().String()
				playerName := uuid.NewV4().String()
				player, err := CreatePlayer(
					testDb,
					logger,
					[]byte(""),
					game.PublicID,
					playerPublicID,
					playerName,
					map[string]interface{}{},
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).To(HaveOccurred())
				Expect(encryptedPlayer).To(BeNil())

			})
		})

		Describe("Update Player", func() {
			It("Should update a Player with UpdatePlayer", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				metadata := map[string]interface{}{"x": "a"}
				updatedPlayer, err := UpdatePlayer(
					testDb,
					logger,
					[]byte(""),
					player.GameID,
					player.PublicID,
					player.Name,
					metadata,
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(updatedPlayer.ID).To(Equal(player.ID))

				dbPlayer, err := GetPlayerByPublicID(testDb, fixtures.GetEncryptionKey(), player.GameID, player.PublicID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Metadata["x"]).To(BeEquivalentTo(metadata["x"]))
			})

			It("Should update a Player encrypting the Player.Name", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playerName := uuid.NewV4().String()

				updatedPlayer, err := UpdatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					player.GameID,
					player.PublicID,
					playerName,
					player.Metadata,
				)

				Expect(updatedPlayer.Name).To(Equal(playerName))

				var dbPlayer *Player
				err = testDb.SelectOne(&dbPlayer, "select * from players where id = $1", updatedPlayer.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Name).NotTo(Equal(updatedPlayer.Name))

				decryptedPlayerName, err := util.DecryptData(dbPlayer.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				Expect(decryptedPlayerName).To(Equal(playerName))

			})

			It("Should create a EncryptedPlayer to trace players encryption process when updating a player", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playerName := uuid.NewV4().String()
				metadata := map[string]interface{}{"x": "1"}
				_, err = UpdatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					player.GameID,
					player.PublicID,
					playerName,
					metadata,
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(encryptedPlayer.PlayerID).To(Equal(player.ID))

			})

			It("Should not create a EncryptedPlayer to trace players encryption process when updating a player don't have valid encryption key", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playerName := uuid.NewV4().String()
				metadata := map[string]interface{}{"x": "1"}
				_, err = UpdatePlayer(
					testDb,
					logger,
					[]byte(""),
					player.GameID,
					player.PublicID,
					playerName,
					metadata,
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).To(HaveOccurred())
				Expect(encryptedPlayer).To(BeNil())

			})

			It("Should create Player with UpdatePlayer if player does not exist", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				gameID := game.PublicID
				publicID := uuid.NewV4().String()

				metadata := map[string]interface{}{"x": "1"}
				updPlayer, err := UpdatePlayer(
					testDb,
					logger,
					[]byte(""),
					gameID,
					publicID,
					publicID,
					metadata,
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(updPlayer.ID).To(BeNumerically(">", 0))

				dbPlayer, err := GetPlayerByPublicID(testDb, fixtures.GetEncryptionKey(), updPlayer.GameID, updPlayer.PublicID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Metadata).To(Equal(metadata))
			})

			It("Should create Player with UpdatePlayer if player does not exist and encrypt player.Name", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				gameID := game.PublicID
				publicID := uuid.NewV4().String()
				playerName := uuid.NewV4().String()

				metadata := map[string]interface{}{"x": "1"}
				createdPlayer, err := UpdatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					gameID,
					publicID,
					playerName,
					metadata,
				)

				Expect(err).NotTo(HaveOccurred())

				Expect(createdPlayer.Name).To(Equal(playerName))

				var dbPlayer *Player
				err = testDb.SelectOne(&dbPlayer, "select * from players where id = $1", createdPlayer.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbPlayer.Name).NotTo(Equal(createdPlayer.Name))

				decryptedPlayerName, err := util.DecryptData(dbPlayer.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				Expect(decryptedPlayerName).To(Equal(playerName))
			})

			It("Should create a EncryptedPlayer to trace players encryption process when creating a player", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerPublicID := uuid.NewV4().String()
				playerName := uuid.NewV4().String()
				metadata := map[string]interface{}{"x": "1"}
				player, err := UpdatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					game.PublicID,
					playerPublicID,
					playerName,
					metadata,
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(encryptedPlayer.PlayerID).To(Equal(player.ID))

			})

			It("Should not create a EncryptedPlayer to trace players encryption process when creating a player without valid encryption key", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				playerPublicID := uuid.NewV4().String()
				playerName := uuid.NewV4().String()
				metadata := map[string]interface{}{"x": "1"}
				player, err := UpdatePlayer(
					testDb,
					logger,
					[]byte(""),
					game.PublicID,
					playerPublicID,
					playerName,
					metadata,
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).To(HaveOccurred())
				Expect(encryptedPlayer).To(BeNil())

			})

			It("Should return player normally if EncryptedPlayer is already created", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				metadata := map[string]interface{}{"x": "1"}
				_, err = CreatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					game.PublicID,
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					metadata,
				)
				Expect(err).NotTo(HaveOccurred())

				player, err := UpdatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					game.PublicID,
					uuid.NewV4().String(),
					uuid.NewV4().String(),
					metadata,
				)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(encryptedPlayer.PlayerID).To(Equal(player.ID))

			})

			It("Should not update a Player with Invalid Data with UpdatePlayer", func() {
				_, err := UpdatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
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
				_, player, err := fixtures.GetTestPlayerWithMemberships(testDb, gameID, 5, 2, 3, 8)
				Expect(err).NotTo(HaveOccurred())

				playerDetails, err := GetPlayerDetails(
					testDb,
					fixtures.GetEncryptionKey(),
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
				//Hard limited at 5 pending memberships
				Expect(len(playerDetails["memberships"].([]map[string]interface{}))).To(Equal(15))

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
				Expect(len(pendingInvites)).To(Equal(5))

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

				pendingInvite := playerDetails["memberships"].([]map[string]interface{})[14]
				Expect(pendingInvite["requestor"]).NotTo(BeEquivalentTo(nil))
				requestor := pendingInvite["requestor"].(map[string]interface{})
				Expect(requestor["name"]).NotTo(BeNil())
				Expect(requestor["publicID"]).NotTo(BeNil())
				Expect(requestor["level"]).NotTo(BeNil())
				Expect(pendingInvite["approver"]).To(BeNil())
				Expect(pendingInvite["deniedAt"]).To(BeEquivalentTo(0))
				Expect(pendingInvite["approvedAt"]).To(BeEquivalentTo(0))
				Expect(pendingInvite["deletedAt"]).To(BeEquivalentTo(0))
			})

			It("Should get Player Details without memberships that were deleted by the player", func() {
				game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				_, err = DeleteMembership(
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
					fixtures.GetEncryptionKey(),
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
				game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				_, err = DeleteMembership(
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
					fixtures.GetEncryptionKey(),
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
					fixtures.GetEncryptionKey(),
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
				game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.MaxClansPerPlayer = 2
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				ownedClan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
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
					fixtures.GetEncryptionKey(),
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
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playerDetails, err := GetPlayerDetails(
					testDb,
					fixtures.GetEncryptionKey(),
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

			It("Should decrypt Player.Name in Details", func() {
				gameID := uuid.NewV4().String()
				owner, player, err := fixtures.GetTestPlayerWithMemberships(testDb, gameID, 5, 2, 3, 8)
				Expect(err).NotTo(HaveOccurred())

				testing.UpdateEncryptingTestPlayer(testDb, fixtures.GetEncryptionKey(), owner)
				testing.UpdateEncryptingTestPlayer(testDb, fixtures.GetEncryptionKey(), player)

				playerDetails, err := GetPlayerDetails(
					testDb,
					fixtures.GetEncryptionKey(),
					player.GameID,
					player.PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				testing.DecryptTestPlayer(fixtures.GetEncryptionKey(), player)

				Expect(playerDetails["name"]).To(Equal(player.Name))

				approvedMembership := playerDetails["memberships"].([]map[string]interface{})[0]

				Expect(approvedMembership["approver"]).NotTo(BeEquivalentTo(nil))
				approver := approvedMembership["approver"].(map[string]interface{})
				Expect(approver["name"]).To(Equal(player.Name))

				deniedMembership := playerDetails["memberships"].([]map[string]interface{})[6]
				Expect(deniedMembership["denier"]).NotTo(BeEquivalentTo(nil))
				denier := deniedMembership["denier"].(map[string]interface{})
				Expect(denier["name"]).To(Equal(player.Name))

				testing.DecryptTestPlayer(fixtures.GetEncryptionKey(), owner)

				pendingInvite := playerDetails["memberships"].([]map[string]interface{})[14]
				Expect(pendingInvite["requestor"]).NotTo(BeEquivalentTo(nil))
				requestor := pendingInvite["requestor"].(map[string]interface{})
				Expect(requestor["name"]).To(Equal(owner.Name))
			})

			It("Should return error if Player does not exist", func() {
				playerDetails, err := GetPlayerDetails(
					testDb,
					fixtures.GetEncryptionKey(),
					"game-id",
					"invalid-player-id",
				)

				Expect(playerDetails).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: invalid-player-id"))
			})
		})

		Describe("Update Player Membership Count", func() {
			It("Should work if membership is created", func() {
				prevMemberships := 5
				_, player, err := fixtures.GetTestPlayerWithMemberships(testDb, "", prevMemberships, 2, 3, 4)
				Expect(err).NotTo(HaveOccurred())

				clan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":          player.GameID,
					"PublicID":        uuid.NewV4().String(),
					"OwnerID":         player.ID,
					"Metadata":        map[string]interface{}{"x": "a"},
					"MembershipCount": 1,
				}).(*Clan)
				err = testDb.Insert(clan)
				Expect(err).NotTo(HaveOccurred())

				membership := fixtures.MembershipFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":      player.GameID,
					"PlayerID":    player.ID,
					"ClanID":      clan.ID,
					"RequestorID": player.ID,
					"Metadata":    map[string]interface{}{"x": "a"},
					"Approved":    true,
					"Denied":      false,
					"Banned":      false,
				}).(*Membership)
				err = testDb.Insert(membership)
				Expect(err).NotTo(HaveOccurred())

				err = UpdatePlayerMembershipCount(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())

				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.MembershipCount).To(Equal(prevMemberships + 1))
			})

			It("Should work if membership is deleted", func() {
				prevMemberships := 5
				_, player, err := fixtures.GetTestPlayerWithMemberships(testDb, "", prevMemberships, 2, 3, 4)
				Expect(err).NotTo(HaveOccurred())

				var membership *Membership
				err = testDb.SelectOne(&membership, "SELECT * FROM memberships WHERE game_id=$1 AND player_id=$2 AND approved=true LIMIT 1", player.GameID, player.ID)
				Expect(membership).ToNot(BeNil())
				membership.DeletedAt = util.NowMilli()
				_, err = testDb.Update(membership)
				Expect(err).NotTo(HaveOccurred())

				err = UpdatePlayerMembershipCount(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())

				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.MembershipCount).To(Equal(prevMemberships - 1))
			})

			It("Should not work if non-existing Player", func() {
				err := UpdatePlayerMembershipCount(testDb, -1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: -1"))
			})
		})

		Describe("Update Player Ownership Count", func() {
			It("Should work if clan is created", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())
				player.OwnershipCount = 123 // some random value
				_, err = testDb.Update(player)
				Expect(err).NotTo(HaveOccurred())

				clan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":          player.GameID,
					"PublicID":        uuid.NewV4().String(),
					"OwnerID":         player.ID,
					"Metadata":        map[string]interface{}{"x": "a"},
					"MembershipCount": 1,
				}).(*Clan)
				err = testDb.Insert(clan)
				Expect(err).NotTo(HaveOccurred())

				err = UpdatePlayerOwnershipCount(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.OwnershipCount).To(Equal(1))
			})

			It("Should work if clan is deleted", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				clan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":          player.GameID,
					"PublicID":        uuid.NewV4().String(),
					"OwnerID":         player.ID,
					"Metadata":        map[string]interface{}{"x": "a"},
					"MembershipCount": 1,
				}).(*Clan)
				err = testDb.Insert(clan)
				Expect(err).NotTo(HaveOccurred())

				err = UpdatePlayerOwnershipCount(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())

				clanToBeDeleted := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":          player.GameID,
					"PublicID":        uuid.NewV4().String(),
					"OwnerID":         player.ID,
					"Metadata":        map[string]interface{}{"x": "a"},
					"MembershipCount": 1,
				}).(*Clan)
				err = testDb.Insert(clanToBeDeleted)
				Expect(err).NotTo(HaveOccurred())

				testDb.Delete(clanToBeDeleted)

				err = UpdatePlayerOwnershipCount(testDb, player.ID)
				Expect(err).NotTo(HaveOccurred())
				dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbPlayer.OwnershipCount).To(Equal(1))
			})

			It("Should not work if non-existing Player", func() {
				err := UpdatePlayerOwnershipCount(testDb, -1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Player was not found with id: -1"))
			})
		})
	})

	Describe("Migration script", func() {
		Describe("GetPlayersToEncrypt", func() {
			BeforeEach(func() {
				_, err := testDb.Exec(`delete from memberships;
					delete from clans;
					delete from encrypted_players;
					delete from players`,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should return a slice of players that has no EncryptedPlayer", func() {
				gameID := uuid.NewV4().String()
				game := fixtures.GameFactory.MustCreateWithOption(map[string]interface{}{
					"PublicID": gameID,
				}).(*Game)

				_, player, err := fixtures.CreatePlayerFactory(testDb, gameID)
				Expect(err).NotTo(HaveOccurred())

				encryptedPlayer, err := CreatePlayer(
					testDb,
					logger,
					fixtures.GetEncryptionKey(),
					game.PublicID,
					uuid.NewV4().String(),
					"player-name",
					map[string]interface{}{},
				)
				Expect(err).NotTo(HaveOccurred())

				playersToEncrypt, err := GetPlayersToEncrypt(testDb, fixtures.GetEncryptionKey(), 10)
				Expect(err).NotTo(HaveOccurred())

				Expect(playersToEncrypt[0].Name).To(Equal(player.Name))
				Expect(playersToEncrypt[0].PublicID).To(Equal(player.PublicID))
				Expect(playersToEncrypt[0].ID).To(Equal(player.ID))

				for _, playerToEncrypt := range playersToEncrypt {
					Expect(playerToEncrypt.ID).NotTo(Equal(encryptedPlayer.ID))
				}

			})

			It("Should return the amount of players", func() {
				_, _, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				_, _, err = fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playersToEncrypt, err := GetPlayersToEncrypt(testDb, fixtures.GetEncryptionKey(), 1)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(playersToEncrypt)).To(Equal(1))
			})
		})

		Describe("ApplySecurityChanges", func() {
			It("Should encrypt and update players", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				_, secondPlayer, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				playerName := player.Name
				secondPlayerName := secondPlayer.Name

				players := []*Player{player, secondPlayer}

				err = ApplySecurityChanges(testDb, fixtures.GetEncryptionKey(), players)
				Expect(err).NotTo(HaveOccurred())

				var recoveredPlayer, secondRecoveredPlayer *Player
				err = testDb.SelectOne(&recoveredPlayer, "select * from players where id = $1", player.ID)
				Expect(err).NotTo(HaveOccurred())

				err = testDb.SelectOne(&secondRecoveredPlayer, "select * from players where id = $1", secondPlayer.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(recoveredPlayer.ID).To(Equal(player.ID))
				Expect(playerName).NotTo(Equal(player.Name))

				decryptedPlayerName, err := util.DecryptData(recoveredPlayer.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				Expect(decryptedPlayerName).To(Equal(playerName))

				decryptedPlayerName, err = util.DecryptData(secondRecoveredPlayer.Name, fixtures.GetEncryptionKey())
				Expect(err).NotTo(HaveOccurred())

				Expect(decryptedPlayerName).To(Equal(secondPlayerName))
			})

			It("Should create EncryptedPlayer to each player", func() {
				_, player, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				_, secondPlayer, err := fixtures.CreatePlayerFactory(testDb, "")
				Expect(err).NotTo(HaveOccurred())

				players := []*Player{player, secondPlayer}

				err = ApplySecurityChanges(testDb, fixtures.GetEncryptionKey(), players)
				Expect(err).NotTo(HaveOccurred())

				var encryptedPlayer *EncryptedPlayer
				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", player.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(encryptedPlayer.PlayerID).To(Equal(player.ID))

				err = testDb.SelectOne(&encryptedPlayer, "select * from encrypted_players where player_id = $1", secondPlayer.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(encryptedPlayer.PlayerID).To(Equal(secondPlayer.ID))
			})
		})
	})
})
