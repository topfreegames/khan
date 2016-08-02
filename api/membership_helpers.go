// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

type applyForMembershipPayload struct {
	Level          string
	PlayerPublicID string
}

type inviteForMembershipPayload struct {
	Level             string
	PlayerPublicID    string
	RequestorPublicID string
}

type basePayloadWithRequestorAndPlayerPublicIDs struct {
	PlayerPublicID    string
	RequestorPublicID string
}

type approveOrDenyMembershipInvitationPayload struct {
	PlayerPublicID string
}

func dispatchMembershipHookByPublicID(app *App, db models.DB, hookType int, gameID, clanID, playerID, requestorID string) error {
	clan, err := models.GetClanByPublicID(db, gameID, clanID)
	if err != nil {
		return err
	}

	player, err := models.GetPlayerByPublicID(db, gameID, playerID)
	if err != nil {
		return err
	}

	requestor := player
	if requestorID != playerID {
		requestor, err = models.GetPlayerByPublicID(db, gameID, requestorID)
		if err != nil {
			return err
		}
	}

	return dispatchMembershipHook(app, db, hookType, gameID, clan, player, requestor, "")
}

func dispatchMembershipHookByID(app *App, db models.DB, hookType int, gameID string, clanID, playerID, requestorID int, message string) error {
	clan, err := models.GetClanByID(db, clanID)
	if err != nil {
		return err
	}

	player, err := models.GetPlayerByID(db, playerID)
	if err != nil {
		return err
	}

	requestor := player
	if requestorID != playerID {
		requestor, err = models.GetPlayerByID(db, requestorID)
		if err != nil {
			return err
		}
	}

	return dispatchMembershipHook(app, db, hookType, gameID, clan, player, requestor, message)
}

func dispatchMembershipHook(app *App, db models.DB, hookType int, gameID string, clan *models.Clan, player *models.Player, requestor *models.Player, message string) error {
	clanJSON := clan.Serialize()
	delete(clanJSON, "gameID")

	playerJSON := player.Serialize()
	delete(playerJSON, "gameID")

	requestorJSON := requestor.Serialize()
	delete(requestorJSON, "gameID")

	result := map[string]interface{}{
		"gameID":    gameID,
		"clan":      clanJSON,
		"player":    playerJSON,
		"requestor": requestorJSON,
	}

	if message != "" {
		result["message"] = message
	}
	app.DispatchHooks(gameID, hookType, result)

	return nil
}

func getPayloadAndGame(app *App, c *iris.Context, l zap.Logger) (*basePayloadWithRequestorAndPlayerPublicIDs, *models.Game, int, error) {
	gameID := c.Param("gameID")

	var payload basePayloadWithRequestorAndPlayerPublicIDs
	if err := LoadJSONPayload(&payload, c, l.With(zap.String("gameID", gameID))); err != nil {
		return nil, nil, 400, err
	}

	game, err := app.GetGame(gameID)
	if err != nil {
		l.Warn("Could not find game.")
		return nil, nil, 404, err
	}

	return &payload, game, 200, nil
}

func registerApplicationMetric(app *App, action, playerPublicID, requestorPublicID string) {
	invite := playerPublicID == requestorPublicID

	var key string

	key = "approvedApplications"
	if invite {
		key = "approvedInvitations"
	}

	if action == "deny" {
		key = "deniedApplications"
		if invite {
			key = "deniedInvitations"
		}
	}

	app.Metrics.IncrCounter(key, 1)
}
