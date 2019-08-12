package loadtest

import (
	"github.com/topfreegames/khan/lib"
)

func (app *App) setMembershipConfigurationDefaults() {
}

func (app *App) configureMembershipOperations() {
	app.appendOperation(app.getApplyForMembershipOperation())
}

func (app *App) getApplyForMembershipOperation() operation {
	operationKey := "applyForMembership"
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

			_, err = app.client.ApplyForMembership(nil, &lib.ApplicationPayload{
				ClanID:         clanPublicID,
				Message:        "",
				Level:          app.config.GetString("loadtest.game.membershipLevel"),
				PlayerPublicID: playerPublicID,
			})
			return app.cache.applyForMembership(clanPublicID, playerPublicID)
		},
	}
}
