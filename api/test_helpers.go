// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/gavv/httpexpect/fasthttpexpect"
)

//GetDefaultTestApp returns a new Khan API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	return GetApp("0.0.0.0", 8888, "./config/test.yaml", true)
}

//Get returns a test request against specified URL
func Get(app *App, url string, t *testing.T) *httpexpect.Response {
	handler := app.App.ServeRequest

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   fasthttpexpect.NewBinder(handler),
	})

	return e.GET(url).Expect()
}
