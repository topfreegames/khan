// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	. "github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var _ = Describe("Game Model", func() {
	var testDb DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("basic functionality", func() {
		Describe("creating a new game", func() {
			It("should create a game", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())
				Expect(game.ID).NotTo(Equal(0))

				dbGame, err := GetGameByID(testDb, game.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbGame.PublicID).To(Equal(game.PublicID))
				Expect(dbGame.Name).To(Equal(game.Name))
				Expect(dbGame.MinMembershipLevel).To(Equal(game.MembershipLevels["Member"]))
				Expect(dbGame.MaxMembershipLevel).To(Equal(game.MembershipLevels["CoLeader"]))
				Expect(dbGame.MinLevelToAcceptApplication).To(Equal(game.MinLevelToAcceptApplication))
				Expect(dbGame.MinLevelToCreateInvitation).To(Equal(game.MinLevelToCreateInvitation))
				Expect(dbGame.MinLevelToRemoveMember).To(Equal(game.MinLevelToRemoveMember))
				Expect(dbGame.MinLevelOffsetToRemoveMember).To(Equal(game.MinLevelOffsetToRemoveMember))
				Expect(dbGame.MinLevelOffsetToPromoteMember).To(Equal(game.MinLevelOffsetToPromoteMember))
				Expect(dbGame.MinLevelOffsetToDemoteMember).To(Equal(game.MinLevelOffsetToDemoteMember))
				Expect(dbGame.MaxMembers).To(Equal(game.MaxMembers))
				Expect(dbGame.MaxClansPerPlayer).To(Equal(game.MaxClansPerPlayer))
				Expect(dbGame.CooldownAfterDeny).To(Equal(game.CooldownAfterDeny))
				Expect(dbGame.CooldownAfterDelete).To(Equal(game.CooldownAfterDelete))
				for k, v := range dbGame.MembershipLevels {
					Expect(int(v.(float64))).To(Equal(game.MembershipLevels[k].(int)))
				}
				Expect(dbGame.Metadata).To(Equal(game.Metadata))
			})
		})
		Describe("updating new game", func() {
			It("should update game", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				dt := game.UpdatedAt

				time.Sleep(time.Millisecond)

				game.Metadata = map[string]interface{}{"x": "a"}
				count, err := testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(count)).To(Equal(1))
				Expect(game.UpdatedAt).To(BeNumerically(">", dt))
			})
		})
	})

	Describe("getting game by ID", func() {
		It("Should get existing Game", func() {
			game := fixtures.GameFactory.MustCreate().(*Game)
			err := testDb.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			dbGame, err := GetGameByID(testDb, game.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.ID).To(Equal(game.ID))
		})

		It("Should not get non-existing Game", func() {
			_, err := GetGameByID(testDb, -1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Game was not found with id: -1"))
		})
	})

	Describe("getting a game by ID", func() {
		It("Should get existing Game", func() {
			game := fixtures.GameFactory.MustCreate().(*Game)
			err := testDb.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			dbGame, err := GetGameByID(testDb, game.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.ID).To(Equal(game.ID))
		})

		It("Should not get non-existing Game", func() {
			_, err := GetGameByID(testDb, -1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Game was not found with id: -1"))
		})
	})

	Describe("getting a game by Public ID", func() {
		It("Should get existing Game by Game and Game", func() {
			game := fixtures.GameFactory.MustCreate().(*Game)
			err := testDb.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			dbGame, err := GetGameByPublicID(testDb, game.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.ID).To(Equal(game.ID))
		})

		It("Should not get non-existing Game by Game and Game", func() {
			_, err := GetGameByPublicID(testDb, "invalid-game")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Game was not found with id: invalid-game"))
		})
	})

	Describe("creating a new game", func() {
		It("Should create a new Game with CreateGame", func() {
			publicID := uuid.NewV4().String()
			name := "game-name"
			levels := map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3}
			metadata := map[string]interface{}{}
			minLevelToAcceptApplication := 8
			minLevelToCreateInvitation := 7
			minLevelToRemoveMember := 8
			minLevelOffsetToRemoveMember := 1
			minLevelOffsetToPromoteMember := 2
			minLevelOffsetToDemoteMember := 3
			maxMembers := 100
			maxClansPerPlayer := 1
			cooldownAfterDeny := 5
			cooldownAfterDelete := 10
			cooldownBeforeInvite := 8
			cooldownBeforeApply := 25
			maxPendingInvites := 20
			clanUpdateMetadataFieldsHookTriggerWhitelist := "x"
			playerUpdateMetadataFieldsHookTriggerWhitelist := "y,z"

			game, err := CreateGame(
				testDb,
				publicID,
				name,
				levels,
				metadata,
				minLevelToAcceptApplication,
				minLevelToCreateInvitation,
				minLevelToRemoveMember,
				minLevelOffsetToRemoveMember,
				minLevelOffsetToPromoteMember,
				minLevelOffsetToDemoteMember,
				maxMembers,
				maxClansPerPlayer,
				cooldownAfterDeny,
				cooldownAfterDelete,
				cooldownBeforeInvite,
				cooldownBeforeApply,
				maxPendingInvites,
				false,
				clanUpdateMetadataFieldsHookTriggerWhitelist,
				playerUpdateMetadataFieldsHookTriggerWhitelist,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(game.ID).NotTo(Equal(0))

			dbGame, err := GetGameByID(testDb, game.ID)
			Expect(err).NotTo(HaveOccurred())

			Expect(dbGame.PublicID).To(Equal(publicID))
			Expect(dbGame.Name).To(Equal(name))
			Expect(dbGame.MinMembershipLevel).To(Equal(1))
			Expect(dbGame.MaxMembershipLevel).To(Equal(3))
			Expect(dbGame.MinLevelToAcceptApplication).To(Equal(minLevelToAcceptApplication))
			Expect(dbGame.MinLevelToCreateInvitation).To(Equal(minLevelToCreateInvitation))
			Expect(dbGame.MinLevelToRemoveMember).To(Equal(minLevelToRemoveMember))
			Expect(dbGame.MinLevelOffsetToRemoveMember).To(Equal(minLevelOffsetToRemoveMember))
			Expect(dbGame.MinLevelOffsetToPromoteMember).To(Equal(minLevelOffsetToPromoteMember))
			Expect(dbGame.MinLevelOffsetToDemoteMember).To(Equal(minLevelOffsetToDemoteMember))
			Expect(dbGame.MaxMembers).To(Equal(maxMembers))
			Expect(dbGame.MaxClansPerPlayer).To(Equal(maxClansPerPlayer))
			Expect(dbGame.CooldownAfterDelete).To(Equal(cooldownAfterDelete))
			Expect(dbGame.CooldownAfterDeny).To(Equal(cooldownAfterDeny))
			Expect(dbGame.MaxPendingInvites).To(Equal(maxPendingInvites))
			Expect(dbGame.ClanUpdateMetadataFieldsHookTriggerWhitelist).To(Equal("x"))
			Expect(dbGame.PlayerUpdateMetadataFieldsHookTriggerWhitelist).To(Equal("y,z"))

			for k, v := range dbGame.MembershipLevels {
				Expect(v.(float64)).To(BeEquivalentTo(game.MembershipLevels[k]))
			}
			Expect(dbGame.Metadata).To(Equal(game.Metadata))
		})
	})

	Describe("Update Game", func() {
		It("Should update a Game with UpdateGame", func() {
			game := fixtures.GameFactory.MustCreate().(*Game)
			err := testDb.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			updGame, err := UpdateGame(
				testDb,
				game.PublicID,
				"game-new-name",
				map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3},
				map[string]interface{}{"x": "a"},
				5, 4, 7, 1, 1, 1, 100, 1, 5, 15, 8, 25, 20,
				"x", "y,z",
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(updGame.ID).To(Equal(game.ID))

			dbGame, err := GetGameByPublicID(testDb, game.PublicID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.PublicID).To(Equal(updGame.PublicID))
			Expect(dbGame.Name).To(Equal(updGame.Name))
			Expect(dbGame.MinMembershipLevel).To(Equal(updGame.MinMembershipLevel))
			Expect(dbGame.MaxMembershipLevel).To(Equal(updGame.MaxMembershipLevel))
			Expect(dbGame.MinLevelToAcceptApplication).To(Equal(updGame.MinLevelToAcceptApplication))
			Expect(dbGame.MinLevelToCreateInvitation).To(Equal(updGame.MinLevelToCreateInvitation))
			Expect(dbGame.MinLevelToRemoveMember).To(Equal(updGame.MinLevelToRemoveMember))
			Expect(dbGame.MinLevelOffsetToRemoveMember).To(Equal(updGame.MinLevelOffsetToRemoveMember))
			Expect(dbGame.MinLevelOffsetToPromoteMember).To(Equal(updGame.MinLevelOffsetToPromoteMember))
			Expect(dbGame.MinLevelOffsetToDemoteMember).To(Equal(updGame.MinLevelOffsetToDemoteMember))
			Expect(dbGame.MaxMembers).To(Equal(updGame.MaxMembers))
			Expect(dbGame.MaxClansPerPlayer).To(Equal(updGame.MaxClansPerPlayer))
			Expect(dbGame.CooldownAfterDelete).To(Equal(updGame.CooldownAfterDelete))
			Expect(dbGame.CooldownAfterDeny).To(Equal(updGame.CooldownAfterDeny))
			Expect(dbGame.CooldownBeforeInvite).To(Equal(updGame.CooldownBeforeInvite))
			Expect(dbGame.CooldownBeforeApply).To(Equal(updGame.CooldownBeforeApply))
			Expect(dbGame.MaxPendingInvites).To(Equal(updGame.MaxPendingInvites))
			for k, v := range dbGame.MembershipLevels {
				Expect(v.(float64)).To(BeEquivalentTo(updGame.MembershipLevels[k]))
			}
			Expect(dbGame.Metadata).To(Equal(updGame.Metadata))
			Expect(dbGame.ClanUpdateMetadataFieldsHookTriggerWhitelist).To(Equal("x"))
			Expect(dbGame.PlayerUpdateMetadataFieldsHookTriggerWhitelist).To(Equal("y,z"))
		})

		It("Should create a Game with UpdateGame if game does not exist", func() {
			gameID := uuid.NewV4().String()
			updGame, err := UpdateGame(
				testDb,
				gameID,
				gameID,
				map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3},
				map[string]interface{}{"x": "a"},
				5, 4, 7, 1, 1, 1, 100, 1, 10, 30, 8, 25, 20,
				"x", "y,z",
			)

			Expect(err).NotTo(HaveOccurred())

			dbGame, err := GetGameByPublicID(testDb, gameID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbGame.PublicID).To(Equal(updGame.PublicID))
			Expect(dbGame.Name).To(Equal(updGame.Name))
			Expect(dbGame.MinMembershipLevel).To(Equal(updGame.MinMembershipLevel))
			Expect(dbGame.MaxMembershipLevel).To(Equal(updGame.MaxMembershipLevel))
			Expect(dbGame.MinLevelToAcceptApplication).To(Equal(updGame.MinLevelToAcceptApplication))
			Expect(dbGame.MinLevelToCreateInvitation).To(Equal(updGame.MinLevelToCreateInvitation))
			Expect(dbGame.MinLevelToRemoveMember).To(Equal(updGame.MinLevelToRemoveMember))
			Expect(dbGame.MinLevelOffsetToRemoveMember).To(Equal(updGame.MinLevelOffsetToRemoveMember))
			Expect(dbGame.MinLevelOffsetToPromoteMember).To(Equal(updGame.MinLevelOffsetToPromoteMember))
			Expect(dbGame.MinLevelOffsetToDemoteMember).To(Equal(updGame.MinLevelOffsetToDemoteMember))
			Expect(dbGame.MaxMembers).To(Equal(updGame.MaxMembers))
			Expect(dbGame.CooldownAfterDelete).To(Equal(updGame.CooldownAfterDelete))
			Expect(dbGame.CooldownAfterDeny).To(Equal(updGame.CooldownAfterDeny))
			Expect(dbGame.CooldownBeforeInvite).To(Equal(updGame.CooldownBeforeInvite))
			Expect(dbGame.CooldownBeforeApply).To(Equal(updGame.CooldownBeforeApply))
			Expect(dbGame.MaxPendingInvites).To(Equal(updGame.MaxPendingInvites))
			for k, v := range dbGame.MembershipLevels {
				Expect(v.(float64)).To(Equal(updGame.MembershipLevels[k].(float64)))
			}
			Expect(dbGame.Metadata).To(Equal(updGame.Metadata))
			Expect(dbGame.ClanUpdateMetadataFieldsHookTriggerWhitelist).To(Equal("x"))
			Expect(dbGame.PlayerUpdateMetadataFieldsHookTriggerWhitelist).To(Equal("y,z"))
		})

		It("Should not update a Game with Invalid Data with UpdateGame", func() {
			game := fixtures.GameFactory.MustCreate().(*Game)
			err := testDb.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			_, err = UpdateGame(
				testDb,
				game.PublicID,
				strings.Repeat("a", 256),
				map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3},
				map[string]interface{}{"x": "a"},
				5, 4, 7, 1, 1, 0, 100, 1, 0, 0, 8, 25, 20,
				"x", "y,z",
			)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("pq: value too long for type character varying(255)"))
		})
	})

	Describe("Get All Games", func() {
		It("Should get all games", func() {
			game := fixtures.GameFactory.MustCreate().(*Game)
			err := testDb.Insert(game)
			Expect(err).NotTo(HaveOccurred())

			games, err := GetAllGames(
				testDb,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(games)).To(BeNumerically(">", 1))
		})
	})
})
