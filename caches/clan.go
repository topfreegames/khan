package caches

import (
	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/khan/models"
)

// ClansSummariesCache represents a cache for the RetrieveClansSummaries operation
type ClansSummariesCache struct {
	cache *gocache.Cache
}

// NewClansSummariesCache returns a new instance of ClansSummariesCache
func NewClansSummariesCache(cache *gocache.Cache) *ClansSummariesCache {
	return &ClansSummariesCache{
		cache: cache,
	}
}

// GetClansSummaries returns a summary of the clans details for a given list of clans by their game
// id and public ids
func (c *ClansSummariesCache) GetClansSummaries(db models.DB, gameID string, publicIDs []string) ([]map[string]interface{}, error) {
	return models.GetClansSummaries(db, gameID, publicIDs)
}
