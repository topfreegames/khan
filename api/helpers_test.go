// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/models"
	kt "github.com/topfreegames/khan/testing"
	"github.com/valyala/fasthttp"
	"gopkg.in/gavv/httpexpect.v1"
)

// GetTestDB returns a connection to the test database
func GetTestDB() (models.DB, error) {
	return models.GetDB("localhost", "khan_test", 5433, "disable", "khan_test", "")
}

// GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB() models.DB {
	faultyDb, _ := models.InitDb("localhost", "khan_tet", 5433, "disable", "khan_test", "")
	return faultyDb
}

// GetDefaultTestApp returns a new Khan API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *api.App {
	l := kt.NewMockLogger()
	app := api.GetApp("0.0.0.0", 8888, "../config/test.yaml", true, l)
	app.Configure()
	return app
}

// Get returns a test request against specified URL
func Get(app *api.App, url string, queryString ...map[string]interface{}) *httpexpect.Response {
	req := sendRequest(app, "GET", url)
	if len(queryString) == 1 {
		for k, v := range queryString[0] {
			req = req.WithQuery(k, v)
		}
	}
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(app *api.App, url string, payload string) *httpexpect.Response {
	return sendBody(app, "POST", url, payload)
}

// PutBody returns a test request against specified URL
func PutBody(app *api.App, url string, payload string) *httpexpect.Response {
	return sendBody(app, "PUT", url, payload)
}

func sendBody(app *api.App, method string, url string, payload string) *httpexpect.Response {
	req := sendRequest(app, method, url)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(app *api.App, url string, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(app, "POST", url, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(app *api.App, url string, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(app, "PUT", url, payload)
}

func sendJSON(app *api.App, method, url string, payload map[string]interface{}) *httpexpect.Response {
	req := sendRequest(app, method, url)
	return req.WithJSON(payload).Expect()
}

// Delete returns a test request against specified URL
func Delete(app *api.App, url string) *httpexpect.Response {
	req := sendRequest(app, "DELETE", url)
	return req.Expect()
}

//GinkgoReporter implements tests for httpexpect
type GinkgoReporter struct {
}

// Errorf implements Reporter.Errorf.
func (g *GinkgoReporter) Errorf(message string, args ...interface{}) {
	Expect(false).To(BeTrue(), fmt.Sprintf(message, args...))
}

//GinkgoPrinter reports errors to stdout
type GinkgoPrinter struct{}

//Logf reports to stdout
func (g *GinkgoPrinter) Logf(source string, args ...interface{}) {
	fmt.Printf(source, args...)
}

func sendRequest(app *api.App, method, url string) *httpexpect.Request {
	api := app.App
	srv := api.Servers.Main()

	if srv == nil { // maybe the user called this after .Listen/ListenTLS/ListenUNIX, the tester can be used as standalone (with no running iris instance) or inside a running instance/app
		srv = api.ListenVirtual(api.Config.Tester.ListeningAddr)
	}

	opened := api.Servers.GetAllOpened()
	h := srv.Handler
	baseURL := srv.FullHost()
	if len(opened) > 1 {
		baseURL = ""
		//we have more than one server, so we will create a handler here and redirect by registered listening addresses
		h = func(reqCtx *fasthttp.RequestCtx) {
			for _, s := range opened {
				if strings.HasPrefix(reqCtx.URI().String(), s.FullHost()) { // yes on :80 should be passed :80 also, this is inneed for multiserver testing
					s.Handler(reqCtx)
					break
				}
			}
		}
	}

	if api.Config.Tester.ExplicitURL {
		baseURL = ""
	}

	testConfiguration := httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(h),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: &GinkgoReporter{},
	}
	if api.Config.Tester.Debug {
		testConfiguration.Printers = []httpexpect.Printer{
			httpexpect.NewDebugPrinter(&GinkgoPrinter{}, true),
		}
	}

	return httpexpect.WithConfig(testConfiguration).Request(method, url)
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
func LoadGames(a *api.App) map[string]*models.Game {
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

func validateMembershipHookResponse(apply map[string]interface{}, gameID string, clan *models.Clan, player *models.Player, requestor *models.Player) {
	Expect(apply["gameID"]).To(Equal(gameID))

	rClan := apply["clan"].(map[string]interface{})
	Expect(rClan["publicID"]).To(Equal(clan.PublicID))
	Expect(rClan["name"]).To(Equal(clan.Name))
	Expect(rClan["membershipCount"].(float64)).To(BeNumerically(">=", 0))
	Expect(rClan["allowApplication"]).To(Equal(clan.AllowApplication))
	Expect(rClan["autoJoin"]).To(Equal(clan.AutoJoin))
	clanMetadata := rClan["metadata"].(map[string]interface{})
	metadata := clan.Metadata
	for k, v := range clanMetadata {
		Expect(v).To(Equal(metadata[k]))
	}

	rPlayer := apply["player"].(map[string]interface{})
	Expect(rPlayer["publicID"]).To(Equal(player.PublicID))
	Expect(rPlayer["name"]).To(Equal(player.Name))
	Expect(rPlayer["membershipCount"].(float64)).To(BeNumerically(">=", 0))
	Expect(rPlayer["ownershipCount"].(float64)).To(BeNumerically(">=", 0))
	playerMetadata := rPlayer["metadata"].(map[string]interface{})
	metadata = player.Metadata
	for pk, pv := range playerMetadata {
		Expect(pv).To(Equal(metadata[pk]))
	}

	rPlayer = apply["requestor"].(map[string]interface{})
	Expect(rPlayer["publicID"]).To(Equal(requestor.PublicID))
	Expect(rPlayer["name"]).To(Equal(requestor.Name))
	Expect(rPlayer["membershipCount"].(float64)).To(BeNumerically(">=", 0))
	Expect(rPlayer["ownershipCount"].(float64)).To(BeNumerically(">=", 0))
	playerMetadata = rPlayer["metadata"].(map[string]interface{})
	metadata = requestor.Metadata
	for rk, rv := range playerMetadata {
		Expect(rv).To(Equal(metadata[rk]))
	}
}

func validateApproveDenyMembershipHookResponse(apply map[string]interface{}, creator *models.Player) {
	cPlayer := apply["creator"].(map[string]interface{})
	Expect(cPlayer["publicID"]).To(Equal(creator.PublicID))
	Expect(cPlayer["name"]).To(Equal(creator.Name))
	Expect(cPlayer["membershipCount"].(float64)).To(BeNumerically(">=", 0))
	Expect(cPlayer["ownershipCount"].(float64)).To(BeNumerically(">=", 0))
	playerMetadata := cPlayer["metadata"].(map[string]interface{})
	metadata := creator.Metadata
	for rk, rv := range playerMetadata {
		Expect(rv).To(Equal(metadata[rk]))
	}
}
