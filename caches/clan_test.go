package caches_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
	"github.com/topfreegames/khan/testing"
)

var _ = Describe("Clan Cache", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = testing.GetTestDB()
		Expect(err).NotTo(HaveOccurred())

		fixtures.ConfigureAndStartGoWorkers()
	})

	Describe("Clans Summaries", func() {
		getPublicIDsAndIDToIndexMap := func(clans []*models.Clan) ([]string, map[string]int) {
			var publicIDs []string
			idToIdx := make(map[string]int)
			for i, clan := range clans {
				publicIDs = append(publicIDs, clan.PublicID)
				idToIdx[clan.PublicID] = i + 1
			}
			return publicIDs, idToIdx
		}

		assertFirstCacheCall := func(clans []*models.Clan, idToIdx map[string]int, clansSummaries []map[string]interface{}) {
			Expect(len(clansSummaries)).To(Equal(len(clans)))
			for _, clanPayload := range clansSummaries {
				// assert public ID
				publicID, ok := clanPayload["publicID"].(string)
				Expect(ok).To(BeTrue())
				Expect(idToIdx[publicID]).To(BeNumerically(">", 0))

				// assert name
				name, ok := clanPayload["name"].(string)
				Expect(ok).To(BeTrue())
				Expect(name).To(Equal(clans[idToIdx[publicID]-1].Name))
			}
		}

		updateClan := func(db models.DB, clan *models.Clan) {
			clan.Name = "different name"
			_, err := db.Update(clan)
			Expect(err).NotTo(HaveOccurred())
		}

		assertSecondCacheCall := func(clans []*models.Clan, clansSummaries, secondClansSummaries []map[string]interface{}, shouldBeChanged bool) {
			Expect(len(secondClansSummaries)).To(Equal(len(clans)))

			// assert public ID
			secondPayload := secondClansSummaries[0]
			secondPublicID, ok := secondPayload["publicID"].(string)
			Expect(ok).To(BeTrue())
			Expect(secondPublicID).To(Equal(clans[0].PublicID))

			// assert name
			firstName, ok := clansSummaries[0]["name"].(string)
			Expect(ok).To(BeTrue())
			secondName, ok := secondPayload["name"].(string)
			Expect(ok).To(BeTrue())
			if shouldBeChanged {
				Expect(secondName).NotTo(Equal(firstName))
				Expect(secondName).To(Equal(clans[0].Name))
			} else {
				Expect(secondName).To(Equal(firstName))
				Expect(secondName).NotTo(Equal(clans[0].Name))
			}
		}

		It("Should return a cached payload for a second call made immediately after the first", func() {
			mongoDB, err := testing.GetTestMongo()
			Expect(err).NotTo(HaveOccurred())

			gameID := uuid.NewV4().String()
			_, clans, err := fixtures.CreateTestClans(testDb, mongoDB, gameID, "test-sort-clan", 10, fixtures.EnqueueClanForMongoUpdate)
			Expect(err).NotTo(HaveOccurred())

			publicIDs, idToIdx := getPublicIDsAndIDToIndexMap(clans)

			cache := testing.GetTestClansSummariesCache(time.Minute, time.Minute)

			// first call
			clansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			assertFirstCacheCall(clans, idToIdx, clansSummaries)

			// update a clan
			updateClan(testDb, clans[0])

			// second call
			secondClansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			assertSecondCacheCall(clans, clansSummaries, secondClansSummaries, false)
		})

		It("Should return fresh information after expiration time is reached", func() {
			mongoDB, err := testing.GetTestMongo()
			Expect(err).NotTo(HaveOccurred())

			gameID := uuid.NewV4().String()
			_, clans, err := fixtures.CreateTestClans(testDb, mongoDB, gameID, "test-sort-clan", 10, fixtures.EnqueueClanForMongoUpdate)
			Expect(err).NotTo(HaveOccurred())

			publicIDs, idToIdx := getPublicIDsAndIDToIndexMap(clans)

			cache := testing.GetTestClansSummariesCache(time.Second/4, time.Minute)

			// first call
			clansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			assertFirstCacheCall(clans, idToIdx, clansSummaries)

			// update a clan
			updateClan(testDb, clans[0])
			time.Sleep(time.Second / 2)

			// second call
			secondClansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			assertSecondCacheCall(clans, clansSummaries, secondClansSummaries, true)
		})
	})
})
