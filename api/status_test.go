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
	"github.com/topfreegames/khan/models"
)

var _ = Describe("Status API Handler", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Status Handler", func() {
		It("Should respond with status", func() {
			a := GetDefaultTestApp()
			res := Get(a, "/status")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["app"]).NotTo(BeEquivalentTo(nil))
			app := result["app"].(map[string]interface{})
			Expect(app["errorRate"]).To(Equal(0.0))

			Expect(result["dispatch"]).NotTo(BeEquivalentTo(nil))
			dispatch := result["dispatch"].(map[string]interface{})
			Expect(dispatch["pendingJobs"].(float64)).To(BeEquivalentTo(0))
		})
	})
})
