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
	uuid "github.com/satori/go.uuid"
	. "github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var _ = Describe("Hook Model", func() {
	var testDb DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Hook Model", func() {

		Describe("Model Basic Tests", func() {
			It("Should create a new Hook", func() {
				hook, err := fixtures.CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/created")
				Expect(err).NotTo(HaveOccurred())
				Expect(hook.ID).NotTo(BeEquivalentTo(0))

				dbHook, err := GetHookByID(testDb, hook.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbHook.GameID).To(Equal(hook.GameID))
				Expect(dbHook.URL).To(Equal(hook.URL))
				Expect(dbHook.EventType).To(Equal(hook.EventType))
			})

			It("Should update a Hook", func() {
				hook, err := fixtures.CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/updated")
				Expect(err).NotTo(HaveOccurred())
				dt := hook.UpdatedAt
				hook.URL = "http://test/updated2"

				time.Sleep(time.Millisecond)

				count, err := testDb.Update(hook)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(BeEquivalentTo(1))
				Expect(hook.UpdatedAt).To(BeNumerically(">", dt))
				Expect(hook.URL).To(Equal("http://test/updated2"))
			})
		})

		Describe("Get Hook By ID", func() {
			It("Should get existing Hook", func() {
				hook, err := fixtures.CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/getbyid")
				Expect(err).NotTo(HaveOccurred())

				dbHook, err := GetHookByID(testDb, hook.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbHook.ID).To(Equal(hook.ID))
			})

			It("Should not get non-existing Hook", func() {
				_, err := GetHookByID(testDb, -1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Hook was not found with id: -1"))
			})
		})

		Describe("Get Hook By Public ID", func() {
			It("Should get existing Hook", func() {
				hook, err := fixtures.CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/getbyid")
				Expect(err).NotTo(HaveOccurred())

				dbHook, err := GetHookByPublicID(testDb, hook.GameID, hook.PublicID)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbHook.ID).To(Equal(hook.ID))
			})

			It("Should not get non-existing Hook", func() {
				_, err := GetHookByPublicID(testDb, "invalid", "key")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Hook was not found with id: key"))
			})
		})

		Describe("Create Hook", func() {
			It("Should create a new Hook with CreateHook", func() {
				game := fixtures.GameFactory.MustCreate().(*Game)
				err := testDb.Insert(game)
				Expect(err).NotTo(HaveOccurred())

				hook, err := CreateHook(
					testDb,
					game.PublicID,
					GameUpdatedHook,
					"http://test/created",
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(hook.ID).NotTo(BeEquivalentTo(0))

				dbHook, err := GetHookByID(testDb, hook.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbHook.GameID).To(Equal(hook.GameID))
				Expect(dbHook.EventType).To(Equal(hook.EventType))
				Expect(dbHook.URL).To(Equal(hook.URL))
			})

			It("Create same Hook works fine", func() {
				gameID := uuid.NewV4().String()
				hook, err := fixtures.CreateHookFactory(testDb, gameID, GameUpdatedHook, "http://test/created")

				hook2, err := CreateHook(
					testDb,
					gameID,
					GameUpdatedHook,
					"http://test/created",
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(hook2.ID == hook.ID).To(BeTrue())

				dbHook, err := GetHookByID(testDb, hook.ID)
				Expect(err).NotTo(HaveOccurred())

				Expect(dbHook.GameID).To(Equal(hook.GameID))
				Expect(dbHook.EventType).To(Equal(hook.EventType))
				Expect(dbHook.URL).To(Equal(hook.URL))
			})

		})

		Describe("Remove Hook", func() {
			It("Should remove a Hook with RemoveHook", func() {
				hook, err := fixtures.CreateHookFactory(testDb, "", GameUpdatedHook, "http://test/update")
				Expect(err).NotTo(HaveOccurred())

				err = RemoveHook(
					testDb,
					hook.GameID,
					hook.PublicID,
				)

				Expect(err).NotTo(HaveOccurred())

				number, err := testDb.SelectInt("select count(*) from hooks where id=$1", hook.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(number == 0).To(BeTrue())
			})
		})

		Describe("Get All Hooks", func() {
			It("Should get all hooks", func() {
				gameID := uuid.NewV4().String()
				_, err := fixtures.GetTestHooks(testDb, gameID, 5)
				Expect(err).NotTo(HaveOccurred())

				hooks, err := GetAllHooks(testDb)

				Expect(err).NotTo(HaveOccurred())
				Expect(len(hooks)).To(BeNumerically(">", 10))
			})
		})

	})
})
