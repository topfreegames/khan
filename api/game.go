// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
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
	maxPendingInvites    int
	cooldownBeforeApply  int
	cooldownBeforeInvite int
}

func getOptionalParameters(app *App, c *iris.Context) (*optionalParams, error) {
	data := c.RequestCtx.Request.Body()
	var jsonPayload map[string]interface{}
	err := json.Unmarshal(data, &jsonPayload)
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

	return &optionalParams{
		maxPendingInvites:    maxPendingInvites,
		cooldownBeforeInvite: cooldownBeforeInvite,
		cooldownBeforeApply:  cooldownBeforeApply,
	}, nil
}

func getCreateGamePayload(app *App, c *iris.Context) (*createGamePayload, *optionalParams, error) {
	var payload createGamePayload
	if err := LoadJSONPayload(&payload, c); err != nil {
		return nil, nil, err
	}

	optional, err := getOptionalParameters(app, c)
	if err != nil {
		return nil, nil, err
	}

	return &payload, optional, nil
}

// CreateGameHandler is the handler responsible for creating new games
func CreateGameHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		l := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "createGame"),
		)

		l.Debug("Retrieving parameters...")
		payload, optional, err := getCreateGamePayload(app, c)
		if err != nil {
			l.Error("Failed to retrieve parameters.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}
		l.Debug(
			"Parameters retrieved successfully.",
			zap.Int("maxPendingInvites", optional.maxPendingInvites),
			zap.Int("cooldownBeforeInvite", optional.cooldownBeforeInvite),
			zap.Int("cooldownBeforeApply", optional.cooldownBeforeApply),
		)

		if payloadErrors := validateGamePayload(payload); len(payloadErrors) != 0 {
			app.Metrics.IncrCounter("validationFailure", 1)
			logPayloadErrors(l, payloadErrors)
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Creating game...")
		game, err := models.CreateGame(
			db,
			payload.PublicID,
			payload.Name,
			payload.MembershipLevels,
			payload.Metadata,
			payload.MinLevelToAcceptApplication,
			payload.MinLevelToCreateInvitation,
			payload.MinLevelToRemoveMember,
			payload.MinLevelOffsetToRemoveMember,
			payload.MinLevelOffsetToPromoteMember,
			payload.MinLevelOffsetToDemoteMember,
			payload.MaxMembers,
			payload.MaxClansPerPlayer,
			payload.CooldownAfterDeny,
			payload.CooldownAfterDelete,
			optional.cooldownBeforeApply,
			optional.cooldownBeforeInvite,
			optional.maxPendingInvites,
			false,
		)

		if err != nil {
			l.Error("Create game failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Game created succesfully.",
			zap.Duration("createGameDuration", time.Now().Sub(start)),
		)

		app.Metrics.IncrCounter("games", 1)

		SucceedWith(map[string]interface{}{
			"publicID": game.PublicID,
		}, c)
	}
}

// UpdateGameHandler is the handler responsible for updating existing
func UpdateGameHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		start := time.Now()

		l := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "updateGame"),
			zap.String("gameID", gameID),
		)

		var payload gamePayload

		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		optional, err := getOptionalParameters(app, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		if payloadErrors := validateGamePayload(&payload); len(payloadErrors) != 0 {
			app.Metrics.IncrCounter("validationFailure", 1)
			logPayloadErrors(l, payloadErrors)
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Updating game...")
		_, err = models.UpdateGame(
			db,
			gameID,
			payload.Name,
			payload.MembershipLevels,
			payload.Metadata,
			payload.MinLevelToAcceptApplication,
			payload.MinLevelToCreateInvitation,
			payload.MinLevelToRemoveMember,
			payload.MinLevelOffsetToRemoveMember,
			payload.MinLevelOffsetToPromoteMember,
			payload.MinLevelOffsetToDemoteMember,
			payload.MaxMembers,
			payload.MaxClansPerPlayer,
			payload.CooldownAfterDeny,
			payload.CooldownAfterDelete,
			optional.cooldownBeforeApply,
			optional.cooldownBeforeInvite,
			optional.maxPendingInvites,
		)

		if err != nil {
			l.Error("Game update failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Game updated succesfully.",
			zap.Duration("createGameDuration", time.Now().Sub(start)),
		)

		app.Metrics.IncrCounter("games", 1)

		successPayload := map[string]interface{}{
			"publicID":                      gameID,
			"name":                          payload.Name,
			"membershipLevels":              payload.MembershipLevels,
			"metadata":                      payload.Metadata,
			"minLevelToAcceptApplication":   payload.MinLevelToAcceptApplication,
			"minLevelToCreateInvitation":    payload.MinLevelToCreateInvitation,
			"minLevelToRemoveMember":        payload.MinLevelToRemoveMember,
			"minLevelOffsetToRemoveMember":  payload.MinLevelOffsetToRemoveMember,
			"minLevelOffsetToPromoteMember": payload.MinLevelOffsetToPromoteMember,
			"minLevelOffsetToDemoteMember":  payload.MinLevelOffsetToDemoteMember,
			"maxMembers":                    payload.MaxMembers,
			"maxClansPerPlayer":             payload.MaxClansPerPlayer,
			"cooldownAfterDeny":             payload.CooldownAfterDeny,
			"cooldownAfterDelete":           payload.CooldownAfterDelete,
			"cooldownBeforeApply":           optional.cooldownBeforeApply,
			"cooldownBeforeInvite":          optional.cooldownBeforeInvite,
			"maxPendingInvites":             optional.maxPendingInvites,
		}
		app.DispatchHooks(gameID, models.GameUpdatedHook, successPayload)

		SucceedWith(map[string]interface{}{}, c)
	}
}
