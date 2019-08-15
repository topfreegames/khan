package loadtest

import (
	"github.com/topfreegames/khan/lib"
)

func (app *App) configureMembershipOperations() {
	app.appendOperation(app.getApplyForMembershipOperation())
	app.appendOperation(app.getSelfDeleteMembershipOperation())
}

func (app *App) getApplyForMembershipOperation() operation {
	operationKey := "applyForMembership"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	membershipLevel := app.config.GetString("loadtest.game.membershipLevel")
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getFreePlayersCount()
			if err != nil {
				return false, err
			}
			if count == 0 {
				return false, nil
			}
			count, err = app.cache.getNotFullClansCount()
			return count > 0, err
		},
		execute: func() error {
			playerPublicID, err := app.cache.chooseRandomFreePlayer()
			if err != nil {
				return err
			}

			clanPublicID, err := app.cache.chooseRandomNotFullClan()
			if err != nil {
				return err
			}

			clanApplyResult, err := app.client.ApplyForMembership(nil, &lib.ApplicationPayload{
				ClanID:         clanPublicID,
				Message:        "",
				Level:          membershipLevel,
				PlayerPublicID: playerPublicID,
			})
			if err != nil {
				return err
			}
			if clanApplyResult == nil {
				return &GenericError{"NilPayloadError", "Operation applyForMembership returned no error with nil payload."}
			}
			if !clanApplyResult.Success {
				return &GenericError{"FailurePayloadError", "Operation applyForMembership returned no error with failure payload."}
			}
			if !clanApplyResult.Approved {
				return &GenericError{"NotApprovedError", "Operation applyForMembership returned no error with not approved membership."}
			}

			return app.cache.applyForMembership(clanPublicID, playerPublicID)
		},
	}
}

func (app *App) getSelfDeleteMembershipOperation() operation {
	operationKey := "selfDeleteMembership"
	app.setOperationProbabilityConfigDefault(operationKey, 1)
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getMemberPlayersCount()
			if err != nil {
				return false, nil
			}
			return count > 0, nil
		},
		execute: func() error {
			playerPublicID, clanPublicID, err := app.cache.chooseRandomMemberPlayerAndClan()
			if err != nil {
				return err
			}

			result, err := app.client.DeleteMembership(nil, &lib.DeleteMembershipPayload{
				ClanID:            clanPublicID,
				PlayerPublicID:    playerPublicID,
				RequestorPublicID: playerPublicID,
			})
			if err != nil {
				return err
			}
			if result == nil {
				return &GenericError{"NilPayloadError", "Operation deleteMembership returned no error with nil payload."}
			}
			if !result.Success {
				return &GenericError{"FailurePayloadError", "Operation deleteMembership returned no error with failure payload."}
			}

			return app.cache.deleteMembership(clanPublicID, playerPublicID)
		},
	}
}
