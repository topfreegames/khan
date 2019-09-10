package caches

import (
	"fmt"
	"math/rand"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/khan/models"
)

// ClansSummariesCache represents a cache for the RetrieveClansSummaries operation
type ClansSummariesCache struct {
	Cache                 *gocache.Cache
	TimeToLive            time.Duration
	TimeToLiveRandomError time.Duration
}

// GetClansSummaries returns a summary of the clans details for a given list of clans by their game
// id and public ids
func (c *ClansSummariesCache) GetClansSummaries(db models.DB, gameID string, publicIDs []string) ([]map[string]interface{}, error) {
	resultMap := c.getCachedClansSummaries(gameID, publicIDs)
	err := c.getAndCacheClansSummaries(db, gameID, resultMap)
	if err != nil {
		if _, ok := err.(*models.CouldNotFindAllClansError); !ok {
			return nil, err
		}
	}
	var result []map[string]interface{}
	for _, publicID := range publicIDs {
		if resultMap[publicID] != nil {
			result = append(result, resultMap[publicID])
		}
	}
	return result, err
}

func (c *ClansSummariesCache) getClanSummaryCacheKey(gameID, publicID string) string {
	return fmt.Sprintf("%s/%s", gameID, publicID)
}

func (c *ClansSummariesCache) getClanSummaryCache(gameID, publicID string) map[string]interface{} {
	clan, present := c.Cache.Get(c.getClanSummaryCacheKey(gameID, publicID))
	if !present {
		return nil
	}
	return clan.(map[string]interface{})
}

func (c *ClansSummariesCache) setClanSummaryCache(gameID, publicID string, clanPayload map[string]interface{}) {
	ttl := c.TimeToLive - c.TimeToLiveRandomError
	ttl += time.Duration(rand.Intn(int(2*c.TimeToLiveRandomError + 1)))
	c.Cache.Set(c.getClanSummaryCacheKey(gameID, publicID), clanPayload, ttl)
}

func (c *ClansSummariesCache) getCachedClansSummaries(gameID string, publicIDs []string) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for _, publicID := range publicIDs {
		result[publicID] = c.getClanSummaryCache(gameID, publicID)
	}
	return result
}

func (c *ClansSummariesCache) getAndCacheClansSummaries(db models.DB, gameID string, resultMap map[string]map[string]interface{}) error {
	missingPublicIDs := c.getMissingPublicIDsFromResultMap(resultMap)
	if len(missingPublicIDs) == 0 {
		return nil
	}
	clans, err := models.GetClansSummaries(db, gameID, missingPublicIDs)
	if err != nil {
		if _, ok := err.(*models.CouldNotFindAllClansError); !ok {
			return err
		}
	}
	for _, clanPayload := range clans {
		publicID := clanPayload["publicID"].(string)
		resultMap[publicID] = clanPayload
		c.setClanSummaryCache(gameID, publicID, clanPayload)
	}
	return err
}

func (c *ClansSummariesCache) getMissingPublicIDsFromResultMap(resultMap map[string]map[string]interface{}) []string {
	var missingPublicIDs []string
	for publicID, clanPayload := range resultMap {
		if clanPayload == nil {
			missingPublicIDs = append(missingPublicIDs, publicID)
		}
	}
	return missingPublicIDs
}
