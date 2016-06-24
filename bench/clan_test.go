// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/util"
)

var result *http.Response

func BenchmarkCreateClan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		clanPublicID := uuid.NewV4().String()
		payload := util.JSON{
			"publicID":         clanPublicID,
			"name":             clanPublicID,
			"ownerPublicID":    "default-owner",
			"metadata":         util.JSON{"x": 1},
			"allowApplication": true,
			"autoJoin":         true,
		}

		route := getRoute(fmt.Sprintf("/games/default-game/clans"))
		res, err := postTo(route, payload)
		if err != nil {
			panic(err)
		}
		if res.StatusCode != 200 {
			bts, _ := ioutil.ReadAll(res.Body)
			fmt.Printf("Request failed with status code", res.StatusCode)
			panic(string(bts))
		}

		res.Body.Close()

		result = res
	}
}
