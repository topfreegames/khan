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
	"github.com/topfreegames/khan/util"
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

func dispatchMembershipHook(app *App, db models.DB, hookType int, gameID string, clanID, playerID, requestorID int) error {
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

	clanJSON := clan.Serialize()
	delete(clanJSON, "gameID")

	playerJSON := player.Serialize()
	delete(playerJSON, "gameID")

	requestorJSON := requestor.Serialize()
	delete(requestorJSON, "gameID")

	result := util.JSON{
		"gameID":    gameID,
		"clan":      clanJSON,
		"player":    playerJSON,
		"requestor": requestorJSON,
	}
	app.DispatchHooks(gameID, hookType, result)

	return nil
}

// ApplyForMembershipHandler is the handler responsible for applying for new memberships
func ApplyForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload applyForMembershipPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		membership, err := models.CreateMembership(
			db,
			game,
			gameID,
			payload.Level,
			payload.PlayerPublicID,
			clanPublicID,
			payload.PlayerPublicID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchMembershipHook(
			app, db, models.MembershipApplicationCreatedHook,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.RequestorID,
		)
		if err != nil {
			FailWith(500, err.Error(), c)
		}

		SucceedWith(util.JSON{}, c)
	}
}

// InviteForMembershipHandler is the handler responsible for creating new memberships
func InviteForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload inviteForMembershipPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		membership, err := models.CreateMembership(
			db,
			game,
			gameID,
			payload.Level,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchMembershipHook(
			app, db, models.MembershipApplicationCreatedHook,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.RequestorID,
		)
		if err != nil {
			FailWith(500, err.Error(), c)
		}

		SucceedWith(util.JSON{}, c)
	}
}

// ApproveOrDenyMembershipApplicationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipApplicationHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		action := c.Param("action")
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		membership, err := models.ApproveOrDenyMembershipApplication(
			db,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			action,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		//TODO: This must be removed once the membership includes who approved it
		requestor, err := models.GetPlayerByPublicID(db, gameID, payload.RequestorPublicID)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		hookType := models.MembershipApprovedHook
		if action == "deny" {
			hookType = models.MembershipDeniedHook
		}

		err = dispatchMembershipHook(
			app, db, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			requestor.ID,
		)
		if err != nil {
			FailWith(500, err.Error(), c)
		}

		SucceedWith(util.JSON{}, c)
	}
}

// ApproveOrDenyMembershipInvitationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipInvitationHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		action := c.Param("action")
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload approveOrDenyMembershipInvitationPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		membership, err := models.ApproveOrDenyMembershipInvitation(
			db,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			action,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		hookType := models.MembershipApprovedHook
		if action == "deny" {
			hookType = models.MembershipDeniedHook
		}

		err = dispatchMembershipHook(
			app, db, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.PlayerID,
		)
		if err != nil {
			FailWith(500, err.Error(), c)
		}

		SucceedWith(util.JSON{}, c)
	}
}

// DeleteMembershipHandler is the handler responsible for deleting a member
func DeleteMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		err = models.DeleteMembership(
			db,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(util.JSON{}, c)
	}
}

// PromoteOrDemoteMembershipHandler is the handler responsible for promoting or demoting a member
func PromoteOrDemoteMembershipHandler(app *App, action string) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		membership, err := models.PromoteOrDemoteMember(
			db,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			action,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(util.JSON{
			"level": membership.Level,
		}, c)
	}
}
