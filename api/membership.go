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

//applyForMembershipPayload maps the payload for the Apply For Membership route
type applyForMembershipPayload struct {
	Level          int
	PlayerPublicID string
}

//inviteForMembershipPayload maps the payload for the Invite For Membership route
type inviteForMembershipPayload struct {
	Level             int
	PlayerPublicID    string
	RequestorPublicID string
}

//ApplyForMembershipHandler is the handler responsible for applying for new memberships
func ApplyForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Get("gameID").(string)
		clanPublicID := c.Get("clanPublicID").(string)

		var payload applyForMembershipPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		membership, err := models.CreateMembership(
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

		SucceedWith(map[string]interface{}{
			"id": membership.ID,
		}, c)
	}
}

//InviteForMembershipHandler is the handler responsible for creating new memberships
func InviteForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Get("gameID").(string)
		clanPublicID := c.Get("clanPublicID").(string)

		var payload inviteForMembershipPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		membership, err := models.CreateMembership(
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

		SucceedWith(map[string]interface{}{
			"id": membership.ID,
		}, c)
	}
}

//SetMembershipHandlersGroup configures the routes for all membership related routes
func SetMembershipHandlersGroup(app *App) {
	gameParty := app.App.Party("/games/:gameID", func(c *iris.Context) {
		gameID := c.Param("gameID")
		c.Set("gameID", gameID)
		c.Next()
	})
	clanParty := gameParty.Party("/clans/:clanPublicID", func(c *iris.Context) {
		clanPublicID := c.Param("clanPublicID")
		c.Set("clanPublicID", clanPublicID)
		c.Next()
	})
	membershipHandlersGroup := clanParty.Party("/memberships", func(c *iris.Context) {
		c.Next()
	})

	membershipHandlersGroup.Post("/apply", ApplyForMembershipHandler(app))
	membershipHandlersGroup.Post("/invite", InviteForMembershipHandler(app))
}
