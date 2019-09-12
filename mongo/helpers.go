package mongo

import (
	"fmt"

	"github.com/globalsign/mgo/bson"
)

// GetClanNameTextIndexCommand returns a mongo command to create the clan names text index.
func GetClanNameTextIndexCommand(gameID string, background bool) bson.D {
	return bson.D{
		{Name: "createIndexes", Value: fmt.Sprintf("clans_%s", gameID)},
		{Name: "indexes", Value: []interface{}{
			bson.M{
				"key": bson.M{
					"name":         "text",
					"namePrefixes": "text",
				},
				"name":             fmt.Sprintf("clans_%s_name_text_namePrefixes_text_index", gameID),
				"background":       background,
				"default_language": "none",
			},
		}},
	}
}
