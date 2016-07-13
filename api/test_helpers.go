// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/franela/goblin"
	"github.com/topfreegames/khan/models"
	kt "github.com/topfreegames/khan/testing"
	"gopkg.in/gavv/httpexpect.v1"
)

// GetDefaultTestApp returns a new Khan API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	l := kt.NewMockLogger()
	app := GetApp("0.0.0.0", 8888, "../config/test.yaml", true, l)
	app.Configure()
	return app
}

// Get returns a test request against specified URL
func Get(app *App, url string, t *testing.T, queryString ...map[string]interface{}) *httpexpect.Response {
	req := sendRequest(app, "GET", url, t)
	if len(queryString) == 1 {
		for k, v := range queryString[0] {
			req = req.WithQuery(k, v)
		}
	}
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

// Delete returns a test request against specified URL
func Delete(app *App, url string, t *testing.T) *httpexpect.Response {
	req := sendRequest(app, "DELETE", url, t)
	return req.Expect()
}

func sendRequest(app *App, method, url string, t *testing.T) *httpexpect.Request {
	handler := app.App.NoListen().Handler

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: fmt.Sprintf("http://%s:%d", app.Host, app.Port),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
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

// LoadGames load the games
func LoadGames(a *App) map[string]*models.Game {
	g := make(map[string]*models.Game)
	games, _ := models.GetAllGames(a.Db)
	for _, game := range games {
		g[game.PublicID] = game
	}
	return g
}

func str(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func validateMembershipHookResponse(g *goblin.G, apply map[string]interface{}, gameID string, clan *models.Clan, player *models.Player, requestor *models.Player) {
	g.Assert(apply["gameID"]).Equal(gameID)

	rClan := apply["clan"].(map[string]interface{})
	g.Assert(rClan["publicID"]).Equal(clan.PublicID)
	g.Assert(rClan["name"]).Equal(clan.Name)
	g.Assert(rClan["membershipCount"].(float64) >= 0).IsTrue()
	g.Assert(rClan["allowApplication"]).Equal(clan.AllowApplication)
	g.Assert(rClan["autoJoin"]).Equal(clan.AutoJoin)
	clanMetadata := rClan["metadata"].(map[string]interface{})
	metadata := clan.Metadata
	for k, v := range clanMetadata {
		g.Assert(v).Equal(metadata[k])
	}

	rPlayer := apply["player"].(map[string]interface{})
	g.Assert(rPlayer["publicID"]).Equal(player.PublicID)
	g.Assert(rPlayer["name"]).Equal(player.Name)
	g.Assert(rPlayer["membershipCount"].(float64) >= 0).IsTrue()
	g.Assert(rPlayer["ownershipCount"].(float64) >= 0).IsTrue()
	playerMetadata := rPlayer["metadata"].(map[string]interface{})
	metadata = player.Metadata
	for pk, pv := range playerMetadata {
		g.Assert(pv).Equal(metadata[pk])
	}

	rPlayer = apply["requestor"].(map[string]interface{})
	g.Assert(rPlayer["publicID"]).Equal(requestor.PublicID)
	g.Assert(rPlayer["name"]).Equal(requestor.Name)
	g.Assert(rPlayer["membershipCount"].(float64) >= 0).IsTrue()
	g.Assert(rPlayer["ownershipCount"].(float64) >= 0).IsTrue()
	playerMetadata = rPlayer["metadata"].(map[string]interface{})
	metadata = requestor.Metadata
	for rk, rv := range playerMetadata {
		g.Assert(rv).Equal(metadata[rk])
	}
}
