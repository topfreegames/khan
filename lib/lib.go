package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	ehttp "github.com/topfreegames/extensions/v9/http"
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

// KhanParams represents the params to create a Khan client
type KhanParams struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	URL                 string
	User                string
	Pass                string
	GameID              string
}

var (
	client *http.Client
	once   sync.Once
)

func getHTTPClient(
	timeout time.Duration,
	maxIdleConns, maxIdleConnsPerHost int,
) *http.Client {
	once.Do(func() {
		client = &http.Client{
			Transport: getHTTPTransport(maxIdleConns, maxIdleConnsPerHost),
			Timeout:   timeout,
		}
		ehttp.Instrument(client)
	})
	return client
}

func getHTTPTransport(
	maxIdleConns, maxIdleConnsPerHost int,
) http.RoundTripper {
	if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		return http.DefaultTransport // tests use a mock transport
	}

	// We can't get http.DefaultTransport here and update its
	// fields since it's an exported variable, so other libs could
	// also change it and overwrite. This hardcoded values are copied
	// from http.DefaultTransport but could be configurable too.
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
	}
}

// NewKhan returns a new khan API application
func NewKhan(config *viper.Viper) KhanInterface {
	config.SetDefault("khan.timeout", 500*time.Millisecond)
	config.SetDefault("khan.maxIdleConnsPerHost", http.DefaultMaxIdleConnsPerHost)
	config.SetDefault("khan.maxIdleConns", 100)

	k := &Khan{
		httpClient: getHTTPClient(
			config.GetDuration("khan.timeout"),
			config.GetInt("khan.maxIdleConns"),
			config.GetInt("khan.maxIdleConnsPerHost"),
		),
		Config: config,
		url:    config.GetString("khan.url"),
		user:   config.GetString("khan.user"),
		pass:   config.GetString("khan.pass"),
		gameID: config.GetString("khan.gameid"),
	}
	return k
}

// NewKhanParams returns a new KhanParams instance with default values
func NewKhanParams() *KhanParams {
	return &KhanParams{
		Timeout:             500 * time.Millisecond,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: http.DefaultMaxIdleConnsPerHost,
	}
}

