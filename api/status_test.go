// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/franela/goblin"
	"github.com/topfreegames/khan/util"
)

func TestStatusHandler(t *testing.T) {
	g := Goblin(t)

	g.Describe("Status Handler", func() {
		g.It("Should respond with status", func() {
			a := GetDefaultTestApp()
			res := Get(a, "/status", t)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)

			var result util.JSON
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["app"] != nil).IsTrue()
			app := result["app"].(map[string]interface{})
			g.Assert(app["errorRate"]).Equal(0.0)

			g.Assert(result["dispatch"] != nil).IsTrue()
			dispatch := result["dispatch"].(map[string]interface{})
			g.Assert(int(dispatch["pendingJobs"].(float64))).Equal(0)
		})
	})
}
