package loadtest

import (
	"github.com/topfreegames/khan/lib"
)

func (app *App) setClanConfigurationDefaults() {
}

func (app *App) configureClanOperations() {
	app.appendOperation(app.getUpdateSharedClanScoreOperation())
	app.appendOperation(app.getCreateClanOperation())
	app.appendOperation(app.getLeaveClanOperation())
}

func (app *App) getUpdateSharedClanScoreOperation() operation {
	operationKey := "updateSharedClanScore"
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getSharedClansCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			err := app.loadSharedClansMembers() // cached
			if err != nil {
				return err
			}

			clanPublicID, playerPublicID, err := app.cache.chooseRandomSharedClanAndPlayer()
			if err != nil {
				return err
			}

			// updatePlayer
			_, err = app.client.UpdatePlayer(nil, playerPublicID, getRandomPlayerName(), getMetadataWithRandomScore())
			if err != nil {
				return err
			}

			// getClan
			clan, err := app.client.RetrieveClan(nil, clanPublicID)
			if err != nil {
				return err
			}

			// updateClan
			clanScore := getScoreFromMetadata(clan.Owner.Metadata)
			for _, member := range clan.Roster {
				clanScore += getScoreFromMetadata(member.Player.Metadata)
			}
			clanMetadata := map[string]interface{}{
				"score": clanScore,
			}
			_, err = app.client.UpdateClan(nil, &lib.ClanPayload{
				PublicID:         clan.PublicID,
				Name:             clan.Name,
				OwnerPublicID:    clan.Owner.PublicID,
				Metadata:         clanMetadata,
				AllowApplication: clan.AllowApplication,
				AutoJoin:         clan.AutoJoin,
			})
			return err
		},
	}
}

func (app *App) getCreateClanOperation() operation {
	operationKey := "createClan"
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			count, err := app.cache.getFreePlayersCount()
			if err != nil {
				return false, err
			}
			return count > 0, nil
		},
		execute: func() error {
			playerPublicID, err := app.cache.chooseRandomFreePlayer()
			if err != nil {
				return err
			}

			clanPublicID := getRandomPublicID()
			_, err = app.client.CreateClan(nil, &lib.ClanPayload{
				PublicID:         clanPublicID,
				Name:             getRandomClanName(),
				OwnerPublicID:    playerPublicID,
				Metadata:         getMetadataWithRandomScore(),
				AllowApplication: getRandomBool(),
				AutoJoin:         getRandomBool(),
			})
			if err != nil {
				return err
			}
			return app.cache.bindPlayer(playerPublicID, clanPublicID)
		},
	}
}

func (app *App) getLeaveClanOperation() operation {
	operationKey := "leaveClan"
	return operation{
		probability: app.getOperationProbabilityConfig(operationKey),
		canExecute: func() (bool, error) {
			return true, nil
		},
		execute: func() error {
			return nil
		},
	}
}
