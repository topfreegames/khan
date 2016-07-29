// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/models"
)

var _ = Describe("Healthcheck API Handler", func() {
	var testDb models.DB

	BeforeEach(func() {
		var err error
		testDb, err = GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Healthcheck Handler", func() {
		It("Should respond with default WORKING string", func() {
			a := GetDefaultTestApp()
			res := Get(a, "/healthcheck")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			Expect(res.Body().Raw()).To(Equal("WORKING"))
		})

		It("Should respond with customized WORKING string", func() {
			a := GetDefaultTestApp()
			a.Config.SetDefault("healthcheck.workingText", "OTHERWORKING")
			res := Get(a, "/healthcheck")

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			Expect(res.Body().Raw()).To(Equal("OTHERWORKING"))
		})
	})
})
