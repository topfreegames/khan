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

type gamePayload struct {
	PublicID                      string
	Name                          string
	Metadata                      string
	MinMembershipLevel            int
	MaxMembershipLevel            int
	MinLevelToAcceptApplication   int
	MinLevelToCreateInvitation    int
	MinLevelToRemoveMember        int
	MinLevelOffsetToPromoteMember int
	MinLevelOffsetToDemoteMember  int
	MaxMembers                    int
}

func validateGamePayload(payload gamePayload) []string {
	var errors []string
	if payload.MaxMembershipLevel < payload.MinMembershipLevel {
		errors = append(errors, "MaxMembershipLevel should be greater or equal to MinMembershipLevel")
	}
	if payload.MinLevelToAcceptApplication < payload.MinMembershipLevel {
		errors = append(errors, "MinLevelToAcceptApplication should be greater or equal to MinMembershipLevel")
	}
	if payload.MinLevelToCreateInvitation < payload.MinMembershipLevel {
		errors = append(errors, "MinLevelToCreateInvitation should be greater or equal to MinMembershipLevel")
	}
	if payload.MinLevelToRemoveMember < payload.MinMembershipLevel {
		errors = append(errors, "MinLevelToRemoveMember should be greater or equal to MinMembershipLevel")
	}
	return errors
}

//CreateGameHandler is the handler responsible for creating new games
func CreateGameHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload gamePayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateGamePayload(payload); len(payloadErrors) != 0 {
			FailWithJSON(422, map[string]interface{}{"reason": payloadErrors}, c)
			return
		}

		db := GetCtxDB(c)

		game, err := models.CreateGame(
			db,
			payload.PublicID,
			payload.Name,
			payload.Metadata,
			payload.MinMembershipLevel,
			payload.MaxMembershipLevel,
			payload.MinLevelToRemoveMember,
			payload.MinLevelToCreateInvitation,
			payload.MinLevelToRemoveMember,
			payload.MinLevelOffsetToPromoteMember,
			payload.MinLevelOffsetToDemoteMember,
			payload.MaxMembers,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"id": game.ID,
		}, c)
	}
}

//UpdateGameHandler is the handler responsible for updating existing
func UpdateGameHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		var payload gamePayload

		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateGamePayload(payload); len(payloadErrors) != 0 {
			FailWithJSON(422, map[string]interface{}{"reason": payloadErrors}, c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.UpdateGame(
			db,
			gameID,
			payload.Name,
			payload.Metadata,
			payload.MinMembershipLevel,
			payload.MaxMembershipLevel,
			payload.MinLevelToAcceptApplication,
			payload.MinLevelToCreateInvitation,
			payload.MinLevelToRemoveMember,
			payload.MinLevelOffsetToPromoteMember,
			payload.MinLevelOffsetToDemoteMember,
			payload.MaxMembers,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}
