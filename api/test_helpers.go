// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/gavv/httpexpect"
	"github.com/gavv/httpexpect/fasthttpexpect"
)

//AssertError asserts that the specified error is not nil
func AssertError(g *goblin.G, err error) {
	g.Assert(err == nil).IsFalse("Expected error to exist, but it was nil")
}

//AssertNotError asserts that the specified error is nil
func AssertNotError(g *goblin.G, err error) {
	if err != nil {
		g.Assert(err == nil).IsTrue(err.Error())
	}
}

//GetDefaultTestApp returns a new Khan API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	return GetApp("0.0.0.0", 8888, "./config/test.yaml", true)
}

//Get returns a test request against specified URL
func Get(app *App, url string, t *testing.T) *httpexpect.Response {
	req := sendRequest(app, "GET", url, t)
	return req.Expect()
}

//PostBody returns a test request against specified URL
func PostBody(app *App, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(app, "POST", url, t, payload)
}

//PutBody returns a test request against specified URL
func PutBody(app *App, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(app, "PUT", url, t, payload)
}

func sendBody(app *App, method string, url string, t *testing.T, payload string) *httpexpect.Response {
	req := sendRequest(app, method, url, t)
	return req.WithBytes([]byte(payload)).Expect()
}

//PostJSON returns a test request against specified URL
func PostJSON(app *App, url string, t *testing.T, payload map[string]string) *httpexpect.Response {
	return sendJSON(app, "POST", url, t, payload)
}

//PutJSON returns a test request against specified URL
func PutJSON(app *App, url string, t *testing.T, payload map[string]string) *httpexpect.Response {
	return sendJSON(app, "PUT", url, t, payload)
}

func sendJSON(app *App, method string, url string, t *testing.T, payload map[string]string) *httpexpect.Response {
	req := sendRequest(app, method, url, t)
	return req.WithJSON(payload).Expect()
}

func sendRequest(app *App, method string, url string, t *testing.T) *httpexpect.Request {
	handler := app.App.ServeRequest

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   fasthttpexpect.NewBinder(handler),
	})

	return e.Request(method, url)
}
