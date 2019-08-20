package testing

import (
	"fmt"

	"github.com/globalsign/mgo/bson"
	"github.com/topfreegames/extensions/mongo/interfaces"
)

// CreateClanNameTextIndexInMongo creates the necessary text index for clan search in mongo
func CreateClanNameTextIndexInMongo(getTestMongo func() (interfaces.MongoDB, error), gameID string) error {
	mongo, err := getTestMongo()
	if err != nil {
		return err
	}

	cmd := bson.D{
		{Name: "createIndexes", Value: fmt.Sprintf("clans_%s", gameID)},
		{Name: "indexes", Value: []interface{}{
			bson.M{
				"key": bson.M{
					"name": "text",
				},
				"name": fmt.Sprintf("clans_%s_name_text_index", gameID),
			},
		}},
	}
	return mongo.Run(cmd, nil)
}
