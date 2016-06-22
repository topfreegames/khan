// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/satori/go.uuid"
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

	g.Describe("App Dispatch Hook", func() {
		g.It("should dispatch hooks", func() {
			responses := []map[string]interface{}{}
			addResponse := func(payload map[string]interface{}) {
				responses = append(responses, payload)
			}

			go func() {
				handleFunc := func(w http.ResponseWriter, r *http.Request) {
					bs, err := ioutil.ReadAll(r.Body)
					g.Assert(err == nil).IsTrue()

					var payload map[string]interface{}
					json.Unmarshal(bs, &payload)

					addResponse(payload)
				}
				http.HandleFunc("/created", handleFunc)
				http.HandleFunc("/created2", handleFunc)

				http.ListenAndServe(":52525", nil)
			}()

			app := GetDefaultTestApp()

			hook, err := models.CreateHookFactory(testDb, "", models.GameCreatedHook, "http://localhost:52525/created")
			g.Assert(err == nil).IsTrue()

			hook2 := models.HookFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":    hook.GameID,
				"PublicID":  uuid.NewV4().String(),
				"EventType": models.GameCreatedHook,
				"URL":       "http://localhost:52525/created2",
			}).(*models.Hook)
			err = testDb.Insert(hook2)
			g.Assert(err == nil).IsTrue()

			app.loadHooks()
			time.Sleep(time.Second)

			resultingPayload := map[string]interface{}{
				"success":  true,
				"publicID": hook.GameID,
			}
			err = app.DispatchHooks(hook.GameID, models.GameCreatedHook, resultingPayload)
			g.Assert(err == nil).IsTrue()

			g.Assert(len(responses)).Equal(2)
		})
	})
}
