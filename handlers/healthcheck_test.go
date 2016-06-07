package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/api"
)

func Test(t *testing.T) {
	g := Goblin(t)

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Healthcheck Handler", func() {
		g.It("should get healthcheck", func() {
			app := api.GetDefaultApp()
			app.AddHandlers(api.URL{
				Method:  "GET",
				Path:    "/healthcheck",
				Handler: HealthcheckHandler,
			})

			r, _ := http.NewRequest("GET", "/healthcheck", nil)
			w := httptest.NewRecorder()
			app.App.ServeHTTP(w, r)
			Expect(w.Code).To(Equal(200))
			Expect(w.Body.String()).To(Equal("WORKING"))
		})
	})
}
