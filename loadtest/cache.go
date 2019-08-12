package loadtest

import (
	"math/rand"

	"github.com/spf13/viper"
	"github.com/topfreegames/khan/lib"
)

type (
	cache interface {
		loadSharedClansMembers(lib.KhanInterface) error
		getSharedClansCount() (int, error)
		chooseRandomSharedClanAndPlayer() (string, string, error)
		getFreePlayersCount() (int, error)
		chooseRandomFreePlayer() (string, error)
		addFreePlayer(string) error
		bindPlayer(string, string) error
	}

	sharedClan struct {
		publicID         string
		membersPublicIDs []string
	}

	cacheImpl struct {
		sharedClans  []sharedClan
		freePlayers  *UnorderedStringMap
		boundPlayers *UnorderedStringMap
	}
)

func getCacheImpl(config *viper.Viper, sharedClansFile string) (*cacheImpl, error) {
	sharedClansConfig := viper.New()
	sharedClansConfig.SetConfigType("yaml")
	sharedClansConfig.SetConfigFile(sharedClansFile)
	sharedClansConfig.AddConfigPath(".")
	if err := sharedClansConfig.ReadInConfig(); err != nil {
		return nil, err
	}
	c := &cacheImpl{
		freePlayers:  NewUnorderedStringMap(),
		boundPlayers: NewUnorderedStringMap(),
	}
	for _, clanPublicID := range sharedClansConfig.GetStringSlice("clans") {
		c.sharedClans = append(c.sharedClans, sharedClan{
			publicID: clanPublicID,
		})
	}
	return c, nil
}

func (c *cacheImpl) loadSharedClansMembers(client lib.KhanInterface) error {
	if len(c.sharedClans) == 0 || len(c.sharedClans[0].membersPublicIDs) > 0 {
		return nil
	}
	for i := range c.sharedClans {
		if err := c.loadSharedClanMembers(client, i); err != nil {
			return err
		}
	}
	return nil
}

func (c *cacheImpl) loadSharedClanMembers(client lib.KhanInterface, clanIdx int) error {
	membersPayload, err := client.RetrieveClanMembers(nil, c.sharedClans[clanIdx].publicID)
	if err != nil {
		return err
	}
	c.sharedClans[clanIdx].membersPublicIDs = membersPayload.Members
	return nil
}

func (c *cacheImpl) getSharedClansCount() (int, error) {
	return len(c.sharedClans), nil
}

func (c *cacheImpl) chooseRandomSharedClanAndPlayer() (string, string, error) {
	clanIdx := rand.Intn(len(c.sharedClans))
	playerIdx := rand.Intn(len(c.sharedClans[clanIdx].membersPublicIDs))
	return c.sharedClans[clanIdx].publicID, c.sharedClans[clanIdx].membersPublicIDs[playerIdx], nil
}

func (c *cacheImpl) getFreePlayersCount() (int, error) {
	return c.freePlayers.Len(), nil
}

func (c *cacheImpl) chooseRandomFreePlayer() (string, error) {
	count := c.freePlayers.Len()
	if count > 0 {
		return c.freePlayers.GetKey(rand.Intn(count))
	}
	return "", &GenericError{"NoFreePlayersError", "Cannot choose free player from empty set."}
}

func (c *cacheImpl) addFreePlayer(playerPublicID string) error {
	c.freePlayers.Set(playerPublicID, nil)
	return nil
}

func (c *cacheImpl) bindPlayer(playerPublicID, clanPublicID string) error {
	c.freePlayers.Remove(playerPublicID)
	c.boundPlayers.Set(playerPublicID, nil)
	return nil
}
