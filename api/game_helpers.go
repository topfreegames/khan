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
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
)

type validatable interface {
	Validate() []string
}

type gamePayload struct {
	Name                          string
	MembershipLevels              map[string]interface{}
	Metadata                      map[string]interface{}
	MinLevelToAcceptApplication   int
	MinLevelToCreateInvitation    int
	MinLevelToRemoveMember        int
	MinLevelOffsetToRemoveMember  int
	MinLevelOffsetToPromoteMember int
	MinLevelOffsetToDemoteMember  int
	MaxMembers                    int
	MaxClansPerPlayer             int
	CooldownAfterDeny             int
	CooldownAfterDelete           int
}

func (p *gamePayload) Validate() []string {
	sortedLevels := util.SortLevels(p.MembershipLevels)
	minMembershipLevel := sortedLevels[0].Value

	var errors []string
	if p.MinLevelToAcceptApplication < minMembershipLevel {
		errors = append(errors, "minLevelToAcceptApplication should be greater or equal to minMembershipLevel")
	}
	if p.MinLevelToCreateInvitation < minMembershipLevel {
		errors = append(errors, "minLevelToCreateInvitation should be greater or equal to minMembershipLevel")
	}
	if p.MinLevelToRemoveMember < minMembershipLevel {
		errors = append(errors, "minLevelToRemoveMember should be greater or equal to minMembershipLevel")
	}
	return errors
}

type createGamePayload struct {
	PublicID                      string
	Name                          string
	MembershipLevels              map[string]interface{}
	Metadata                      map[string]interface{}
	MinLevelToAcceptApplication   int
	MinLevelToCreateInvitation    int
	MinLevelToRemoveMember        int
	MinLevelOffsetToRemoveMember  int
	MinLevelOffsetToPromoteMember int
	MinLevelOffsetToDemoteMember  int
	MaxMembers                    int
	MaxClansPerPlayer             int
	CooldownAfterDeny             int
	CooldownAfterDelete           int
}

func (p *createGamePayload) Validate() []string {
	sortedLevels := util.SortLevels(p.MembershipLevels)
	minMembershipLevel := sortedLevels[0].Value

	var errors []string
	if p.MinLevelToAcceptApplication < minMembershipLevel {
		errors = append(errors, "minLevelToAcceptApplication should be greater or equal to minMembershipLevel")
	}
	if p.MinLevelToCreateInvitation < minMembershipLevel {
		errors = append(errors, "minLevelToCreateInvitation should be greater or equal to minMembershipLevel")
	}
	if p.MinLevelToRemoveMember < minMembershipLevel {
		errors = append(errors, "minLevelToRemoveMember should be greater or equal to minMembershipLevel")
	}
	return errors
}

func validateGamePayload(payload validatable) []string {
	return payload.Validate()
}

func logPayloadErrors(l zap.Logger, errors []string) {
	var fields []zap.Field
	for _, err := range errors {
		fields = append(fields, zap.String("validationError", err))
	}
	l.Warn(
		"Payload is not valid",
		fields...,
	)
}

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
	err = json.Unmarshal([]byte(data), &jsonPayload)
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

func getCreateGamePayload(app *App, c echo.Context, l zap.Logger) (*createGamePayload, *optionalParams, error) {
	var payload createGamePayload
	if err := LoadJSONPayload(&payload, c, l); err != nil {
		return nil, nil, err
	}

	optional, err := getOptionalParameters(app, c)
	if err != nil {
		return nil, nil, err
	}

	return &payload, optional, nil
}
