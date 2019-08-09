package loadtest

import (
	"math/rand"

	"github.com/topfreegames/khan/lib"
)

func (app *App) setClanConfigurationDefaults() {
}

func (app *App) configureClanOperations() {
	app.appendOperation(app.getUpdateSharedClanScoreOperation())
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
			playerMetadata := make(map[string]interface{})
			playerMetadata["score"] = int(rand.Float64() * 1000)
			_, err = app.client.UpdatePlayer(nil, playerPublicID, "PlayerName", playerMetadata)
			if err != nil {
				return err
			}

			// getClan
			clan, err := app.client.RetrieveClan(nil, clanPublicID)
			if err != nil {
				return err
			}

			// updateClan
			clanScore := getPlayerScoreFromMetadata(clan.Owner.Metadata)
			for _, member := range clan.Roster {
				clanScore += getPlayerScoreFromMetadata(member.Player.Metadata)
			}
			clanMetadata := make(map[string]interface{})
			clanMetadata["score"] = clanScore
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

func getPlayerScoreFromMetadata(metadata interface{}) int {
	if metadata != nil {
		metadataMap := metadata.(map[string]interface{})
		if score, ok := metadataMap["score"]; ok {
			return int(score.(float64))
		}
	}
	return 0
}
