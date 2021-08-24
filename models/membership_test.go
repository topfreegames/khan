// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	. "github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
	"github.com/topfreegames/khan/util"
)

var _ = Describe("Membership Model", func() {
	var testDb DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Membership Model", func() {
		Describe("Get Number of Pending Invites", func() {
			It("Should get number of pending invites", func() {
				_, player, err := fixtures.GetTestPlayerWithMemberships(testDb, uuid.NewV4().String(), 0, 0, 0, 20)
				Expect(err).NotTo(HaveOccurred())
				Expect(player).NotTo(BeEquivalentTo(nil))

				totalInvites, err := GetNumberOfPendingInvites(testDb, player)
				Expect(err).NotTo(HaveOccurred())
				Expect(totalInvites).To(Equal(20))
			})
		})

		Describe("GetMembershipByID", func() {
			It("Should get a  Membership", func() {
				_, _, _, _, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())

				membership := memberships[0]

				Expect(membership.ID).NotTo(BeEquivalentTo(0))

				dbMembership, err := GetMembershipByID(testDb, membership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(membership.GameID))
				Expect(dbMembership.PlayerID).To(Equal(membership.PlayerID))
				Expect(dbMembership.ClanID).To(Equal(membership.ClanID))
			})

			It("Should get existing Membership", func() {
				_, _, _, _, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMembership.ID).To(Equal(memberships[0].ID))
			})
		})

		Describe("GetOldestMemberWithHighestLevel", func() {
			It("Should get the approved member with the highest level", func() {
				_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 2, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				memberships[0].Level = "CoLeader"
				_, err = testDb.Update(memberships[0])

				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, clan.GameID, clan.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMembership.ID).To(Equal(memberships[0].ID))
				Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
			})

			It("Should not get pending memberships", func() {
				_, clan, _, _, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
				Expect(err).NotTo(HaveOccurred())

				memberships[0].Level = "CoLeader"
				_, err = testDb.Update(memberships[0])

				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, clan.GameID, clan.PublicID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Clan %v has no members", clan.PublicID)))
				Expect(dbMembership).To(BeNil())
			})

			It("Should return an error if clan has no members", func() {
				_, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, clan.GameID, clan.PublicID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Clan %v has no members", clan.PublicID)))
				Expect(dbMembership).To(BeNil())
			})

			It("Should return an error if clan does not exist", func() {
				dbMembership, err := GetOldestMemberWithHighestLevel(testDb, "abc", "def")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Clan def has no members"))
				Expect(dbMembership).To(BeNil())
			})
		})

		Describe("Create Membership", func() {
			It("Should not create a new membership for a member with max number of invitations", func() {
				gameID := uuid.NewV4().String()
				_, player, err := fixtures.GetTestPlayerWithMemberships(testDb, gameID, 0, 0, 0, 20)
				Expect(err).NotTo(HaveOccurred())
				Expect(player).NotTo(BeEquivalentTo(nil))

				game, err := GetGameByPublicID(testDb, gameID)
				Expect(err).NotTo(HaveOccurred())
				Expect(game).NotTo(BeEquivalentTo(nil))

				_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, gameID, "", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(clan).NotTo(BeEquivalentTo(nil))

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game,
					game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)

				Expect(err).To(HaveOccurred())

				expected := fmt.Sprintf(
					"Player %s reached max number of pending invites",
					player.PublicID,
				)
				Expect(err.Error()).To(Equal(expected))
			})

			It("Should not create a new membership for the clan owner", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(clan).NotTo(BeEquivalentTo(nil))

				game.MaxClansPerPlayer = 2
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game,
					game.PublicID,
					"Member",
					owner.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)

				Expect(err).To(HaveOccurred())

				expected := fmt.Sprintf(
					"Player %s already has a valid membership in clan %s.",
					owner.PublicID, clan.PublicID,
				)
				Expect(err.Error()).To(Equal(expected))
			})

			It("Should create a new membership for a member when game MaxPendingInvites is -1", func() {
				gameID := uuid.NewV4().String()
				_, player, err := fixtures.GetTestPlayerWithMemberships(testDb, gameID, 0, 0, 0, 20)
				Expect(err).NotTo(HaveOccurred())
				Expect(player).NotTo(BeEquivalentTo(nil))

				game, err := GetGameByPublicID(testDb, gameID)
				Expect(err).NotTo(HaveOccurred())
				Expect(game).NotTo(BeEquivalentTo(nil))
				game.MaxPendingInvites = -1
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, gameID, "", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(clan).NotTo(BeEquivalentTo(nil))

				membership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game,
					game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(membership.GameID).To(Equal(game.PublicID))
				Expect(membership.PlayerID).To(Equal(player.ID))
				Expect(membership.ClanID).To(Equal(clan.ID))
			})

			It("Should allow users to recreate an invitation if no cooldown", func() {
				game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())
				membership := memberships[0]
				Expect(membership.ID).NotTo(BeEquivalentTo(0))

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					players[0].PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(updMembership.GameID))
				Expect(dbMembership.PlayerID).To(Equal(updMembership.PlayerID))
				Expect(dbMembership.ClanID).To(Equal(updMembership.ClanID))
			})

			It("Should allow users to recreate an application if no cooldown", func() {
				game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.CooldownBeforeApply = 0
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(updMembership.GameID))
				Expect(dbMembership.PlayerID).To(Equal(updMembership.PlayerID))
				Expect(dbMembership.ClanID).To(Equal(updMembership.ClanID))
			})

			It("Should fail if user re-applies before cooldown", func() {
				game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must wait 3600 seconds before creating a membership in clan"))
			})

			It("Should allow users to recreate an invitation if no cooldown", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(updMembership.GameID))
				Expect(dbMembership.PlayerID).To(Equal(updMembership.PlayerID))
				Expect(dbMembership.ClanID).To(Equal(updMembership.ClanID))
			})

			It("Should fail if user to be re-invited before cooldown", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.CooldownBeforeInvite = 1000
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must wait 1000 seconds before creating a membership in clan"))
			})

			It("Should allow users to create an application after an invitation if no cooldown", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.CooldownBeforeInvite = 1000
				game.CooldownBeforeApply = 0
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(updMembership.GameID))
				Expect(dbMembership.PlayerID).To(Equal(updMembership.PlayerID))
				Expect(dbMembership.ClanID).To(Equal(updMembership.ClanID))
			})

			It("Should allow users to create an invitation after an application if no cooldown", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.CooldownBeforeInvite = 0
				game.CooldownBeforeApply = 1000
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(updMembership.GameID))
				Expect(dbMembership.PlayerID).To(Equal(updMembership.PlayerID))
				Expect(dbMembership.ClanID).To(Equal(updMembership.ClanID))
			})

			It("Should not fail if an application after an invitation has cooldown", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.CooldownBeforeInvite = 2000
				game.CooldownBeforeApply = 1000
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				clan.AutoJoin = true
				clan.AllowApplication = true
				_, err = testDb.Update(clan)
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(game.PublicID))
				Expect(dbMembership.PlayerID).To(Equal(player.ID))
				Expect(dbMembership.ClanID).To(Equal(clan.ID))
				Expect(dbMembership.Approved).To(BeTrue())
				Expect(dbMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))
				Expect(dbMembership.ApproverID.Int64).To(BeEquivalentTo(player.ID))
			})

			It("Should not fail if an invitation after an application has cooldown", func() {
				game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				game.CooldownBeforeInvite = 2000
				game.CooldownBeforeApply = 1000
				_, err = testDb.Update(game)
				Expect(err).NotTo(HaveOccurred())

				_, player, err := fixtures.CreatePlayerFactory(testDb, game.PublicID, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					player.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				updMembership, err := CreateMembership(
					testDb,
					fixtures.GetEncryptionKey(),
					game, game.PublicID,
					"Member",
					player.PublicID,
					clan.PublicID,
					owner.PublicID,
					"",
				)
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByID(testDb, updMembership.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbMembership.GameID).To(Equal(game.PublicID))
				Expect(dbMembership.PlayerID).To(Equal(player.ID))
				Expect(dbMembership.ClanID).To(Equal(clan.ID))
				Expect(dbMembership.Approved).To(BeFalse())
				Expect(dbMembership.ApproverID.Valid).To(BeFalse())
			})

			Describe("Should create a new Membership with CreateMembership", func() {
				It("If requestor is the player and clan.AllowApplication = true", func() {
					game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					clan.AllowApplication = true
					_, err = testDb.Update(clan)
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						player.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(player.ID))
					Expect(dbMembership.RequestorID).To(Equal(player.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If requestor is the player and clan.AllowApplication = true and previous deleted membership", func() {
					game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownBeforeApply = 0
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					clan.AllowApplication = true
					_, err = testDb.Update(clan)
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
					)

					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
						"Please accept me",
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(players[0].ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.DeletedAt).To(Equal(int64(0)))
					Expect(dbMembership.Message).To(Equal("Please accept me"))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), players[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If requestor is the player and clan.AllowApplication = true and autoJoin = true and previous deleted membership", func() {
					game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownBeforeApply = 0
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					clan.AllowApplication = true
					clan.AutoJoin = true
					_, err = testDb.Update(clan)
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
					)

					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
						"Please accept me",
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(players[0].ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(true))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.DeletedAt).To(Equal(int64(0)))
					Expect(dbMembership.Message).To(Equal("Please accept me"))
					Expect(dbMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), players[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(2))
				})

				It("Should approve it automatically if requestor is the player, clan.AllowApplication=true and clan.AutoJoin=true", func() {
					game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					clan.AllowApplication = true
					clan.AutoJoin = true
					_, err = testDb.Update(clan)
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						player.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(player.ID))
					Expect(dbMembership.RequestorID).To(Equal(player.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(true))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))
					Expect(dbMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), player.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(2))
				})

				It("If requestor is the clan owner", func() {
					game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(player.ID))
					Expect(dbMembership.RequestorID).To(Equal(clan.OwnerID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))
				})

				It("If requestor is a member of the clan with level greater than the min level", func() {
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Level = "CoLeader"
					memberships[0].Approved = true
					memberships[0].Denied = false
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						players[0].PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(player.ID))
					Expect(dbMembership.RequestorID).To(Equal(players[0].ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))
				})

				It("If deleted previous membership", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = memberships[0].PlayerID
					memberships[0].Approved = false
					memberships[0].Denied = false
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(owner.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), dbMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If deleted previous membership, after waiting cooldown seconds", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDelete = 1
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = memberships[0].PlayerID
					memberships[0].Approved = false
					memberships[0].Denied = false
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					time.Sleep(time.Second)
					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(owner.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), dbMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If deleted previous membership and deleter is not the membership player and requestor is not membership player before waiting cooldown seconds", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDelete = 10
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = owner.ID
					memberships[0].Approved = false
					memberships[0].Denied = false
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(owner.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))
					Expect(dbMembership.DeletedAt).To(Equal(int64(0)))
				})

				It("If denied previous membership and denier is the membership player, before waiting cooldown seconds", func() {
					game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDeny = 10
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
				})

				It("If denied previous membership, after waiting cooldown seconds", func() {
					game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDeny = 1
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					time.Sleep(time.Second)
					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(owner.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), dbMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If deleted previous membership and deleter is the membership player, before waiting cooldown seconds", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDelete = 10
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = memberships[0].PlayerID
					memberships[0].Approved = false
					memberships[0].Denied = false
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.ID).NotTo(BeEquivalentTo(0))

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.GameID).To(Equal(membership.GameID))
					Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
					Expect(dbMembership.RequestorID).To(Equal(owner.ID))
					Expect(dbMembership.ClanID).To(Equal(clan.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.Message).To(Equal("Please accept me"))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), dbMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If denied previous membership request and clan autoJoin is changed to true, before waiting cooldown seconds", func() {
					// create a pending membership application
					game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "", false, false)
					Expect(err).NotTo(HaveOccurred())

					// deny application
					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"deny",
					)
					Expect(err).NotTo(HaveOccurred())

					game.CooldownBeforeApply = 0
					game.CooldownAfterDeny = 100000
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					// change clan autoJoin to true
					clan.AutoJoin = true
					_, err = testDb.Update(clan)
					Expect(err).NotTo(HaveOccurred())

					// apply for membership again
					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
						"Please accept me",
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(membership.Approved).To(BeTrue())

					dbMembership, err := GetMembershipByID(testDb, membership.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.Approved).To(BeTrue())
					Expect(dbMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))
					Expect(dbMembership.Denied).To(Equal(false))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), dbMembership.PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(2))
				})
			})

			Describe("Should not create a new Membership with CreateMembership if", func() {
				It("If deleted previous membership and deleter is not the membership player and requestor is the membership player before waiting cooldown seconds", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDelete = 10
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = owner.ID
					memberships[0].Approved = false
					memberships[0].Denied = false
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s must wait 10 seconds before creating a membership in clan %s.", players[0].PublicID, clan.PublicID)))
				})

				It("If denied previous membership and denier is not the membership player, before waiting cooldown seconds", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					game.CooldownAfterDeny = 10
					_, err = testDb.Update(game)
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DenierID.Int64 = int64(owner.ID)
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s must wait 10 seconds before creating a membership in clan %s.", players[0].PublicID, clan.PublicID)))
				})

				It("If clan reached the game's MaxMembers", func() {
					game, clan, owner, _, _, err := fixtures.GetClanReachedMaxMemberships(testDb)
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Clan %s reached max members", clan.PublicID)))
				})

				It("If player reached the game's MaxClansPerPlayer (member of another clans)", func() {
					game, _, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					anotherClan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":   owner.GameID,
						"PublicID": uuid.NewV4().String(),
						"OwnerID":  owner.ID,
						"Metadata": map[string]interface{}{"x": "a"},
					}).(*Clan)
					err = testDb.Insert(anotherClan)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						game.PublicID,
						"Member",
						players[0].PublicID,
						anotherClan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID)))
				})

				It("If player reached the game's MaxClansPerPlayer (owner of another clan)", func() {
					game, _, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					anotherClan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":   players[0].GameID,
						"PublicID": uuid.NewV4().String(),
						"OwnerID":  players[0].ID,
						"Metadata": map[string]interface{}{"x": "a"},
					}).(*Clan)
					err = testDb.Insert(anotherClan)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						game.PublicID,
						"Member",
						players[0].PublicID,
						anotherClan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID)))
				})

				It("If requestor is the player and clan.AllowApplication = false", func() {
					game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					clan.AllowApplication = false
					_, err = testDb.Update(clan)
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						player.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", player.PublicID, clan.PublicID)))

				})

				It("Unexistent player", func() {
					game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					playerPublicID := randomdata.FullName(randomdata.RandomGender)

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						owner.GameID,
						"Member",
						playerPublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player was not found with id: %s", playerPublicID)))
				})

				It("Unexistent clan", func() {
					game, player, err := fixtures.CreatePlayerFactory(testDb, "")
					Expect(err).NotTo(HaveOccurred())

					clanPublicID := randomdata.FullName(randomdata.RandomGender)
					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clanPublicID,
						player.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Clan was not found with id: %s", clanPublicID)))

				})

				It("Unexistent requestor", func() {
					game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					requestorPublicID := randomdata.FullName(randomdata.RandomGender)
					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						requestorPublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player was not found with id: %s", requestorPublicID)))
				})

				It("Requestor is not clan member/owner", func() {
					game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					requestor := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(requestor)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						requestor.PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot create membership for clan %s",
						requestor.PublicID, clan.PublicID)))
				})

				It("Requestor's level is less than min level", func() {
					game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						"Member",
						player.PublicID,
						clan.PublicID,
						players[0].PublicID,
						"Please accept me",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot create membership for clan %s", players[0].PublicID, clan.PublicID)))
				})

				It("Membership already exists", func() {
					game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					membership, err := CreateMembership(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						"Member",
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"Please accept me",
					)

					Expect(membership).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s already has a valid membership in clan %s.", players[0].PublicID, clan.PublicID)))
				})
			})
		})

		Describe("GetDeletedMembershipByClanAndPlayerID", func() {
			It("Should get a deleted Membership by the clan ID and player private ID using GetDeletedMembershipByClanAndPlayerID", func() {
				_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = players[0].ID
				_, err = testDb.Update(memberships[0])
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetDeletedMembershipByClanAndPlayerID(testDb, clan.GameID, clan.ID, players[0].ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMembership.ID).To(Equal(memberships[0].ID))
				Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
				Expect(dbMembership.DeletedBy).To(Equal(players[0].ID))
			})

		})

		Describe("GetValidMembershipByClanAndPlayerPublicID", func() {
			It("Should get an existing Membership by the player public ID", func() {
				_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, players[0].GameID, clan.PublicID, players[0].PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMembership.ID).To(Equal(memberships[0].ID))
				Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
			})

			Describe("Should not get Membership by the player public ID", func() {
				It("If non-existing Membership", func() {
					_, player, err := fixtures.CreatePlayerFactory(testDb, "")
					Expect(err).NotTo(HaveOccurred())

					clan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":  player.GameID,
						"OwnerID": player.ID,
					}).(*Clan)
					err = testDb.Insert(clan)
					Expect(err).NotTo(HaveOccurred())

					dbMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, player.GameID, clan.PublicID, player.PublicID)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID)))
					Expect(dbMembership).To(BeNil())
				})

				It("If Membership was deleted", func() {
					_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = players[0].ID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					dbMembership, err := GetValidMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, players[0].PublicID)
					Expect(err).To(HaveOccurred())
					Expect(dbMembership).To(BeNil())
				})
			})

			It("Should get a deleted Membership by the clan and player public ID using GetMembershipByClanAndPlayerPublicID", func() {
				_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
				Expect(err).NotTo(HaveOccurred())

				memberships[0].DeletedAt = util.NowMilli()
				memberships[0].DeletedBy = players[0].ID
				_, err = testDb.Update(memberships[0])
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, players[0].PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMembership.ID).To(Equal(memberships[0].ID))
				Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
				Expect(dbMembership.DeletedBy).To(Equal(players[0].ID))
			})

			It("Should get a denied Membership by the player public ID using GetMembershipByClanAndPlayerPublicID", func() {
				_, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 1, 0, 0, "", "")
				Expect(err).NotTo(HaveOccurred())

				dbMembership, err := GetMembershipByClanAndPlayerPublicID(testDb, clan.GameID, clan.PublicID, players[0].PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMembership.ID).To(Equal(memberships[0].ID))
				Expect(dbMembership.PlayerID).To(Equal(players[0].ID))
			})
		})

		Describe("ApproveOrDenyMembershipInvitation", func() {
			Describe("Should approve a Membership invitation with ApproveOrDenyMembershipInvitation if", func() {
				It("Player is not the membership requestor", func() {
					action := "approve"
					game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					updatedMembership, err := ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.PlayerID).To(Equal(players[0].ID))
					Expect(updatedMembership.Approved).To(Equal(true))
					Expect(updatedMembership.Denied).To(Equal(false))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.Approved).To(Equal(true))
					Expect(dbMembership.Denied).To(Equal(false))

					Expect(dbMembership.ApproverID.Valid).To(BeTrue())
					Expect(dbMembership.ApproverID.Int64).To(Equal(int64(players[0].ID)))
					Expect(dbMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), players[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(2))
				})
			})

			Describe("Should not approve a Membership invitation with ApproveOrDenyMembershipInvitation if", func() {
				It("If clan reached the game's MaxMembers", func() {
					action := "approve"
					game, clan, _, players, _, err := fixtures.GetClanReachedMaxMemberships(testDb)
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[1].GameID,
						players[1].PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Clan %s reached max members", clan.PublicID)))
				})

				It("If player reached the game's MaxClansPerPlayer", func() {
					action := "approve"
					game, _, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					anotherClan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":   owner.GameID,
						"PublicID": uuid.NewV4().String(),
						"OwnerID":  owner.ID,
						"Metadata": map[string]interface{}{"x": "a"},
					}).(*Clan)
					err = testDb.Insert(anotherClan)
					Expect(err).NotTo(HaveOccurred())

					membership := fixtures.MembershipFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":      game.PublicID,
						"PlayerID":    players[0].ID,
						"ClanID":      anotherClan.ID,
						"RequestorID": owner.ID,
						"Metadata":    map[string]interface{}{"x": "a"},
						"Level":       "Member",
					}).(*Membership)
					err = testDb.Insert(membership)
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						anotherClan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID)))
				})

				It("Player is the membership requestor", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = players[0].ID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[0].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Membership does not exist", func() {
					action := "approve"
					game, clan, _, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						player.GameID,
						player.PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID)))
				})

				It("Membership is deleted", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].DeletedAt = util.NowMilli()
					memberships[0].DeletedBy = players[0].ID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", players[0].PublicID)))
				})

				It("Membership is already approved", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action)))
				})

				It("Membership is already denied", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Denied = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipInvitation(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action)))
				})

				Describe("Should deny a Membership invitation with ApproveOrDenyMembershipInvitation if", func() {
					It("Player is not the membership requestor", func() {
						action := "deny"
						game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
						Expect(err).NotTo(HaveOccurred())

						updatedMembership, err := ApproveOrDenyMembershipInvitation(
							testDb,
							fixtures.GetEncryptionKey(),
							game,
							players[0].GameID,
							players[0].PublicID,
							clan.PublicID,
							action,
						)

						Expect(err).NotTo(HaveOccurred())
						Expect(updatedMembership.PlayerID).To(Equal(players[0].ID))
						Expect(updatedMembership.Approved).To(Equal(false))
						Expect(updatedMembership.Denied).To(Equal(true))

						dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
						Expect(err).NotTo(HaveOccurred())
						Expect(dbMembership.Approved).To(Equal(false))
						Expect(dbMembership.Denied).To(Equal(true))
						Expect(dbMembership.DenierID.Valid).To(BeTrue())
						Expect(dbMembership.DenierID.Int64).To(Equal(int64(players[0].ID)))
						Expect(dbMembership.DeniedAt).To(BeNumerically(">", util.NowMilli()-1000))
					})
				})

				Describe("Should not ApproveOrDenyMembershipInvitation if", func() {
					It("Invalid action", func() {
						game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
						Expect(err).NotTo(HaveOccurred())

						_, err = ApproveOrDenyMembershipInvitation(
							testDb,
							fixtures.GetEncryptionKey(),
							game,
							players[0].GameID,
							players[0].PublicID,
							clan.PublicID,
							"invalid-action",
						)

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("invalid-action a membership is not a valid action."))
					})
				})
			})
		})

		Describe("ApproveOrDenyMembershipApplication", func() {
			Describe("Should approve a Membership application with ApproveOrDenyMembershipApplication if", func() {
				It("Owner", func() {
					action := "approve"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])

					updatedMembership, err := ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.Approved).To(Equal(true))
					Expect(updatedMembership.Denied).To(Equal(false))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.Approved).To(Equal(true))
					Expect(dbMembership.Denied).To(Equal(false))

					Expect(dbMembership.ApproverID.Valid).To(BeTrue())
					Expect(dbMembership.ApproverID.Int64).To(Equal(int64(owner.ID)))

					Expect(dbMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), players[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(2))
				})

				It("Requestor is member of the clan with level > minLevel", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Level = "CoLeader"
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					memberships[1].RequestorID = memberships[1].PlayerID
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					updatedMembership, err := ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[1].GameID,
						players[1].PublicID,
						clan.PublicID,
						players[0].PublicID,
						action,
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.ID).To(Equal(memberships[1].ID))
					Expect(updatedMembership.Approved).To(Equal(true))
					Expect(updatedMembership.Denied).To(Equal(false))

					Expect(updatedMembership.ApproverID.Valid).To(BeTrue())
					Expect(updatedMembership.ApproverID.Int64).To(Equal(int64(players[0].ID)))

					Expect(updatedMembership.ApprovedAt).To(BeNumerically("~", util.NowMilli(), 1000))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.Approved).To(Equal(true))
					Expect(dbMembership.Denied).To(Equal(false))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), players[1].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(1))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(3))
				})
			})

			Describe("Should not approve a Membership application with ApproveOrDenyMembershipApplication if", func() {
				It("If clan reached the game's MaxMembers", func() {
					action := "approve"
					game, clan, owner, players, memberships, err := fixtures.GetClanReachedMaxMemberships(testDb)
					Expect(err).NotTo(HaveOccurred())

					memberships[1].RequestorID = memberships[1].PlayerID
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[1].GameID,
						players[1].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Clan %s reached max members", clan.PublicID)))
				})

				It("If player reached the game's MaxClansPerPlayer", func() {
					action := "approve"
					game, _, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					anotherClan := fixtures.ClanFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":   owner.GameID,
						"PublicID": uuid.NewV4().String(),
						"OwnerID":  owner.ID,
						"Metadata": map[string]interface{}{"x": "a"},
					}).(*Clan)
					err = testDb.Insert(anotherClan)
					Expect(err).NotTo(HaveOccurred())

					membership := fixtures.MembershipFactory.MustCreateWithOption(map[string]interface{}{
						"GameID":      game.PublicID,
						"PlayerID":    players[0].ID,
						"ClanID":      anotherClan.ID,
						"RequestorID": players[0].ID,
						"Metadata":    map[string]interface{}{"x": "a"},
						"Level":       "Member",
					}).(*Membership)
					err = testDb.Insert(membership)
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						anotherClan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s reached max clans", players[0].PublicID)))
				})

				It("Requestor is member of the clan with level < minLevel", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "Member"
					memberships[1].Approved = true
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Requestor is not approved member of the clan", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "CoLeader"
					memberships[1].Approved = false
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Requestor is not member of the clan", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					requestor := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(requestor)
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						requestor.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Requestor membership is deleted", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					memberships[1].DeletedAt = util.NowMilli()
					memberships[1].DeletedBy = players[1].ID
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Requestor is the player of the membership", func() {
					action := "approve"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[0].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Player was not the membership requestor", func() {
					action := "approve"
					game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", owner.PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Membership does not exist", func() {
					action := "approve"
					game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						player.PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID)))
				})

				It("Membership is already approved", func() {
					action := "approve"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action)))
				})

				It("Membership is already denied", func() {
					action := "approve"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					memberships[0].Denied = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s membership that was already approved or denied", action)))
				})
			})

			Describe("Should deny a Membership application with ApproveOrDenyMembershipApplication if", func() {
				It("Owner", func() {
					action := "deny"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					updatedMembership, err := ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.Approved).To(Equal(false))
					Expect(updatedMembership.Denied).To(Equal(true))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())

					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(true))
					Expect(dbMembership.DenierID.Valid).To(BeTrue())
					Expect(dbMembership.DenierID.Int64).To(Equal(int64(owner.ID)))
					Expect(dbMembership.DeniedAt).To(BeNumerically(">", util.NowMilli()-1000))

				})
			})

			Describe("Should not ApproveOrDenyMembershipApplication if", func() {
				It("Invalid action", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].RequestorID = memberships[0].PlayerID
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = ApproveOrDenyMembershipApplication(
						testDb,
						fixtures.GetEncryptionKey(),
						game,
						players[0].GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"invalid-action",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("invalid-action a membership is not a valid action."))
				})
			})
		})

		Describe("PromoteOrDemoteMember", func() {
			Describe("Should promote a member with PromoteOrDemoteMember", func() {
				It("If requestor is the owner", func() {
					action := "promote"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					updatedMembership, err := PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.ID).To(Equal(memberships[0].ID))
					Expect(updatedMembership.Level).To(Equal("Elder"))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.Level).To(Equal("Elder"))
				})

				It("If requestor has enough level", func() {
					action := "promote"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "CoLeader"
					memberships[1].Approved = true
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					updatedMembership, err := PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.ID).To(Equal(memberships[0].ID))
					Expect(updatedMembership.Level).To(Equal("Elder"))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.Level).To(Equal("Elder"))
				})
			})

			Describe("Should not promote a member with PromoteOrDemoteMember", func() {
				It("If requestor is the player", func() {
					action := "promote"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[0].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("If requestor does not have enough level", func() {
					action := "promote"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "Member"
					memberships[1].Approved = true
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("If requestor is not a clan member", func() {
					action := "promote"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					requestorPublicID := randomdata.FullName(randomdata.RandomGender)
					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						requestorPublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestorPublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Requestor membership is deleted", func() {
					action := "promote"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])

					memberships[1].DeletedAt = util.NowMilli()
					memberships[1].DeletedBy = clan.OwnerID
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("If player is not a clan member", func() {
					action := "promote"
					game, clan, owner, _, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					player := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(player)
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						player.PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Membership was not found with id: %s", player.PublicID)))
				})

				It("If player membership is not approved", func() {
					action := "promote"
					game, clan, owner, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s membership that is denied or not yet approved", action)))
				})

				It("If player membership is denied", func() {
					action := "promote"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Denied = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s membership that is denied or not yet approved", action)))
				})

				It("If requestor membership is not approved", func() {
					action := "promote"
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, action, players[0].PublicID, clan.PublicID)))
				})

				It("Player is already max level", func() {
					action := "promote"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					memberships[0].Level = "CoLeader"
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s member that is already level %d", action, 3)))
				})
			})

			Describe("Should demote a member with PromoteOrDemoteMember", func() {
				It("If requestor is the owner", func() {
					action := "demote"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Level = "CoLeader"
					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])

					updatedMembership, err := PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).NotTo(HaveOccurred())
					Expect(updatedMembership.ID).To(Equal(memberships[0].ID))
					Expect(updatedMembership.Level).To(Equal("Elder"))

					dbMembership, err := GetMembershipByID(testDb, updatedMembership.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.Level).To(Equal("Elder"))
				})
			})

			Describe("Should not demote a member with PromoteOrDemoteMember if", func() {
				It("Player is already min level", func() {
					action := "demote"
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					memberships[0].Level = "Member"
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						action,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Cannot %s member that is already level %d", action, 1)))
				})
			})

			Describe("Should not PromoteOrDemoteMember", func() {
				It("Invalid action", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[0].Approved = true
					_, err = testDb.Update(memberships[0])
					Expect(err).NotTo(HaveOccurred())

					_, err = PromoteOrDemoteMember(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
						"invalid-action",
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("invalid-action a membership is not a valid action."))
				})
			})
		})

		Describe("DeleteMembership", func() {
			Describe("Should delete a membership with DeleteMembership", func() {
				It("If requestor is the owner", func() {
					game, clan, owner, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						owner.PublicID,
					)

					Expect(err).NotTo(HaveOccurred())

					dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.DeletedBy).To(Equal(owner.ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.DeletedAt).To(BeNumerically(">", util.NowMilli()-1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), memberships[0].PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If requestor is the player", func() {
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 1, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[0].PublicID,
					)

					Expect(err).NotTo(HaveOccurred())

					dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.DeletedBy).To(Equal(players[0].ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.DeletedAt).To(BeNumerically(">", util.NowMilli()-1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), memberships[0].PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(1))
				})

				It("If requestor has enough level and offset", func() {
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 2, 0, 0, 0, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "CoLeader"
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
					)
					Expect(err).NotTo(HaveOccurred())

					dbMembership, err := GetMembershipByID(testDb, memberships[0].ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMembership.DeletedBy).To(Equal(players[1].ID))
					Expect(dbMembership.Approved).To(Equal(false))
					Expect(dbMembership.Denied).To(Equal(false))
					Expect(dbMembership.DeletedAt).To(BeNumerically(">", util.NowMilli()-1000))

					dbPlayer, err := GetPlayerByID(testDb, fixtures.GetEncryptionKey(), memberships[0].PlayerID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbPlayer.MembershipCount).To(Equal(0))

					dbClan, err := GetClanByID(testDb, clan.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbClan.MembershipCount).To(Equal(2))
				})
			})

			Describe("Should not delete a membership with DeleteMembership", func() {
				It("If requestor does not have enough level", func() {
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "Member"
					memberships[1].Approved = true
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID)))
				})

				It("If requestor has enough level but not enough offset", func() {
					game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID)))
				})

				It("If requestor is not a clan member", func() {
					game, clan, _, players, _, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 1, "", "")
					Expect(err).NotTo(HaveOccurred())

					requestor := fixtures.PlayerFactory.MustCreateWithOption(map[string]interface{}{
						"GameID": clan.GameID,
					}).(*Player)
					err = testDb.Insert(requestor)
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						requestor.PublicID,
					)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", requestor.PublicID, "delete", players[0].PublicID, clan.PublicID)))
				})

				It("If requestor membership is denied", func() {
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "CoLeader"
					memberships[1].Denied = true
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID)))
				})

				It("If requestor membership is not approved", func() {
					game, clan, _, players, memberships, err := fixtures.GetClanWithMemberships(testDb, 0, 0, 0, 2, "", "")
					Expect(err).NotTo(HaveOccurred())

					memberships[1].Level = "CoLeader"
					memberships[1].Approved = false
					_, err = testDb.Update(memberships[1])
					Expect(err).NotTo(HaveOccurred())

					_, err = DeleteMembership(
						testDb,
						game,
						clan.GameID,
						players[0].PublicID,
						clan.PublicID,
						players[1].PublicID,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("Player %s cannot %s membership for player %s and clan %s", players[1].PublicID, "delete", players[0].PublicID, clan.PublicID)))
				})
			})
		})
	})
})
