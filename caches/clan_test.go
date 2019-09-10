package caches_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/testing"
)

var _ = Describe("Clan Cache", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = testing.GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Clans Summaries", func() {
		It("Should return a cached payload for a second call made immediately after the first", func() {
			gameID := uuid.NewV4().String()
			_, clans, err := models.GetTestClans(testDb, gameID, "test-sort-clan", 10)
			Expect(err).NotTo(HaveOccurred())

			var publicIDs []string
			clansMap := make(map[string]int)
			for i, clan := range clans {
				publicIDs = append(publicIDs, clan.PublicID)
				clansMap[clan.PublicID] = i + 1
			}

			cache := testing.GetTestClansSummariesCache()

			// first call
			clansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(clansSummaries)).To(Equal(len(clans)))
			for _, clanPayload := range clansSummaries {
				// assert public ID
				publicID, ok := clanPayload["publicID"].(string)
				Expect(ok).To(BeTrue())
				Expect(clansMap[publicID]).To(BeNumerically(">", 0))

				// assert name
				name, ok := clanPayload["name"].(string)
				Expect(ok).To(BeTrue())
				Expect(name).To(Equal(clans[clansMap[publicID]-1].Name))
			}

			// update a clan
			const diffName = "different name"
			clans[0].Name = diffName
			_, err = testDb.Update(clans[0])
			Expect(err).NotTo(HaveOccurred())

			// second call
			secondClansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(secondClansSummaries)).To(Equal(len(clans)))
			secondPayload := secondClansSummaries[0]
			secondPublicID, ok := secondPayload["publicID"].(string)
			Expect(ok).To(BeTrue())
			Expect(secondPublicID).To(Equal(clans[0].PublicID))
			firstName, ok := clansSummaries[0]["name"].(string)
			Expect(ok).To(BeTrue())
			secondName, ok := secondPayload["name"].(string)
			Expect(ok).To(BeTrue())
			Expect(secondName).To(Equal(firstName))
			Expect(secondName).NotTo(Equal(clans[0].Name))
		})

		It("Should return fresh information after expiration time is reached", func() {
			gameID := uuid.NewV4().String()
			_, clans, err := models.GetTestClans(testDb, gameID, "test-sort-clan", 10)
			Expect(err).NotTo(HaveOccurred())

			var publicIDs []string
			clansMap := make(map[string]int)
			for i, clan := range clans {
				publicIDs = append(publicIDs, clan.PublicID)
				clansMap[clan.PublicID] = i + 1
			}

			cache := testing.GetTestClansSummariesCache()
			cache.TTL = time.Second / 2
			cache.TTLRandomError = 0

			// first call
			clansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(clansSummaries)).To(Equal(len(clans)))
			for _, clanPayload := range clansSummaries {
				// assert public ID
				publicID, ok := clanPayload["publicID"].(string)
				Expect(ok).To(BeTrue())
				Expect(clansMap[publicID]).To(BeNumerically(">", 0))

				// assert name
				name, ok := clanPayload["name"].(string)
				Expect(ok).To(BeTrue())
				Expect(name).To(Equal(clans[clansMap[publicID]-1].Name))
			}

			// update a clan
			const diffName = "different name"
			clans[0].Name = diffName
			_, err = testDb.Update(clans[0])
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)

			// second call
			secondClansSummaries, err := cache.GetClansSummaries(testDb, gameID, publicIDs)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(secondClansSummaries)).To(Equal(len(clans)))
			secondPayload := secondClansSummaries[0]
			secondPublicID, ok := secondPayload["publicID"].(string)
			Expect(ok).To(BeTrue())
			Expect(secondPublicID).To(Equal(clans[0].PublicID))
			firstName, ok := clansSummaries[0]["name"].(string)
			Expect(ok).To(BeTrue())
			secondName, ok := secondPayload["name"].(string)
			Expect(ok).To(BeTrue())
			Expect(secondName).NotTo(Equal(firstName))
			Expect(secondName).To(Equal(clans[0].Name))
		})
	})
})
