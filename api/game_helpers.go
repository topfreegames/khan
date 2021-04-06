// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"

	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

type optionalParams struct {
	maxPendingInvites                              int
	cooldownBeforeApply                            int
	cooldownBeforeInvite                           int
	clanUpdateMetadataFieldsHookTriggerWhitelist   string
	playerUpdateMetadataFieldsHookTriggerWhitelist string
}

func getOptionalParameters(app *App, c echo.Context) (*optionalParams, error) {
	data, err := GetRequestBody(c)
	if err != nil {
		return nil, err
	}

	var jsonPayload map[string]interface{}
	err = json.Unmarshal(data, &jsonPayload)
	if err != nil {
		return nil, err
	}

	var maxPendingInvites int
	if val, ok := jsonPayload["maxPendingInvites"]; ok {
		maxPendingInvites = int(val.(float64))
	} else {
		maxPendingInvites = app.Config.GetInt("khan.maxPendingInvites")
	}

	var cooldownBeforeInvite int
	if val, ok := jsonPayload["cooldownBeforeInvite"]; ok {
		cooldownBeforeInvite = int(val.(float64))
	} else {
		cooldownBeforeInvite = app.Config.GetInt("khan.defaultCooldownBeforeInvite")
	}

	var cooldownBeforeApply int
	if val, ok := jsonPayload["cooldownBeforeApply"]; ok {
		cooldownBeforeApply = int(val.(float64))
	} else {
		cooldownBeforeApply = app.Config.GetInt("khan.defaultCooldownBeforeApply")
	}

	var clanWhitelist string
	if val, ok := jsonPayload["clanHookFieldsWhitelist"]; ok {
		clanWhitelist = val.(string)
	} else {
		clanWhitelist = ""
	}

	var playerWhitelist string
	if val, ok := jsonPayload["playerHookFieldsWhitelist"]; ok {
		playerWhitelist = val.(string)
	} else {
		playerWhitelist = ""
	}

	return &optionalParams{
		maxPendingInvites:                              maxPendingInvites,
		cooldownBeforeInvite:                           cooldownBeforeInvite,
		cooldownBeforeApply:                            cooldownBeforeApply,
		clanUpdateMetadataFieldsHookTriggerWhitelist:   clanWhitelist,
		playerUpdateMetadataFieldsHookTriggerWhitelist: playerWhitelist,
	}, nil
}

func getCreateGamePayload(app *App, c echo.Context, logger zap.Logger) (*CreateGamePayload, *optionalParams, error) {
	var payload CreateGamePayload
	if err := LoadJSONPayload(&payload, c, logger); err != nil {
		return nil, nil, err
	}
	optional, err := getOptionalParameters(app, c)
	if err != nil {
		return nil, nil, err
	}

	return &payload, optional, nil
}
