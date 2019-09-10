package caches

import (
	"fmt"
	"math/rand"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/khan/models"
)

// ClansSummaries represents a cache for the RetrieveClansSummaries operation.
type ClansSummaries struct {
	// Cache points to an instance of gocache.Cache used as the backend cache object.
	Cache *gocache.Cache

	// The cache for one clan will live by a random amount of time within [TTL - TTLRandomError, TTL + TTLRandomError].
	TTL            time.Duration
	TTLRandomError time.Duration
}

// GetClansSummaries is a cache in front of models.GetClansSummaries() with the exact same interface.
// The map[string]interface{} return type represents a summary of one clan with the following keys/values:
// "membershipCount":  int
// "publicID":         string
// "metadata":         map[string]interface{} (user-defined arbitrary JSON object with clan metadata)
// "name":             string
// "allowApplication": bool
// "autoJoin":         bool
// TODO: replace this map with a richer type
func (c *ClansSummaries) GetClansSummaries(db models.DB, gameID string, publicIDs []string) ([]map[string]interface{}, error) {
	resultMap := c.getCachedClansSummaries(gameID, publicIDs)
	err := c.getAndCacheClansSummaries(db, gameID, resultMap)
	if err != nil {
		if _, ok := err.(*models.CouldNotFindAllClansError); !ok {
			return nil, err
		}
	}
	var result []map[string]interface{}
	for _, publicID := range publicIDs {
		if summary := resultMap[publicID]; summary != nil {
			result = append(result, summary)
		}
	}
	return result, err
}

func (c *ClansSummaries) getClanSummaryCacheKey(gameID, publicID string) string {
	return fmt.Sprintf("%s/%s", gameID, publicID)
}

func (c *ClansSummaries) getClanSummaryCache(gameID, publicID string) map[string]interface{} {
	clan, present := c.Cache.Get(c.getClanSummaryCacheKey(gameID, publicID))
	if !present {
		return nil
	}
	return clan.(map[string]interface{})
}

func (c *ClansSummaries) setClanSummaryCache(gameID, publicID string, clanPayload map[string]interface{}) {
	ttl := c.TTL - c.TTLRandomError
	ttl += time.Duration(rand.Intn(int(2*c.TTLRandomError + 1)))
	c.Cache.Set(c.getClanSummaryCacheKey(gameID, publicID), clanPayload, ttl)
}

func (c *ClansSummaries) getCachedClansSummaries(gameID string, publicIDs []string) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for _, publicID := range publicIDs {
		result[publicID] = c.getClanSummaryCache(gameID, publicID)
	}
	return result
}

func (c *ClansSummaries) getAndCacheClansSummaries(db models.DB, gameID string, resultMap map[string]map[string]interface{}) error {
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

func (c *ClansSummaries) getMissingPublicIDsFromResultMap(resultMap map[string]map[string]interface{}) []string {
	var missingPublicIDs []string
	for publicID, clanPayload := range resultMap {
		if clanPayload == nil {
			missingPublicIDs = append(missingPublicIDs, publicID)
		}
	}
	return missingPublicIDs
}
