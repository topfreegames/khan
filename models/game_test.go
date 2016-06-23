// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/satori/go.uuid"
)

func TestGameModel(t *testing.T) {
	t.Parallel()
	g := Goblin(t)
	testDb, err := GetTestDB()
	g.Assert(err == nil).IsTrue()

	g.Describe("Game Model", func() {

		g.Describe("Model Basic Tests", func() {
			g.It("Should create a new Game", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()
				g.Assert(game.ID != 0).IsTrue()

				dbGame, err := GetGameByID(testDb, game.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbGame.PublicID).Equal(game.PublicID)
				g.Assert(dbGame.Name).Equal(game.Name)
				g.Assert(dbGame.MinMembershipLevel).Equal(game.MinMembershipLevel)
				g.Assert(dbGame.MaxMembershipLevel).Equal(game.MaxMembershipLevel)
				g.Assert(dbGame.MinLevelToAcceptApplication).Equal(game.MinLevelToAcceptApplication)
				g.Assert(dbGame.MinLevelToCreateInvitation).Equal(game.MinLevelToCreateInvitation)
				g.Assert(dbGame.MinLevelToRemoveMember).Equal(game.MinLevelToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToRemoveMember).Equal(game.MinLevelOffsetToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(game.MinLevelOffsetToPromoteMember)
				g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(game.MinLevelOffsetToDemoteMember)
				g.Assert(dbGame.MaxMembers).Equal(game.MaxMembers)
				// g.Assert(dbGame.MembershipLevels).Equal(game.MembershipLevels)
				g.Assert(dbGame.Metadata).Equal(game.Metadata)
			})

			g.It("Should update a new Game", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()
				dt := game.UpdatedAt

				time.Sleep(time.Millisecond)

				game.Metadata = "{ \"x\": 1 }"
				count, err := testDb.Update(game)
				g.Assert(err == nil).IsTrue()
				g.Assert(int(count)).Equal(1)
				g.Assert(game.UpdatedAt > dt).IsTrue()
			})
		})

		g.Describe("Get Game By ID", func() {
			g.It("Should get existing Game", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()

				dbGame, err := GetGameByID(testDb, game.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbGame.ID).Equal(game.ID)
			})

			g.It("Should not get non-existing Game", func() {
				_, err := GetGameByID(testDb, -1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Game was not found with id: -1")
			})
		})

		g.Describe("Get Game By Public ID", func() {
			g.It("Should get existing Game by Game and Game", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()

				dbGame, err := GetGameByPublicID(testDb, game.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbGame.ID).Equal(game.ID)
			})

			g.It("Should not get non-existing Game by Game and Game", func() {
				_, err := GetGameByPublicID(testDb, "invalid-game")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Game was not found with id: invalid-game")
			})
		})

		g.Describe("Create Game", func() {
			g.It("Should create a new Game with CreateGame", func() {
				game, err := CreateGame(
					testDb,
					"create-1",
					"game-name",
					"{\"Member\": 1, \"Elder\": 2, \"CoLeader\": 3}",
					"{}",
					5, 10, 8, 7, 8, 1, 2, 3, 100,
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(game.ID != 0).IsTrue()

				dbGame, err := GetGameByID(testDb, game.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbGame.PublicID).Equal(game.PublicID)
				g.Assert(dbGame.Name).Equal(game.Name)
				g.Assert(dbGame.MinMembershipLevel).Equal(game.MinMembershipLevel)
				g.Assert(dbGame.MaxMembershipLevel).Equal(game.MaxMembershipLevel)
				g.Assert(dbGame.MinLevelToAcceptApplication).Equal(game.MinLevelToAcceptApplication)
				g.Assert(dbGame.MinLevelToCreateInvitation).Equal(game.MinLevelToCreateInvitation)
				g.Assert(dbGame.MinLevelToRemoveMember).Equal(game.MinLevelToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToRemoveMember).Equal(game.MinLevelOffsetToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(game.MinLevelOffsetToPromoteMember)
				g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(game.MinLevelOffsetToDemoteMember)
				g.Assert(dbGame.MaxMembers).Equal(game.MaxMembers)
				// g.Assert(dbGame.MembershipLevels).Equal(game.MembershipLevels)
				g.Assert(dbGame.Metadata).Equal(game.Metadata)
			})
		})

		g.Describe("Update Game", func() {
			g.It("Should update a Game with UpdateGame", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()

				updGame, err := UpdateGame(
					testDb,
					game.PublicID,
					"game-new-name",
					"{\"Member\": 1, \"Elder\": 2, \"CoLeader\": 3}",
					"{\"x\": 1}",
					2, 12, 5, 4, 7, 1, 1, 1, 100,
				)

				g.Assert(err == nil).IsTrue()
				g.Assert(updGame.ID).Equal(game.ID)

				dbGame, err := GetGameByPublicID(testDb, game.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbGame.PublicID).Equal(updGame.PublicID)
				g.Assert(dbGame.Name).Equal(updGame.Name)
				g.Assert(dbGame.MinMembershipLevel).Equal(updGame.MinMembershipLevel)
				g.Assert(dbGame.MaxMembershipLevel).Equal(updGame.MaxMembershipLevel)
				g.Assert(dbGame.MinLevelToAcceptApplication).Equal(updGame.MinLevelToAcceptApplication)
				g.Assert(dbGame.MinLevelToCreateInvitation).Equal(updGame.MinLevelToCreateInvitation)
				g.Assert(dbGame.MinLevelToRemoveMember).Equal(updGame.MinLevelToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToRemoveMember).Equal(updGame.MinLevelOffsetToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(updGame.MinLevelOffsetToPromoteMember)
				g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(updGame.MinLevelOffsetToDemoteMember)
				g.Assert(dbGame.MaxMembers).Equal(updGame.MaxMembers)
				// g.Assert(dbGame.MembershipLevels).Equal(updGame.MembershipLevels)
				g.Assert(dbGame.Metadata).Equal(updGame.Metadata)
			})

			g.It("Should create a Game with UpdateGame if game does not exist", func() {
				gameID := uuid.NewV4().String()
				updGame, err := UpdateGame(
					testDb,
					gameID,
					gameID,
					"{\"Member\": 1, \"Elder\": 2, \"CoLeader\": 3}",
					"{\"x\": 1}",
					2, 12, 5, 4, 7, 1, 1, 1, 100,
				)

				g.Assert(err == nil).IsTrue()

				dbGame, err := GetGameByPublicID(testDb, gameID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbGame.PublicID).Equal(updGame.PublicID)
				g.Assert(dbGame.Name).Equal(updGame.Name)
				g.Assert(dbGame.MinMembershipLevel).Equal(updGame.MinMembershipLevel)
				g.Assert(dbGame.MaxMembershipLevel).Equal(updGame.MaxMembershipLevel)
				g.Assert(dbGame.MinLevelToAcceptApplication).Equal(updGame.MinLevelToAcceptApplication)
				g.Assert(dbGame.MinLevelToCreateInvitation).Equal(updGame.MinLevelToCreateInvitation)
				g.Assert(dbGame.MinLevelToRemoveMember).Equal(updGame.MinLevelToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToRemoveMember).Equal(updGame.MinLevelOffsetToRemoveMember)
				g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(updGame.MinLevelOffsetToPromoteMember)
				g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(updGame.MinLevelOffsetToDemoteMember)
				g.Assert(dbGame.MaxMembers).Equal(updGame.MaxMembers)
				// g.Assert(dbGame.MembershipLevels).Equal(updGame.MembershipLevels)
				g.Assert(dbGame.Metadata).Equal(updGame.Metadata)
			})

			g.It("Should not update a Game with Invalid Data with UpdateGame", func() {
				game := GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				g.Assert(err == nil).IsTrue()

				_, err = UpdateGame(
					testDb,
					game.PublicID,
					"game-new-name",
					"{\"Member\": 1, \"Elder\": 2, \"CoLeader\": 3}",
					"it-will-fail-beacause-metada-is-not-a-json",
					2, 12, 5, 4, 7, 1, 1, 0, 100,
				)

				g.Assert(err == nil).IsFalse()
				g.Assert(err.Error()).Equal("pq: invalid input syntax for type json")
			})
		})
	})
}
