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
	"github.com/topfreegames/khan/models/fixtures"
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
		Describe("Pruning Stale data", func() {
			It("Should remove pending applications", func() {
				gameID, err := fixtures.GetTestClanWithStaleData(testDb, 5, 6, 7, 8)
				Expect(err).NotTo(HaveOccurred())

				expiration := int((2 * time.Hour).Seconds())
				options := &PruneOptions{
					GameID:                        gameID,
					PendingApplicationsExpiration: expiration,
					PendingInvitesExpiration:      expiration,
					DeniedMembershipsExpiration:   expiration,
					DeletedMembershipsExpiration:  expiration,
				}
				pruneStats, err := PruneStaleData(options, testDb, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(pruneStats).NotTo(BeNil())
				Expect(pruneStats.PendingApplicationsPruned).To(Equal(5))
				Expect(pruneStats.PendingInvitesPruned).To(Equal(6))
				Expect(pruneStats.DeniedMembershipsPruned).To(Equal(7))
				Expect(pruneStats.DeletedMembershipsPruned).To(Equal(8))

				count, err := testDb.SelectInt(`SELECT COUNT(*) FROM memberships WHERE game_id=$1`, gameID)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(count)).To(Equal(52))
			})
		})
	})
})
