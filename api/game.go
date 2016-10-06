// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"strings"
	"time"

	gorp "gopkg.in/gorp.v1"

	"github.com/labstack/echo"
	"github.com/topfreegames/khan/log"
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

		log.D(l, "Retrieving parameters...")
		var err error
		var payload *CreateGamePayload
		var optional *optionalParams

		err = WithSegment("payload", c, func() error {
			payload, optional, err = getCreateGamePayload(app, c, l)
			return err
		})
		if err != nil {
			log.E(l, "Failed to retrieve parameters.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Parameters retrieved successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("maxPendingInvites", optional.maxPendingInvites),
				zap.Int("cooldownBeforeInvite", optional.cooldownBeforeInvite),
				zap.Int("cooldownBeforeApply", optional.cooldownBeforeApply),
			)
		})

		var game *models.Game
		var tx *gorp.Transaction
		err = WithSegment("game-create", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			log.D(l, "Creating game...")
			game, err = models.CreateGame(
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
				txErr := app.Rollback(tx, "Create game failed", c, l, err)
				if txErr != nil {
					return txErr
				}
				log.E(l, "Create game failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}

			err = app.Commit(tx, "Create game", c, l)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.I(l, "Game created succesfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "gameHandler"),
			zap.String("operation", "updateGame"),
			zap.String("gameID", gameID),
		)

		var payload UpdateGamePayload
		var optional *optionalParams
		var status int

		err := WithSegment("payload", c, func() error {
			var err error
			log.D(l, "Retrieving parameters...")
			if err = LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Failed to retrieve parameters.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				status = 400
				return err
			}

			optional, err = getOptionalParameters(app, c)
			if err != nil {
				log.E(l, "Failed to retrieve optional parameters.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				status = 400
				return err
			}

			log.D(l, "Parameters retrieved successfully.", func(cm log.CM) {
				cm.Write(
					zap.Int("maxPendingInvites", optional.maxPendingInvites),
					zap.Int("cooldownBeforeInvite", optional.cooldownBeforeInvite),
					zap.Int("cooldownBeforeApply", optional.cooldownBeforeApply),
				)
			})
			log.D(l, "Validating payload...")
			if payloadErrors := ValidatePayload(&payload); len(payloadErrors) != 0 {
				status = 422
				logPayloadErrors(l, payloadErrors)
				errorString := strings.Join(payloadErrors[:], ", ")
				return fmt.Errorf(errorString)
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		var tx *gorp.Transaction
		err = WithSegment("game-update", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}

			log.D(l, "Updating game...")
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
			return err
		})

		if err != nil {
			txErr := app.Rollback(tx, "Game update failed", c, l, err)
			if txErr != nil {
				return FailWith(500, txErr.Error(), c)
			}

			log.E(l, "Game update failed.", func(cm log.CM) {
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

		err = WithSegment("hook-dispatch", c, func() error {
			dErr := app.DispatchHooks(gameID, models.GameUpdatedHook, successPayload)
			if dErr != nil {
				txErr := app.Rollback(tx, "Game update hook dispatch failed", c, l, dErr)
				if txErr != nil {
					return txErr
				}

				log.E(l, "Game update hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(dErr))
				})
				return dErr
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		err = app.Commit(tx, "Update game", c, l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.I(l, "Game updated succesfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{}, c)
	}
}
