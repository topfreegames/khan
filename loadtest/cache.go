package loadtest

import (
	"github.com/spf13/viper"
)

type (
	cache interface {
		getSharedClansCount() (int, error)
		getSharedClanByPublicID(int) (string, error)
		getSharedClans() []string
	}

	cacheImpl struct {
		sharedClans []string
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
	return &cacheImpl{
		sharedClans: sharedClansConfig.GetStringSlice("clans"),
	}, nil
}

func (c *cacheImpl) getSharedClansCount() (int, error) {
	return len(c.sharedClans), nil
}

func (c *cacheImpl) getSharedClanByPublicID(idx int) (string, error) {
	return c.sharedClans[idx], nil
}

func (c *cacheImpl) getSharedClans() []string {
	return c.sharedClans
}
