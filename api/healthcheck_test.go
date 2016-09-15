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
)

var _ = Describe("Healthcheck API Handler", func() {
	Describe("Healthcheck Handler", func() {
		It("Should respond with default WORKING string", func() {
			a := GetDefaultTestApp()
			status, body := Get(a, "/healthcheck")
			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("WORKING"))
		})

		It("Should respond with customized WORKING string", func() {
			a := GetDefaultTestApp()
			a.Config.Set("healthcheck.workingText", "OTHERWORKING")
			status, body := Get(a, "/healthcheck")
			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("OTHERWORKING"))
		})
	})
})
