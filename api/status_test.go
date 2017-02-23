// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status API Handler", func() {
	Describe("Status Handler", func() {
		It("Should respond with status", func() {
			a := GetDefaultTestApp()
			status, body := Get(a, "/status")

			Expect(status).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["app"]).NotTo(BeEquivalentTo(nil))
			app := result["app"].(map[string]interface{})
			Expect(app["errorRate"]).To(Equal(0.0))
		})
	})
})
