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
	"github.com/uber-go/zap"
)

var _ = Describe("Prune Stale Data Model", func() {
	var testDb DB
	var logger zap.Logger

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())

		logger = zap.New(
			zap.NewJSONEncoder(), // drop timestamps in tests
			zap.FatalLevel,
		)
	})

	Describe("Prune Stale Data Model", func() {
		Describe("Pruning Pending Applications", func() {
			It("Should remove pending applications", func() {
				err := GetTestClanWithStaleData(testDb, 20)
				Expect(err).NotTo(HaveOccurred())

				expiration := int((2 * time.Hour).Seconds())
				pruneStats, err := PruneStaleData(expiration, testDb, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(pruneStats).NotTo(BeNil())
				Expect(pruneStats.PendingApplicationsPruned).To(Equal(20))
			})
		})
	})
})
