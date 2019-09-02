package loadtest

import (
	"fmt"
	"math/rand"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func getRandomScore() int {
	return rand.Intn(1000)
}

func getRandomPlayerName() string {
	return fmt.Sprintf("PlayerName-%s", uuid.NewV4().String()[:8])
}

func getRandomClanName(numberOfWords int) string {
	pieces := []string{}
	for i := 0; i < numberOfWords; i++ {
		pieces = append(pieces, dictionary[rand.Int()%len(dictionary)])
	}
	return strings.Join(pieces, " ")
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
