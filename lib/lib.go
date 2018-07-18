package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	ehttp "github.com/topfreegames/extensions/http"
)

// Khan is a struct that represents a khan API application
type Khan struct {
	httpClient *http.Client
	Config     *viper.Viper
	url        string
	user       string
	pass       string
	gameID     string
}

var (
	client *http.Client
	once   sync.Once
)

func getHTTPClient(timeoutMs int) *http.Client {
	once.Do(func() {
		client = &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		}
		ehttp.Instrument(client)
	})
	return client
}

// NewKhan returns a new khan API application
func NewKhan(config *viper.Viper) KhanInterface {
	config.SetDefault("khan.timeout", 1000)
	k := &Khan{
		httpClient: getHTTPClient(config.GetInt("khan.timeout")),
		Config:     config,
		url:        config.GetString("khan.url"),
		user:       config.GetString("khan.user"),
		pass:       config.GetString("khan.pass"),
		gameID:     config.GetString("khan.gameid"),
	}
	return k
}

func (k *Khan) sendTo(ctx context.Context, method, url string, payload interface{}) ([]byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(k.user, k.pass)
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.WithContext(ctx)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, respErr := ioutil.ReadAll(resp.Body)
	if respErr != nil {
		return nil, respErr
	}

	if resp.StatusCode > 399 {
		return nil, newRequestError(resp.StatusCode, string(body))
	}

	return body, nil
}

func (k *Khan) buildURL(pathname string) string {
	return fmt.Sprintf("%s/games/%s/%s", k.url, k.gameID, pathname)
}

func (k *Khan) buildCreatePlayerURL() string {
	pathname := "players"
	return k.buildURL(pathname)
}

func (k *Khan) buildUpdatePlayerURL(playerID string) string {
	pathname := fmt.Sprintf("players/%s", playerID)
	return k.buildURL(pathname)
}

func (k *Khan) buildRetrievePlayerURL(playerID string) string {
	pathname := fmt.Sprintf("players/%s", playerID)
	return k.buildURL(pathname)
}

func (k *Khan) buildCreateClanURL() string {
	pathname := "clans"
	return k.buildURL(pathname)
}

func (k *Khan) buildUpdateClanURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildRetrieveClanURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildRetrieveClanSummaryURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/summary", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildRetrieveClansSummaryURL(clanIDs []string) string {
	pathname := fmt.Sprintf("clans-summary?clanPublicIds=%s", strings.Join(clanIDs, ","))
	return k.buildURL(pathname)
}

// CreatePlayer calls Khan to create a new player
func (k *Khan) CreatePlayer(ctx context.Context, publicID, name string, metadata interface{}) (string, error) {
	route := k.buildCreatePlayerURL()
	playerPayload := &Player{
		PublicID: publicID,
		Name:     name,
		Metadata: metadata,
	}
	body, err := k.sendTo(ctx, "POST", route, playerPayload)

	if err != nil {
		return "", err
	}

	var player Player
	err = json.Unmarshal(body, &player)

	return player.PublicID, err
}

// UpdatePlayer calls khan to update the player
func (k *Khan) UpdatePlayer(ctx context.Context, publicID, name string, metadata interface{}) error {
	route := k.buildUpdatePlayerURL(publicID)
	playerPayload := &Player{
		Name:     name,
		Metadata: metadata,
	}
	_, err := k.sendTo(ctx, "PUT", route, playerPayload)
	return err
}

// RetrievePlayer calls the retrieve player route from khan
func (k *Khan) RetrievePlayer(ctx context.Context, publicID string) (*Player, error) {
	route := k.buildRetrievePlayerURL(publicID)
	body, err := k.sendTo(ctx, "GET", route, nil)

	if err != nil {
		return nil, err
	}

	var player Player
	err = json.Unmarshal(body, &player)
	return &player, err
}

// CreateClan calls the create clan route from khan
func (k *Khan) CreateClan(ctx context.Context, clan *ClanPayload) (string, error) {
	route := k.buildCreateClanURL()
	body, err := k.sendTo(ctx, "POST", route, clan)

	if err != nil {
		return "", err
	}

	var player Player
	err = json.Unmarshal(body, &player)
	return player.PublicID, err
}

// UpdateClan calls the update clan route from khan
func (k *Khan) UpdateClan(ctx context.Context, clan *ClanPayload) error {
	route := k.buildUpdateClanURL(clan.PublicID)
	_, err := k.sendTo(ctx, "PUT", route, clan)
	return err
}

// RetrieveClanSummary calls the route to retrieve clan summary from khan
func (k *Khan) RetrieveClanSummary(ctx context.Context, clanID string) (*ClanSummary, error) {
	route := k.buildRetrieveClanSummaryURL(clanID)
	body, err := k.sendTo(ctx, "GET", route, nil)

	if err != nil {
		return nil, err
	}

	var clanSummary ClanSummary
	err = json.Unmarshal(body, &clanSummary)
	return &clanSummary, err
}

// RetrieveClansSummary calls the route to retrieve clans summary from khan
func (k *Khan) RetrieveClansSummary(ctx context.Context, clanIDs []string) ([]*ClanSummary, error) {
	route := k.buildRetrieveClansSummaryURL(clanIDs)
	body, err := k.sendTo(ctx, "GET", route, nil)

	if err != nil {
		return nil, err
	}

	var clansSummary ClansSummary
	err = json.Unmarshal(body, &clansSummary)
	if err != nil {
		return nil, err
	}
	return clansSummary.Clans, nil
}

// RetrieveClan calls the route to retrieve clan from khan
func (k *Khan) RetrieveClan(ctx context.Context, clanID string) (*Clan, error) {
	route := k.buildRetrieveClanURL(clanID)
	body, err := k.sendTo(ctx, "GET", route, nil)

	if err != nil {
		return nil, err
	}

	var clan Clan
	err = json.Unmarshal(body, &clan)
	return &clan, err
}
