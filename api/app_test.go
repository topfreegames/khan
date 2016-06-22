// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/topfreegames/khan/models"
)

func startRouteHandler(routes []string, port int) *[]map[string]interface{} {
	responses := []map[string]interface{}{}

	go func() {
		handleFunc := func(w http.ResponseWriter, r *http.Request) {
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				responses = append(responses, map[string]interface{}{"reason": err})
				return
			}

			var payload map[string]interface{}
			json.Unmarshal(bs, &payload)

			responses = append(responses, payload)
		}
		for _, route := range routes {
			http.HandleFunc(route, handleFunc)
		}

		http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	}()

	return &responses
}

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
			hooks, err := models.GetHooksForRoutes(testDb, []string{
				"http://localhost:52525/created",
				"http://localhost:52525/created2",
			}, models.GameCreatedHook)
			g.Assert(err == nil).IsTrue()
			responses := startRouteHandler([]string{"/created", "/created2"}, 52525)

			app := GetDefaultTestApp()
			//app.loadHooks()
			time.Sleep(time.Second)

			resultingPayload := map[string]interface{}{
				"success":  true,
				"publicID": hooks[0].GameID,
			}
			err = app.DispatchHooks(hooks[0].GameID, models.GameCreatedHook, resultingPayload)
			g.Assert(err == nil).IsTrue()

			g.Assert(len(*responses)).Equal(2)
		})
	})
}
