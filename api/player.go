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
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

// CreatePlayerHandler is the handler responsible for creating new players
func CreatePlayerHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreatePlayer")
		start := time.Now()
		gameID := c.Param("gameID")

		l := app.Logger.With(
			zap.String("source", "playerHandler"),
			zap.String("operation", "createPlayer"),
			zap.String("gameID", gameID),
		)

		var payload createPlayerPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		l.Debug("Creating player...")
		player, err := models.CreatePlayer(
			tx,
			gameID,
			payload.PublicID,
			payload.Name,
			payload.Metadata,
			false,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Player creation failed", l, err)
			if txErr != nil {
				return FailWith(http.StatusInternalServerError, txErr.Error(), c)
			}

			l.Error("Player creation failed.", zap.Error(err))
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		result := map[string]interface{}{
			"success":  true,
			"gameID":   gameID,
			"publicID": player.PublicID,
			"name":     player.Name,
			"metadata": player.Metadata,
		}

		err = app.DispatchHooks(gameID, models.PlayerCreatedHook, player.Serialize())
		if err != nil {
			txErr := app.Rollback(tx, "Player creation hook dispatch failed", l, err)
			if txErr != nil {
				return FailWith(http.StatusInternalServerError, txErr.Error(), c)
			}

			l.Error("Player creation hook dispatch failed.", zap.Error(err))
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Create player", l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		l.Info(
			"Player created successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

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

		l := app.Logger.With(
			zap.String("source", "playerHandler"),
			zap.String("operation", "updatePlayer"),
			zap.String("gameID", gameID),
			zap.String("playerPublicID", playerPublicID),
		)

		var payload updatePlayerPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Updating player failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		l.Debug("Retrieving game...")
		game, err := models.GetGameByPublicID(tx, gameID)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Updating player failed.", zap.Error(err))
			}
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}
		l.Debug("Game retrieved successfully")

		l.Debug("Retrieving player...")
		beforeUpdatePlayer, err := models.GetPlayerByPublicID(tx, gameID, playerPublicID)
		if err != nil && err.Error() != (&models.ModelNotFoundError{"Player", playerPublicID}).Error() {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Updating player failed.", zap.Error(err))
			}
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}
		l.Debug("Player retrieved successfully")

		l.Debug("Updating player...")
		player, err := models.UpdatePlayer(
			tx,
			gameID,
			playerPublicID,
			payload.Name,
			payload.Metadata,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Updating player failed.", zap.Error(err))
			}
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		shouldDispatch := validateUpdatePlayerDispatch(game, beforeUpdatePlayer, player, payload.Metadata, l)
		if shouldDispatch {
			l.Debug("Dispatching player update hooks...")
			err = app.DispatchHooks(gameID, models.PlayerUpdatedHook, player.Serialize())
			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					l.Error("Update player hook dispatch failed.", zap.Error(err))
				}
				return FailWith(http.StatusInternalServerError, err.Error(), c)
			}
		}

		err = app.Commit(tx, "Update game", l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		l.Info(
			"Player updated successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

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

		l := app.Logger.With(
			zap.String("source", "playerHandler"),
			zap.String("operation", "retrievePlayer"),
			zap.String("gameID", gameID),
			zap.String("playerPublicID", publicID),
		)

		l.Debug("Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}
		l.Debug("DB Connection successful.")

		l.Debug("Retrieving player details...")
		player, err := models.GetPlayerDetails(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			if err.Error() == fmt.Sprintf("Player was not found with id: %s", publicID) {
				l.Warn("Player was not found.", zap.Error(err))
				return FailWith(http.StatusNotFound, err.Error(), c)
			}

			l.Error("Retrieve player details failed.", zap.Error(err))
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		l.Info(
			"Player details retrieved successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(player, c)
	}
}
