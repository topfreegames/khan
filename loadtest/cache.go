package loadtest

import (
	"math/rand"

	"github.com/spf13/viper"
	"github.com/topfreegames/khan/lib"
)

type (
	cache interface {
		loadInitialData(lib.KhanInterface) error
		getSharedClansCount() (int, error)
		chooseRandomSharedClanAndPlayer() (string, string, error)
		getFreePlayersCount() (int, error)
		chooseRandomFreePlayer() (string, error)
		getOwnerPlayersCount() (int, error)
		chooseRandomClan() (string, error)
		getNotFullClansCount() (int, error)
		chooseRandomNotFullClan() (string, error)
		getMemberPlayersCount() (int, error)
		chooseRandomMemberPlayerAndClan() (string, string, error)
		createPlayer(string) error
		createClan(string, string) error
		leaveClan(string, string, string) error
		transferClanOwnership(string, string, string) error
		applyForMembership(string, string) error
		deleteMembership(string, string) error
	}

	sharedClan struct {
		publicID         string
		membersPublicIDs []string
	}

	cacheImpl struct {
		gameMaxMembers int
		sharedClans    []sharedClan
		freePlayers    *UnorderedStringMap
		ownerPlayers   *UnorderedStringMap
		memberPlayers  *UnorderedStringMap
		fullClans      *UnorderedStringMap
		notFullClans   *UnorderedStringMap
	}
)

func newCacheImpl(gameMaxMembers int, sharedClansFile string) (*cacheImpl, error) {
	c := &cacheImpl{
		gameMaxMembers: gameMaxMembers,
		freePlayers:    NewUnorderedStringMap(),
		ownerPlayers:   NewUnorderedStringMap(),
		memberPlayers:  NewUnorderedStringMap(),
		fullClans:      NewUnorderedStringMap(),
		notFullClans:   NewUnorderedStringMap(),
	}

	// shared clans config
	sharedClansConfig := viper.New()
	sharedClansConfig.SetConfigType("yaml")
	sharedClansConfig.SetConfigFile(sharedClansFile)
	sharedClansConfig.AddConfigPath(".")
	if err := sharedClansConfig.ReadInConfig(); err != nil {
		return nil, err
	}
	for _, clanPublicID := range sharedClansConfig.GetStringSlice("clans") {
		c.sharedClans = append(c.sharedClans, sharedClan{
			publicID: clanPublicID,
		})
	}

	return c, nil
}

