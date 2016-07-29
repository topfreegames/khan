// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/khan/cmd"
	"github.com/topfreegames/khan/models"
)

// GetTestDB returns a connection to the test database
func GetTestDB() (models.DB, error) {
	return models.GetDB("localhost", "khan_test", 5432, "disable", "khan_test", "")
}

func dropDB() error {
	cmd := exec.Cmd{
		Dir:  "../",
		Path: "/usr/bin/make",
		Args: []string{
			"drop-test",
		},
	}
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

var _ = Describe("Migrate Command", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())

		err = dropDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Migrate Cmd", func() {
		It("Should run migrations up", func() {
			ConfigFile = "../config/test.yaml"
			InitConfig()
			err := RunMigrations(-1)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
