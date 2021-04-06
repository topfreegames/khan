// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

// CreatePlayerHandler is the handler responsible for creating new players
func CreatePlayerHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreatePlayer")
		start := time.Now()
		gameID := c.Param("gameID")

		logger := app.Logger.With(
			zap.String("source", "playerHandler"),
			zap.String("operation", "createPlayer"),
			zap.String("gameID", gameID),
		)

		var payload CreatePlayerPayload
		if err := LoadJSONPayload(&payload, c, logger); err != nil {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		var transaction interfaces.Transaction
		transaction, err := app.BeginTrans(c.StdContext(), logger)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.D(logger, "Creating player...")
		player, err := models.CreatePlayer(
			transaction,
			logger,
			app.EncryptionKey,
			gameID,
			payload.PublicID,
			payload.Name,
			payload.Metadata,
		)

		if err != nil {
			log.E(logger, "Player creation failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})

			txErr := app.Rollback(transaction, "Player creation failed, rolling back", c, logger, err)
			if txErr != nil {
				return FailWith(http.StatusInternalServerError, fmt.Sprint(err.Error(), ", rolback error: ", txErr.Error()), c)
			}
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(transaction, "Player created successful", c, logger)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		result := map[string]interface{}{
			"success":  true,
			"gameID":   gameID,
			"publicID": player.PublicID,
			"name":     player.Name,
			"metadata": player.Metadata,
		}

		err = app.DispatchHooks(
			gameID,
			models.PlayerCreatedHook,
			player.Serialize(app.EncryptionKey),
		)
		if err != nil {
			log.E(logger, "Player creation hook dispatch failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.D(logger, "Player created successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(result, c)
	}
}

// UpdatePlayerHandler is the handler responsible for updating existing
func UpdatePlayerHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "UpdatePlayer")
		start := time.Now()
		gameID := c.Param("gameID")
		playerPublicID := c.Param("playerPublicID")

		db := app.Db(c.StdContext())

		logger := app.Logger.With(
			zap.String("source", "playerHandler"),
			zap.String("operation", "updatePlayer"),
			zap.String("gameID", gameID),
			zap.String("playerPublicID", playerPublicID),
		)

		var payload UpdatePlayerPayload
		err := LoadJSONPayload(&payload, c, logger)
		if err != nil {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		var player, beforeUpdatePlayer *models.Player
		var game *models.Game

		log.D(logger, "Retrieving game...")
		game, err = models.GetGameByPublicID(db, gameID)

		if err != nil {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}
		log.D(logger, "Game retrieved successfully")

		log.D(logger, "Retrieving player...")
		beforeUpdatePlayer, err = models.GetPlayerByPublicID(db, app.EncryptionKey, gameID, playerPublicID)
		if err != nil && err.Error() != (&models.ModelNotFoundError{Type: "Player", ID: playerPublicID}).Error() {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}
		log.D(logger, "Player retrieved successfully")

		var transaction interfaces.Transaction
		transaction, err = app.BeginTrans(c.StdContext(), logger)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.D(logger, "Updating player...")
		player, err = models.UpdatePlayer(
			transaction,
			logger,
			app.EncryptionKey,
			gameID,
			playerPublicID,
			payload.Name,
			payload.Metadata,
		)

		if err != nil {
			log.E(logger, "Updating player failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})

			txErr := app.Rollback(transaction, "Player update failed, rolling back", c, logger, err)
			if txErr != nil {
				return FailWith(http.StatusInternalServerError, fmt.Sprint(err.Error(), ", rolback error: ", txErr.Error()), c)
			}
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(transaction, "Player created successful", c, logger)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		shouldDispatch := validateUpdatePlayerDispatch(game, beforeUpdatePlayer, player, payload.Metadata, logger)
		if shouldDispatch {
			log.D(logger, "Dispatching player update hooks...")
			err = app.DispatchHooks(
				gameID,
				models.PlayerUpdatedHook,
				player.Serialize(app.EncryptionKey),
			)
			if err != nil {
				log.E(logger, "Update player hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(http.StatusInternalServerError, err.Error(), c)
			}
		}

		log.D(logger, "Player updated successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})
		return SucceedWith(map[string]interface{}{}, c)
	}
}

// RetrievePlayerHandler is the handler responsible for returning details for a given player
func RetrievePlayerHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "RetrievePlayer")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("playerPublicID")

		logger := app.Logger.With(
			zap.String("source", "playerHandler"),
			zap.String("operation", "retrievePlayer"),
			zap.String("gameID", gameID),
			zap.String("playerPublicID", publicID),
		)

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		log.D(logger, "Retrieving player details...")
		player, err := models.GetPlayerDetails(
			db,
			app.EncryptionKey,
			gameID,
			publicID,
		)

		if err != nil {
			if err.Error() == fmt.Sprintf("Player was not found with id: %s", publicID) {
				log.D(logger, "Player was not found.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(http.StatusNotFound, err.Error(), c)
			}

			log.E(logger, "Retrieve player details failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.D(logger, "Player details retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(player, c)
	}
}
