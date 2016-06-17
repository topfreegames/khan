// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/gavv/httpexpect/fasthttpexpect"
)

// GetDefaultTestApp returns a new Khan API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	return GetApp("0.0.0.0", 8888, "../config/test.yaml", true)
}

// Get returns a test request against specified URL
func Get(app *App, url string, t *testing.T) *httpexpect.Response {
	req := sendRequest(app, "GET", url, t)
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(app *App, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(app, "POST", url, t, payload)
}

// PutBody returns a test request against specified URL
func PutBody(app *App, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(app, "PUT", url, t, payload)
}

func sendBody(app *App, method string, url string, t *testing.T, payload string) *httpexpect.Response {
	req := sendRequest(app, method, url, t)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(app *App, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(app, "POST", url, t, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(app *App, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(app, "PUT", url, t, payload)
}

func sendJSON(app *App, method, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	req := sendRequest(app, method, url, t)
	return req.WithJSON(payload).Expect()
}

func sendRequest(app *App, method, url string, t *testing.T) *httpexpect.Request {
	handler := app.App.NoListen().Handler

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   fasthttpexpect.NewBinder(handler),
	})

	return e.Request(method, url)
}

// GetGameRoute returns a clan route for the given game id.
func GetGameRoute(gameID, route string) string {
	return fmt.Sprintf("/games/%s/%s", gameID, route)
}

// CreateMembershipRoute returns a create membership route for the given game and clan id.
func CreateMembershipRoute(gameID, clanPublicID, route string) string {
	return fmt.Sprintf("/games/%s/clans/%s/memberships/%s", gameID, clanPublicID, route)
}