// NewKhanWithParams returns a new khan API application initialized with passed params
func NewKhanWithParams(params *KhanParams) KhanInterface {
	return &Khan{
		httpClient: getHTTPClient(
			params.Timeout,
			params.MaxIdleConns,
			params.MaxIdleConnsPerHost,
		),
		url:    params.URL,
		user:   params.User,
		pass:   params.Pass,
		gameID: params.GameID,
	}
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

func (k *Khan) buildRetrieveClanMembersURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/members", clanID)
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

func (k *Khan) buildApplyForMembershipURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/memberships/application", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildInviteForMembershipURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/memberships/invitation", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildApproveDenyMembershipApplicationURL(clanID, action string) string {
	pathname := fmt.Sprintf("clans/%s/memberships/application/%s", clanID, action)
	return k.buildURL(pathname)
}

func (k *Khan) buildApproveDenyMembershipInvitationURL(clanID, action string) string {
	pathname := fmt.Sprintf("clans/%s/memberships/invitation/%s", clanID, action)
	return k.buildURL(pathname)
}

func (k *Khan) buildPromoteDemoteURL(clanID, action string) string {
	pathname := fmt.Sprintf("clans/%s/memberships/%s", clanID, action)
	return k.buildURL(pathname)
}

func (k *Khan) buildDeleteMembershipURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/memberships/delete", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildLeaveClanURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/leave", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildTransferOwnershipURL(clanID string) string {
	pathname := fmt.Sprintf("clans/%s/transfer-ownership", clanID)
	return k.buildURL(pathname)
}

func (k *Khan) buildSearchClansURL(clanName string) string {
	pathname := fmt.Sprintf("clans/search?term=%s", clanName)
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
func (k *Khan) UpdatePlayer(
	ctx context.Context,
	publicID, name string,
	metadata interface{},
) (*Result, error) {
	route := k.buildUpdatePlayerURL(publicID)
	playerPayload := &Player{Name: name, Metadata: metadata}
	body, err := k.sendTo(ctx, "PUT", route, playerPayload)
	if err != nil {
		return nil, err
	}

	var result Result
	err = json.Unmarshal(body, &result)
	return &result, err
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
func (k *Khan) UpdateClan(ctx context.Context, clan *ClanPayload) (*Result, error) {
	route := k.buildUpdateClanURL(clan.PublicID)
	body, err := k.sendTo(ctx, "PUT", route, clan)
	if err != nil {
		return nil, err
	}

	var result Result
	err = json.Unmarshal(body, &result)
	return &result, err
}

// RetrieveClanMembers calls the route to retrieve clan members from khan
func (k *Khan) RetrieveClanMembers(ctx context.Context, clanID string) (*ClanMembers, error) {
	route := k.buildRetrieveClanMembersURL(clanID)
	body, err := k.sendTo(ctx, "GET", route, nil)

	if err != nil {
		return nil, err
	}

	var clanMembers ClanMembers
	err = json.Unmarshal(body, &clanMembers)
	return &clanMembers, err
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

// ApplyForMembership calls apply for membership route on khan
func (k *Khan) ApplyForMembership(
	ctx context.Context,
	payload *ApplicationPayload,
) (*ClanApplyResult, error) {
	route := k.buildApplyForMembershipURL(payload.ClanID)
	body, err := k.sendTo(ctx, "POST", route, payload)

	if err != nil {
		return nil, err
	}

	var application ClanApplyResult
	err = json.Unmarshal(body, &application)
	return &application, err
}

// InviteForMembership invites a clan member to join clan
func (k *Khan) InviteForMembership(
	ctx context.Context,
	payload *InvitationPayload,
) (*Result, error) {
	route := k.buildInviteForMembershipURL(payload.ClanID)
	return k.defaultPostRequest(ctx, route, payload)
}

// ApproveDenyMembershipApplication approves or deny player
// application on clan
func (k *Khan) ApproveDenyMembershipApplication(
	ctx context.Context,
	payload *ApplicationApprovalPayload,
) (*Result, error) {
	route := k.buildApproveDenyMembershipApplicationURL(payload.ClanID, payload.Action)
	return k.defaultPostRequest(ctx, route, payload)
}

// ApproveDenyMembershipInvitation approves or deny player
// invitation on clan
func (k *Khan) ApproveDenyMembershipInvitation(
	ctx context.Context,
	payload *InvitationApprovalPayload,
) (*Result, error) {
	route := k.buildApproveDenyMembershipInvitationURL(payload.ClanID, payload.Action)
	return k.defaultPostRequest(ctx, route, payload)
}

// PromoteDemote promotes or demotes player on clan
func (k *Khan) PromoteDemote(
	ctx context.Context,
	payload *PromoteDemotePayload,
) (*Result, error) {
	route := k.buildPromoteDemoteURL(payload.ClanID, payload.Action)
	return k.defaultPostRequest(ctx, route, payload)
}

// DeleteMembership deletes membership on clan
func (k *Khan) DeleteMembership(
	ctx context.Context,
	payload *DeleteMembershipPayload,
) (*Result, error) {
	route := k.buildDeleteMembershipURL(payload.ClanID)
	return k.defaultPostRequest(ctx, route, payload)
}

// LeaveClan allows member to leave clan
func (k *Khan) LeaveClan(
	ctx context.Context,
	clanID string,
) (*LeaveClanResult, error) {
	route := k.buildLeaveClanURL(clanID)
	body, err := k.sendTo(ctx, "POST", route, nil)

	if err != nil {
		return nil, err
	}

	var result LeaveClanResult
	err = json.Unmarshal(body, &result)
	return &result, err
}

// TransferOwnership transfers clan ownership to another member
func (k *Khan) TransferOwnership(
	ctx context.Context,
	playerPublicID, clanID string,
) (*TransferOwnershipResult, error) {
	route := k.buildTransferOwnershipURL(clanID)
	body, err := k.sendTo(ctx, "POST", route, map[string]interface{}{
		"playerPublicID": playerPublicID,
	})

	if err != nil {
		return nil, err
	}

	var result TransferOwnershipResult
	err = json.Unmarshal(body, &result)
	return &result, err
}

func (k *Khan) defaultPostRequest(
	ctx context.Context,
	route string,
	payload interface{},
) (*Result, error) {
	body, err := k.sendTo(ctx, "POST", route, payload)
	if err != nil {
		return nil, err
	}

	var result Result
	err = json.Unmarshal(body, &result)
	return &result, err
}

// SearchClans returns clan summaries for all clans that contain the string "clanName".
func (k *Khan) SearchClans(ctx context.Context, clanName string) (*SearchClansResult, error) {
	route := k.buildSearchClansURL(clanName)
	body, err := k.sendTo(ctx, "GET", route, nil)
	if err != nil {
		return nil, err
	}

	var result SearchClansResult
	err = json.Unmarshal(body, &result)
	return &result, err
}
