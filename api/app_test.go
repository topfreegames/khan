// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/topfreegames/khan/models"
)

func Test(t *testing.T) {
	g := Goblin(t)

	testDb, err := models.GetTestDB()

	g.Assert(err == nil).IsTrue()

	g.Describe("App Struct", func() {
		g.It("should create app with custom arguments", func() {
			app := GetApp("127.0.0.1", 9999, "../config/test.yaml", false)
			g.Assert(app.Port).Equal(9999)
			g.Assert(app.Host).Equal("127.0.0.1")
		})
	})

	g.Describe("App Load Hooks", func() {
		g.It("should load all hooks", func() {
			app := GetDefaultTestApp()

			_, err := models.GetTestHooks(testDb, "app-game-id", 2)
			g.Assert(err == nil).IsTrue()

			app.loadHooks()
			time.Sleep(time.Second)

			g.Assert(len(app.Hooks["app-game-id"])).Equal(2)
			g.Assert(len(app.Hooks["app-game-id"][0])).Equal(2)
			g.Assert(len(app.Hooks["app-game-id"][1])).Equal(2)
		})
	})
}
