// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

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
			zap.Duration("duration", time.Now().Sub(start)),
		)

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

		l.Debug("Retrieving parameters...")
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			l.Error("Failed to retrieve parameters.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		optional, err := getOptionalParameters(app, c)
		if err != nil {
			l.Error("Failed to retrieve optional parameters.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		l.Debug(
			"Parameters retrieved successfully.",
			zap.Int("maxPendingInvites", optional.maxPendingInvites),
			zap.Int("cooldownBeforeInvite", optional.cooldownBeforeInvite),
			zap.Int("cooldownBeforeApply", optional.cooldownBeforeApply),
		)

		l.Debug("Validating payload...")
		if payloadErrors := validateGamePayload(&payload); len(payloadErrors) != 0 {
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
			zap.Duration("duration", time.Now().Sub(start)),
		)

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
