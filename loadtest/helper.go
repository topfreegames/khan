package loadtest

import (
	"fmt"
	"math/rand"

	uuid "github.com/satori/go.uuid"
)

func getRandomScore() int {
	return rand.Intn(1000)
}

func getRandomPlayerName() string {
	return fmt.Sprintf("PlayerName-%s", uuid.NewV4().String()[:8])
}

func getRandomClanName() string {
	return fmt.Sprintf("ClanName-%s", uuid.NewV4().String()[:8])
}

func getRandomPublicID() string {
	return uuid.NewV4().String()
}

func getScoreFromMetadata(metadata interface{}) int {
	if metadata != nil {
		metadataMap := metadata.(map[string]interface{})
		if score, ok := metadataMap["score"]; ok {
			return int(score.(float64))
		}
	}
	return 0
}

func getMetadataWithRandomScore() map[string]interface{} {
	return map[string]interface{}{
		"score": getRandomScore(),
	}
}
