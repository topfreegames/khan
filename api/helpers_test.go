// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/uber-go/zap"

	"github.com/labstack/echo/engine/standard"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/api"
	"github.com/topfreegames/khan/es"
	"github.com/topfreegames/khan/models"
	kt "github.com/topfreegames/khan/testing"
)

var testES *es.Client

func GetTestES() *es.Client {
	if testES != nil {
		return testES
	}
	testES = es.GetTestClient("localhost", 9200, "", false, zap.New(zap.NewJSONEncoder(), zap.ErrorLevel), false)
	return testES
}

func DestroyTestES() {
	es.DestroyClient()
}

//InvalidJSONError returned by the API
var InvalidJSONError = "syntax error near offset"

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
	logger := kt.NewMockLogger()
	app := api.GetApp("0.0.0.0", 8888, "../config/test.yaml", true, logger, false, true)
	app.Configure()
	return app
}

// GetTestAppWithBasicAuth returns a new Khan API application bound to 0.0.0.0:8888 for test with basic auth configs
func GetTestAppWithBasicAuth(username, password string) *api.App {
	logger := kt.NewMockLogger()
	app := api.GetApp("0.0.0.0", 8888, "../config/test.yaml", true, logger, false, true)
	app.Config.Set("basicauth.username", username)
	app.Config.Set("basicauth.password", password)
	app.Configure()
	return app
}

//Get from server
func Get(app *api.App, url string) (int, string) {
	return doRequest(app, "GET", url, "")
}

//Post to server
func Post(app *api.App, url, body string) (int, string) {
	return doRequest(app, "POST", url, body)
}

//PostJSON to server
func PostJSON(app *api.App, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Post(app, url, string(result))
}

//Put to server
func Put(app *api.App, url, body string) (int, string) {
	return doRequest(app, "PUT", url, body)
}

//PutJSON to server
func PutJSON(app *api.App, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Put(app, url, string(result))
}

//Delete from server
func Delete(app *api.App, url string) (int, string) {
	return doRequest(app, "DELETE", url, "")
}

var client *http.Client
var transport *http.Transport

func initClient() {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}
}

func InitializeTestServer(app *api.App) *httptest.Server {
	initClient()
	app.Engine.SetHandler(app.App)
	return httptest.NewServer(app.Engine.(*standard.Server))
}

func GetRequest(app *api.App, ts *httptest.Server, method, url, body string) *http.Request {
	var bodyBuff io.Reader
	if body != "" {
		bodyBuff = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", ts.URL, url), bodyBuff)
	req.Header.Set("Connection", "close")
	req.Close = true
	Expect(err).NotTo(HaveOccurred())

	return req
}

func PerformRequest(ts *httptest.Server, req *http.Request) (int, string) {
	res, err := client.Do(req)
	//Wait for port of httptest to be reclaimed by OS
	time.Sleep(50 * time.Millisecond)
	Expect(err).NotTo(HaveOccurred())

	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	Expect(err).NotTo(HaveOccurred())

	return res.StatusCode, string(b)
}

func doRequest(app *api.App, method, url, body string) (int, string) {
	ts := InitializeTestServer(app)
	defer transport.CloseIdleConnections()
	defer ts.Close()

	req := GetRequest(app, ts, method, url, body)
	return PerformRequest(ts, req)
}

// GetGameRoute returns a clan route for the given game id.
func GetGameRoute(gameID, route string) string {
	return fmt.Sprintf("/games/%s/%s", gameID, strings.TrimPrefix(route, "/"))
}

// CreateMembershipRoute returns a create membership route for the given game and clan id.
func CreateMembershipRoute(gameID, clanPublicID, route string) string {
	return fmt.Sprintf("/games/%s/clans/%s/memberships/%s", gameID, clanPublicID, strings.TrimPrefix(route, "/"))
}

// LoadGames load the games
func LoadGames(a *api.App) map[string]*models.Game {
	g := make(map[string]*models.Game)
	games, _ := models.GetAllGames(a.Db(nil))
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
	Expect(rPlayer["membershipLevel"]).ToNot(BeEmpty())
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
