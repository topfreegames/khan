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
)

type applyForMembershipPayload struct {
	Level          int
	PlayerPublicID string
}

type inviteForMembershipPayload struct {
	Level             int
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

//ApplyForMembershipHandler is the handler responsible for applying for new memberships
func ApplyForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload applyForMembershipPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.CreateMembership(
			db,
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

		SucceedWith(map[string]interface{}{}, c)
	}
}

//InviteForMembershipHandler is the handler responsible for creating new memberships
func InviteForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload inviteForMembershipPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.CreateMembership(
			db,
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

		SucceedWith(map[string]interface{}{}, c)

	}
}

//ApproveOrDenyMembershipApplicationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipApplicationHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		action := c.Param("action")
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.ApproveOrDenyMembershipApplication(
			db,
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

		SucceedWith(map[string]interface{}{}, c)
	}
}

//ApproveOrDenyMembershipInvitationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipInvitationHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		action := c.Param("action")
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload approveOrDenyMembershipInvitationPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.ApproveOrDenyMembershipInvitation(
			db,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			action,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

//DeleteMembershipHandler is the handler responsible for deleting a member
func DeleteMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		err := models.DeleteMembership(
			db,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

//PromoteOrDemoteMembershipHandler is the handler responsible for promoting or demoting a member
func PromoteOrDemoteMembershipHandler(app *App, action string) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		membership, err := models.PromoteOrDemoteMember(
			db,
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

		SucceedWith(map[string]interface{}{
			"level": membership.Level,
		}, c)
	}
}
