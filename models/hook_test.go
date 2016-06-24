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

func TestHookModel(t *testing.T) {
	g := Goblin(t)
	testDb, err := GetTestDB()

	g.Assert(err == nil).IsTrue()

	g.Describe("Hook Model", func() {

		g.Describe("Model Basic Tests", func() {
			g.It("Should create a new Hook", func() {
				hook, err := CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/created")
				g.Assert(err == nil).IsTrue()
				g.Assert(hook.ID != 0).IsTrue()

				dbHook, err := GetHookByID(testDb, hook.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbHook.GameID).Equal(hook.GameID)
				g.Assert(dbHook.URL).Equal(hook.URL)
				g.Assert(dbHook.EventType).Equal(hook.EventType)
			})

			g.It("Should update a Hook", func() {
				hook, err := CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/updated")
				g.Assert(err == nil).IsTrue()
				dt := hook.UpdatedAt
				hook.URL = "http://test/updated2"

				time.Sleep(time.Millisecond)

				count, err := testDb.Update(hook)
				g.Assert(err == nil).IsTrue()
				g.Assert(int(count)).Equal(1)
				g.Assert(hook.UpdatedAt > dt).IsTrue()
				g.Assert(hook.URL).Equal("http://test/updated2")
			})
		})

		g.Describe("Get Hook By ID", func() {
			g.It("Should get existing Hook", func() {
				hook, err := CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/getbyid")
				g.Assert(err == nil).IsTrue()

				dbHook, err := GetHookByID(testDb, hook.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbHook.ID).Equal(hook.ID)
			})

			g.It("Should not get non-existing Hook", func() {
				_, err := GetHookByID(testDb, -1)
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Hook was not found with id: -1")
			})
		})

		g.Describe("Get Hook By Public ID", func() {
			g.It("Should get existing Hook", func() {
				hook, err := CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/getbyid")
				g.Assert(err == nil).IsTrue()

				dbHook, err := GetHookByPublicID(testDb, hook.GameID, hook.PublicID)
				g.Assert(err == nil).IsTrue()
				g.Assert(dbHook.ID).Equal(hook.ID)
			})

			g.It("Should not get non-existing Hook", func() {
				_, err := GetHookByPublicID(testDb, "invalid", "key")
				g.Assert(err != nil).IsTrue()
				g.Assert(err.Error()).Equal("Hook was not found with id: key")
			})
		})

		g.Describe("Create Hook", func() {
			g.It("Should create a new Hook with CreateHook", func() {
				hook, err := CreateHook(
					testDb,
					"create-1",
					GameUpdatedHook,
					"http://test/created",
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(hook.ID != 0).IsTrue()

				dbHook, err := GetHookByID(testDb, hook.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbHook.GameID).Equal(hook.GameID)
				g.Assert(dbHook.EventType).Equal(hook.EventType)
				g.Assert(dbHook.URL).Equal(hook.URL)
			})

			g.It("Create same Hook works fine", func() {
				gameID := uuid.NewV4().String()
				hook, err := CreateHookFactory(testDb, gameID, GameUpdatedHook, "http://test/created")

				hook2, err := CreateHook(
					testDb,
					gameID,
					GameUpdatedHook,
					"http://test/created",
				)
				g.Assert(err == nil).IsTrue()
				g.Assert(hook2.ID == hook.ID).IsTrue()

				dbHook, err := GetHookByID(testDb, hook.ID)
				g.Assert(err == nil).IsTrue()

				g.Assert(dbHook.GameID).Equal(hook.GameID)
				g.Assert(dbHook.EventType).Equal(hook.EventType)
				g.Assert(dbHook.URL).Equal(hook.URL)
			})

		})

		g.Describe("Remove Hook", func() {
			g.It("Should remove a Hook with RemoveHook", func() {
				hook, err := CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/update")
				g.Assert(err == nil).IsTrue()

				err = RemoveHook(
					testDb,
					hook.GameID,
					hook.PublicID,
				)

				g.Assert(err == nil).IsTrue()

				number, err := testDb.SelectInt("select count(*) from hooks where id=$1", hook.ID)
				g.Assert(err == nil).IsTrue()
				g.Assert(number == 0).IsTrue()
			})
		})

		g.Describe("Get All Hooks", func() {
			g.It("Should get all hooks", func() {
				_, err := GetTestHooks(testDb, "", 5)
				g.Assert(err == nil).IsTrue()

				hooks, err := GetAllHooks(testDb)

				g.Assert(err == nil).IsTrue()
				g.Assert(len(hooks) > 10).IsTrue()
			})
		})

	})
}
