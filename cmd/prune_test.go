// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	. "github.com/topfreegames/khan/cmd"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/models/fixtures"
)

var _ = Describe("Prune Command", func() {
	var db models.DB
	var err error

	BeforeEach(func() {
		ConfigFile = "../config/test.yaml"
		InitConfig()

		host := viper.GetString("postgres.host")
		user := viper.GetString("postgres.user")
		dbName := viper.GetString("postgres.dbname")
		password := viper.GetString("postgres.password")
		port := viper.GetInt("postgres.port")
		sslMode := viper.GetString("postgres.sslMode")

		db, err = models.GetDB(host, user, port, sslMode, dbName, password)
		Expect(err).NotTo(HaveOccurred())

		_, err = db.Exec("TRUNCATE TABLE memberships CASCADE")
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("TRUNCATE TABLE players CASCADE")
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("TRUNCATE TABLE clans CASCADE")
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("TRUNCATE TABLE games CASCADE")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Prune Cmd", func() {
		It("Should prune old data", func() {
			totalApps := 0
			totalInvites := 0
			totalDenies := 0
			totalDeletes := 0

			for i := 0; i < 5; i++ {
				apps := rand.Intn(10)
				invites := rand.Intn(10)
				denies := rand.Intn(10)
				deletes := rand.Intn(10)
				_, err := fixtures.GetTestClanWithStaleData(db, apps, invites, denies, deletes)
				Expect(err).NotTo(HaveOccurred())
				totalApps += apps
				totalInvites += invites
				totalDenies += denies
				totalDeletes += deletes
			}
			stats, err := PruneStaleData(false, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(stats.PendingApplicationsPruned).To(Equal(totalApps))
			Expect(stats.PendingInvitesPruned).To(Equal(totalInvites))
			Expect(stats.DeniedMembershipsPruned).To(Equal(totalDenies))
			Expect(stats.DeletedMembershipsPruned).To(Equal(totalDeletes))

			count, err := db.SelectInt("select count(*) from memberships")
			Expect(err).NotTo(HaveOccurred())
			Expect(int(count)).To(Equal((totalApps + totalInvites + totalDenies + totalDeletes) * 2))
		})

		It("Should not prune games without metadata", func() {
			totalApps := 0
			totalInvites := 0
			totalDenies := 0
			totalDeletes := 0

			for i := 0; i < 5; i++ {
				apps := rand.Intn(10)
				invites := rand.Intn(10)
				denies := rand.Intn(10)
				deletes := rand.Intn(10)
				gameID, err := fixtures.GetTestClanWithStaleData(db, apps, invites, denies, deletes)
				Expect(err).NotTo(HaveOccurred())

				_, err = db.Exec("UPDATE games SET metadata='{}' WHERE public_id=$1", gameID)
				Expect(err).NotTo(HaveOccurred())

				totalApps += apps
				totalInvites += invites
				totalDenies += denies
				totalDeletes += deletes
			}
			stats, err := PruneStaleData(false, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(stats.PendingApplicationsPruned).To(Equal(0))
			Expect(stats.PendingInvitesPruned).To(Equal(0))
			Expect(stats.DeniedMembershipsPruned).To(Equal(0))
			Expect(stats.DeletedMembershipsPruned).To(Equal(0))

			count, err := db.SelectInt("select count(*) from memberships")
			Expect(err).NotTo(HaveOccurred())
			Expect(int(count)).To(Equal((totalApps + totalInvites + totalDenies + totalDeletes) * 3))
		})
	})
})
