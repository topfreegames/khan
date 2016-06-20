// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"

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
		errors = append(errors, "maxMembershipLevel should be greater or equal to minMembershipLevel")
	}
	if payload.MinLevelToAcceptApplication < payload.MinMembershipLevel {
		errors = append(errors, "minLevelToAcceptApplication should be greater or equal to minMembershipLevel")
	}
	if payload.MinLevelToCreateInvitation < payload.MinMembershipLevel {
		errors = append(errors, "minLevelToCreateInvitation should be greater or equal to minMembershipLevel")
	}
	if payload.MinLevelToRemoveMember < payload.MinMembershipLevel {
		errors = append(errors, "minLevelToRemoveMember should be greater or equal to minMembershipLevel")
	}
	return errors
}

// CreateGameHandler is the handler responsible for creating new games
func CreateGameHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload gamePayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateGamePayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
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
			"publicID": game.PublicID,
		}, c)
	}
}

// UpdateGameHandler is the handler responsible for updating existing
func UpdateGameHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		var payload gamePayload

		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateGamePayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
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