func (c *cacheImpl) loadInitialData(client lib.KhanInterface) error {
	if err := c.loadSharedClansMembers(client); err != nil {
		return err
	}
	return nil
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
	clanMembers, err := client.RetrieveClanMembers(nil, c.sharedClans[clanIdx].publicID)
	if err != nil {
		return err
	}
	if clanMembers == nil {
		return &GenericError{"NilPayloadError", "Operation retrieveClanMembers returned no error with nil payload."}
	}
	if len(clanMembers.Members) == 0 {
		return &GenericError{"EmptyClanError", "Operation retrieveClanMembers returned no error with empty clan."}
	}
	c.sharedClans[clanIdx].membersPublicIDs = clanMembers.Members
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

func (c *cacheImpl) getOwnerPlayersCount() (int, error) {
	return c.ownerPlayers.Len(), nil
}

func (c *cacheImpl) chooseRandomClan() (string, error) {
	count := c.ownerPlayers.Len()
	if count > 0 {
		clanPublicID, err := c.ownerPlayers.GetValue(rand.Intn(count))
		if err != nil {
			return "", err
		}
		return clanPublicID.(string), nil
	}
	return "", &GenericError{"NoClansError", "Cannot choose clan from empty set."}
}

func (c *cacheImpl) getNotFullClansCount() (int, error) {
	return c.notFullClans.Len(), nil
}

func (c *cacheImpl) chooseRandomNotFullClan() (string, error) {
	count := c.notFullClans.Len()
	if count > 0 {
		return c.notFullClans.GetKey(rand.Intn(count))
	}
	return "", &GenericError{"NoNotFullClansError", "Cannot choose not full clan from empty set."}
}

func (c *cacheImpl) getMemberPlayersCount() (int, error) {
	return c.memberPlayers.Len(), nil
}

func (c *cacheImpl) chooseRandomMemberPlayerAndClan() (string, string, error) {
	count := c.memberPlayers.Len()
	if count > 0 {
		idx := rand.Intn(count)
		playerPublicID, err := c.memberPlayers.GetKey(idx)
		if err != nil {
			return "", "", err
		}
		clanPublicID, err := c.memberPlayers.GetValue(idx)
		if err != nil {
			return "", "", err
		}
		return playerPublicID, clanPublicID.(string), nil
	}
	return "", "", &GenericError{"NoMemberPlayersError", "Cannot choose member player from empty set."}
}

func (c *cacheImpl) createPlayer(playerPublicID string) error {
	c.freePlayers.Set(playerPublicID, nil)
	return nil
}

func (c *cacheImpl) createClan(clanPublicID, playerPublicID string) error {
	c.freePlayers.Remove(playerPublicID)
	c.ownerPlayers.Set(playerPublicID, clanPublicID)
	c.incrementMembershipCount(clanPublicID)
	return nil
}

func (c *cacheImpl) leaveClan(clanPublicID, oldOnwerPublicID, newOwnerPublicID string) error {
	c.ownerPlayers.Remove(oldOnwerPublicID)
	c.freePlayers.Set(oldOnwerPublicID, nil)
	c.decrementMembershipCount(clanPublicID)
	if newOwnerPublicID != "" {
		c.memberPlayers.Remove(newOwnerPublicID)
		c.ownerPlayers.Set(newOwnerPublicID, clanPublicID)
	}
	return nil
}

func (c *cacheImpl) transferClanOwnership(clanPublicID, oldOnwerPublicID, newOwnerPublicID string) error {
	c.ownerPlayers.Remove(oldOnwerPublicID)
	c.memberPlayers.Set(oldOnwerPublicID, clanPublicID)
	c.memberPlayers.Remove(newOwnerPublicID)
	c.ownerPlayers.Set(newOwnerPublicID, clanPublicID)
	return nil
}

func (c *cacheImpl) applyForMembership(clanPublicID, playerPublicID string) error {
	c.freePlayers.Remove(playerPublicID)
	c.memberPlayers.Set(playerPublicID, clanPublicID)
	c.incrementMembershipCount(clanPublicID)
	return nil
}

func (c *cacheImpl) deleteMembership(clanPublicID, playerPublicID string) error {
	c.memberPlayers.Remove(playerPublicID)
	c.freePlayers.Set(playerPublicID, nil)
	c.decrementMembershipCount(clanPublicID)
	return nil
}

func (c *cacheImpl) incrementMembershipCount(clanPublicID string) {
	if !c.notFullClans.Has(clanPublicID) {
		c.notFullClans.Set(clanPublicID, 1)
	} else {
		count := c.notFullClans.Get(clanPublicID).(int) + 1
		if count < c.gameMaxMembers {
			c.notFullClans.Set(clanPublicID, count)
		} else {
			c.notFullClans.Remove(clanPublicID)
			c.fullClans.Set(clanPublicID, count)
		}
	}
}

func (c *cacheImpl) decrementMembershipCount(clanPublicID string) {
	if c.fullClans.Has(clanPublicID) {
		count := c.fullClans.Get(clanPublicID).(int) - 1
		c.fullClans.Remove(clanPublicID)
		c.notFullClans.Set(clanPublicID, count)
	} else {
		count := c.notFullClans.Get(clanPublicID).(int) - 1
		if count == 0 {
			c.notFullClans.Remove(clanPublicID)
		} else {
			c.notFullClans.Set(clanPublicID, count)
		}
	}
}
