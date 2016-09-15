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

	"github.com/labstack/echo"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

//CreateGameHandler is the handler responsible for creating new games
func CreateGameHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreateGame")
		start := time.Now()
		l := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "createGame"),
		)

		l.Debug("Retrieving parameters...")
		payload, optional, err := getCreateGamePayload(app, c, l)
		if err != nil {
			l.Error("Failed to retrieve parameters.", zap.Error(err))
			return FailWith(400, err.Error(), c)
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
			return FailWith(422, errorString, c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}
		l.Debug("DB Tx begun successful.")

		l.Debug("Creating game...")
		game, err := models.CreateGame(
			tx,
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
			optional.clanUpdateMetadataFieldsHookTriggerWhitelist,
			optional.playerUpdateMetadataFieldsHookTriggerWhitelist,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Create game failed", l, err)
			if txErr != nil {
				return FailWith(500, txErr.Error(), c)
			}
			l.Error("Create game failed.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

		err = app.Commit(tx, "Create game", l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		l.Info(
			"Game created succesfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{
			"publicID": game.PublicID,
		}, c)
	}
}

//UpdateGameHandler is the handler responsible for updating existing
func UpdateGameHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "UpdateGame")
		start := time.Now()
		gameID := c.Param("gameID")

		l := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "updateGame"),
			zap.String("gameID", gameID),
		)

		var payload gamePayload

		l.Debug("Retrieving parameters...")
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			l.Error("Failed to retrieve parameters.", zap.Error(err))
			return FailWith(400, err.Error(), c)
		}

		optional, err := getOptionalParameters(app, c)
		if err != nil {
			l.Error("Failed to retrieve optional parameters.", zap.Error(err))
			return FailWith(400, err.Error(), c)
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
			return FailWith(422, errorString, c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		l.Debug("Updating game...")
		_, err = models.UpdateGame(
			tx,
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
			optional.clanUpdateMetadataFieldsHookTriggerWhitelist,
			optional.playerUpdateMetadataFieldsHookTriggerWhitelist,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Game update failed", l, err)
			if txErr != nil {
				return FailWith(500, txErr.Error(), c)
			}

			l.Error("Game update failed.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

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
		dErr := app.DispatchHooks(gameID, models.GameUpdatedHook, successPayload)
		if dErr != nil {
			txErr := app.Rollback(tx, "Game update hook dispatch failed", l, dErr)
			if txErr != nil {
				return FailWith(500, txErr.Error(), c)
			}

			l.Error("Game update hook dispatch failed.", zap.Error(dErr))
			return FailWith(500, dErr.Error(), c)
		}

		err = app.Commit(tx, "Update game", l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		l.Info(
			"Game updated succesfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{}, c)
	}
}
