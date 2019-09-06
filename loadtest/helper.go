package loadtest

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var dictionary []string

// LoadRandomWords loads random words into memory for random clan name generation
func LoadRandomWords() {
	if dictionary != nil {
		return
	}
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	dictionary = strings.Split(string(bytes), "\n")
}

func getRandomScore() int {
	return rand.Intn(1000)
}

func getRandomPlayerName() string {
	return fmt.Sprintf("PlayerName-%s", uuid.NewV4().String()[:8])
}

func getRandomClanName(numberOfWords int) string {
	if dictionary == nil {
		return fmt.Sprintf("ClanName-%s", uuid.NewV4().String()[:8])
	}
	pieces := []string{}
	for i := 0; i < numberOfWords; i++ {
		pieces = append(pieces, dictionary[rand.Intn(len(dictionary))])
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
