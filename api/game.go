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
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
	"github.com/uber-go/zap"
)

//CreateGameHandler is the handler responsible for creating new games
func CreateGameHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreateGame")
		start := time.Now()

		logger := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "createGame"),
		)

		log.D(logger, "Retrieving parameters...")
		var err error
		var optional *optionalParams

		payload, optional, err := getCreateGamePayload(app, c, logger)
		if err != nil {
			log.E(logger, "Failed to retrieve parameters.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		log.D(logger, "Parameters retrieved successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("maxPendingInvites", optional.maxPendingInvites),
				zap.Int("cooldownBeforeInvite", optional.cooldownBeforeInvite),
				zap.Int("cooldownBeforeApply", optional.cooldownBeforeApply),
			)
		})

		tx, err := app.BeginTrans(c.StdContext(), logger)
		if err != nil {
			log.E(logger, "Could not start transaction", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		log.D(logger, "Creating game...")
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
			log.E(logger, "Create game failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			app.Rollback(tx, "Game", c, logger, err)
			return FailWith(500, err.Error(), c)
		}

		if app.MongoDB != nil {
			err = app.MongoDB.Run(mongo.GetClanNameTextIndexCommand(game.PublicID, false), nil)
			if err != nil {
				app.Rollback(tx, "Game", c, logger, err)
			}
		}

		txErr := app.Commit(tx, "Game", c, logger)
		if txErr != nil {
			return FailWith(500, err.Error(), c)
		}

		log.I(logger, "Game created succesfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

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

		db := app.Db(c.StdContext())

		logger := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "updateGame"),
			zap.String("gameID", gameID),
		)

		var payload UpdateGamePayload

		log.D(logger, "Retrieving parameters...")
		if err := LoadJSONPayload(&payload, c, logger); err != nil {
			log.E(logger, "Failed to retrieve parameters.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)

		}

		optional, err := getOptionalParameters(app, c)
		if err != nil {
			log.E(logger, "Failed to retrieve optional parameters.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		log.D(logger, "Parameters retrieved successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("maxPendingInvites", optional.maxPendingInvites),
				zap.Int("cooldownBeforeInvite", optional.cooldownBeforeInvite),
				zap.Int("cooldownBeforeApply", optional.cooldownBeforeApply),
			)
		})
		log.D(logger, "Validating payload...")
		if payloadErrors := ValidatePayload(&payload); len(payloadErrors) != 0 {
			logPayloadErrors(logger, payloadErrors)
			errorString := strings.Join(payloadErrors[:], ", ")
			return FailWith(422, errorString, c)
		}

		log.D(logger, "Updating game...")
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
			optional.clanUpdateMetadataFieldsHookTriggerWhitelist,
			optional.playerUpdateMetadataFieldsHookTriggerWhitelist,
		)

		if err != nil {
			log.E(logger, "Game update failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
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
			log.E(logger, "Game update hook dispatch failed.", func(cm log.CM) {
				cm.Write(zap.Error(dErr))
			})
			return FailWith(500, err.Error(), c)
		}

		log.I(logger, "Game updated succesfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{}, c)
	}
}
