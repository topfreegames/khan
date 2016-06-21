// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"os/exec"
	"testing"

	. "github.com/franela/goblin"
)

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

func TestMigrationCommand(t *testing.T) {
	g := Goblin(t)

	g.Describe("Migrate Cmd", func() {
		g.BeforeEach(func() {
			err := dropDB()
			g.Assert(err == nil).IsTrue()
		})

		g.It("Should run migrations up", func() {
			ConfigFile = "../config/test.yaml"
			initConfig()
			err := runMigrations(-1)
			g.Assert(err == nil).IsTrue()
		})
	})
}
