// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/franela/goblin"
	"github.com/topfreegames/khan/api"
)

func runVersion() (string, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(goBin, "run", "main.go", "version")
	cmd.Dir = ".."
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func TestVersionCommand(t *testing.T) {
	g := Goblin(t)

	g.Describe("Version Cmd", func() {
		g.It("Should get version", func() {
			version, err := runVersion()
			g.Assert(err == nil).IsTrue()
			g.Assert(version).Equal(fmt.Sprintf("Khan v%s\n", api.VERSION))
		})
	})
}
