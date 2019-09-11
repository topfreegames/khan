package caches

import (
	"fmt"

	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/khan/models"
)

// ClansSummaries represents a cache for the RetrieveClansSummaries operation.
type ClansSummaries struct {
	// Cache points to an instance of gocache.Cache used as the backend cache object.
	Cache *gocache.Cache
}

// GetClansSummaries is a cache in front of models.GetClansSummaries() with the exact same interface.
// Like models.GetClansSummaries(), this function may return partial results + CouldNotFindAllClansError.
// The map[string]interface{} return type represents a summary of one clan with the following keys/values:
// "membershipCount":  int
// "publicID":         string
// "metadata":         map[string]interface{} (user-defined arbitrary JSON object with clan metadata)
// "name":             string
// "allowApplication": bool
// "autoJoin":         bool
// TODO(matheuscscp): replace this map with a richer type
func (c *ClansSummaries) GetClansSummaries(db models.DB, gameID string, publicIDs []string) ([]map[string]interface{}, error) {
	// first, assemble a result map with cached payloads. also assemble a missingPublicIDs string slice
	idToPayload := make(map[string]map[string]interface{})
	var missingPublicIDs []string
	for _, publicID := range publicIDs {
		if clanPayload, present := c.Cache.Get(c.getClanSummaryCacheKey(gameID, publicID)); present {
			idToPayload[publicID] = clanPayload.(map[string]interface{})
		} else {
			missingPublicIDs = append(missingPublicIDs, publicID)
		}
	}

	// fetch and cache missing clans
	var err error
	if len(missingPublicIDs) > 0 {
		// fetch
		var clans []map[string]interface{}
		clans, err = models.GetClansSummaries(db, gameID, missingPublicIDs)
		if err != nil {
			if _, ok := err.(*models.CouldNotFindAllClansError); !ok {
				return nil, err
			}
		}

		// cache
		for _, clanPayload := range clans {
			publicID := clanPayload["publicID"].(string)
			idToPayload[publicID] = clanPayload
			c.Cache.Set(c.getClanSummaryCacheKey(gameID, publicID), clanPayload, gocache.DefaultExpiration)
		}
	}

	// assemble final result with input order
	var result []map[string]interface{}
	for _, publicID := range publicIDs {
		if summary, ok := idToPayload[publicID]; ok {
			result = append(result, summary)
		}
	}
	return result, err
}

func (c *ClansSummaries) getClanSummaryCacheKey(gameID, publicID string) string {
	return fmt.Sprintf("%s/%s", gameID, publicID)
}
