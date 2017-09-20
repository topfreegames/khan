// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

//go:generate easyjson -no_std_marshalers $GOFILE

package models

import (
	"fmt"
	"os"
	"strings"

	workers "github.com/jrallison/go-workers"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"github.com/topfreegames/khan/es"
	"github.com/topfreegames/khan/mongo"
	"github.com/topfreegames/khan/queues"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
	"gopkg.in/gorp.v1"
)

// ClanByName allows sorting clans by name
type ClanByName []*Clan

func (a ClanByName) Len() int           { return len(a) }
func (a ClanByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ClanByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// Clan identifies uniquely one clan in a given game
//easyjson:json
type Clan struct {
	ID               int                    `db:"id" json:"id" bson:"id"`
	GameID           string                 `db:"game_id" json:"gameId" bson:"gameId"`
	PublicID         string                 `db:"public_id" json:"publicId" bson:"publicId"`
	Name             string                 `db:"name" json:"name" bson:"name"`
	OwnerID          int                    `db:"owner_id" json:"ownerId" bson:"ownerId"`
	MembershipCount  int                    `db:"membership_count" json:"membershipCount" bson:"membershipCount"`
	Metadata         map[string]interface{} `db:"metadata" json:"metadata" bson:"metadata"`
	AllowApplication bool                   `db:"allow_application" json:"allowApplication" bson:"allowApplication"`
	AutoJoin         bool                   `db:"auto_join"  json:"autoJoin" bson:"autoJoin"`
	CreatedAt        int64                  `db:"created_at" json:"createdAt" bson:"createdAt"`
	UpdatedAt        int64                  `db:"updated_at" json:"updatedAt" bson:"updatedAt"`
	DeletedAt        int64                  `db:"deleted_at" json:"deletedAt" bson:"deletedAt"`
}

//ToJSON returns the clan as JSON
func (c *Clan) ToJSON() ([]byte, error) {
	w := jwriter.Writer{}
	c.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

//GetClanFromJSON unmarshals the clan from the specified JSON
func GetClanFromJSON(data []byte) (*Clan, error) {
	clan := &Clan{}
	l := jlexer.Lexer{Data: data}
	clan.UnmarshalEasyJSON(&l)
	return clan, l.Error()
}

//PreInsert populates fields before inserting a new clan
func (c *Clan) PreInsert(s gorp.SqlExecutor) error {
	c.CreatedAt = util.NowMilli()
	c.UpdatedAt = c.CreatedAt
	return nil
}

//PostInsert indexes clan in ES after creation in PG
func (c *Clan) PostInsert(s gorp.SqlExecutor) error {
	err := c.IndexClanIntoElasticSearch()
	if err != nil {
		return err
	}
	err = c.UpdateClanIntoMongoDB()
	return err
}

//PreUpdate populates fields before updating a clan
func (c *Clan) PreUpdate(s gorp.SqlExecutor) error {
	c.UpdatedAt = util.NowMilli()
	return nil
}

//PostUpdate indexes clan in ES after update in PG
func (c *Clan) PostUpdate(s gorp.SqlExecutor) error {
	err := c.UpdateClanIntoElasticSearch()
	if err != nil {
		return err
	}
	err = c.UpdateClanIntoMongoDB()
	return err
}

//PostDelete deletes clan from elasticsearch after deleting from PG
func (c *Clan) PostDelete(s gorp.SqlExecutor) error {
	err := c.DeleteClanFromElasticSearch()
	if err != nil {
		return err
	}
	err = c.DeleteClanFromMongoDB()
	return err
}

func getNewLogger() zap.Logger {
	ll := zap.InfoLevel
	if os.Getenv("SKIP_ELASTIC_LOG") == "true" {
		ll = zap.FatalLevel
	}

	return zap.New(
		zap.NewJSONEncoder(), // drop timestamps in tests
		ll,
	)
}

//IndexClanIntoElasticSearch after operation in PG
func (c *Clan) IndexClanIntoElasticSearch() error {
	es := es.GetConfiguredClient()
	if es != nil {
		workers.Enqueue(queues.KhanESQueue, "Add", map[string]interface{}{
			"index":  es.GetIndexName(c.GameID),
			"op":     "index",
			"clan":   c,
			"clanID": c.PublicID,
		})
	}
	return nil
}

// UpdateClanIntoMongoDB after operation in PG
func (c *Clan) UpdateClanIntoMongoDB() error {
	mongo, _ := mongo.GetConfiguredMongoClient()
	if mongo != nil {
		workers.Enqueue(queues.KhanMongoQueue, "Add", map[string]interface{}{
			"game":   c.GameID,
			"op":     "update",
			"clan":   c,
			"clanID": c.PublicID,
		})
	}
	return nil
}

//DeleteClanFromMongoDB after deletion in PG
func (c *Clan) DeleteClanFromMongoDB() error {
	mongo, _ := mongo.GetConfiguredMongoClient()
	if mongo != nil {
		workers.Enqueue(queues.KhanMongoQueue, "Add", map[string]interface{}{
			"game":   c.GameID,
			"op":     "delete",
			"clan":   c,
			"clanID": c.PublicID,
		})
	}
	return nil
}

//UpdateClanIntoElasticSearch after operation in PG
func (c *Clan) UpdateClanIntoElasticSearch() error {
	es := es.GetConfiguredClient()
	if es != nil {
		workers.Enqueue(queues.KhanESQueue, "Add", map[string]interface{}{
			"index":  es.GetIndexName(c.GameID),
			"op":     "update",
			"clan":   c,
			"clanID": c.PublicID,
		})
	}
	return nil
}

//DeleteClanFromElasticSearch after deletion in PG
func (c *Clan) DeleteClanFromElasticSearch() error {
	es := es.GetConfiguredClient()
	if es != nil {
		workers.Enqueue(queues.KhanESQueue, "Add", map[string]interface{}{
			"index":  es.GetIndexName(c.GameID),
			"op":     "delete",
			"clan":   c,
			"clanID": c.PublicID,
		})
	}
	return nil
}

func updateClanIntoES(db DB, id int) error {
	clan, err := GetClanByID(db, id)
	if err != nil {
		return err
	}
	if clan == nil {
		return &ModelNotFoundError{"Clan", id}
	}
	return clan.UpdateClanIntoElasticSearch()
}

func updateClanIntoMongo(db DB, id int) error {
	clan, err := GetClanByID(db, id)
	if err != nil {
		return err
	}
	if clan == nil {
		return &ModelNotFoundError{"Clan", id}
	}
	return clan.UpdateClanIntoMongoDB()
}

// Serialize returns a JSON with clan details
func (c *Clan) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"gameID":           c.GameID,
		"publicID":         c.PublicID,
		"name":             c.Name,
		"membershipCount":  c.MembershipCount,
		"metadata":         c.Metadata,
		"allowApplication": c.AllowApplication,
		"autoJoin":         c.AutoJoin,
	}
}

// UpdateClanMembershipCount updates the clan membership count
func UpdateClanMembershipCount(db DB, id int) error {
	query := `
	UPDATE clans SET membership_count=membership.count+1
	FROM (
		SELECT COUNT(*) as count
		FROM memberships m
		WHERE
			m.clan_id = $1 AND m.deleted_at = 0 AND m.approved = true AND
			m.denied = false AND m.banned = false
	) as membership
	WHERE clans.id=$1
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return &ModelNotFoundError{"Clan", id}
	}

	err = updateClanIntoES(db, id)

	if err != nil {
		return err
	}

	err = updateClanIntoMongo(db, id)

	if err != nil {
		return err
	}

	return nil
}

// GetClanByID returns a clan by id
func GetClanByID(db DB, id int) (*Clan, error) {
	obj, err := db.Get(Clan{}, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, &ModelNotFoundError{"Clan", id}
	}
	return obj.(*Clan), nil
}

// GetClanByPublicID returns a clan by its public id
func GetClanByPublicID(db DB, gameID, publicID string) (*Clan, error) {
	var clans []*Clan
	_, err := db.Select(&clans, "SELECT * FROM clans WHERE game_id=$1 AND public_id=$2", gameID, publicID)
	if err != nil {
		return nil, err
	}
	if clans == nil || len(clans) < 1 {
		return nil, &ModelNotFoundError{"Clan", publicID}
	}
	return clans[0], nil
}

func sliceDiff(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

// GetClansByPublicIDs returns clans by their public ids
func GetClansByPublicIDs(db DB, gameID string, publicIDs []string) ([]Clan, error) {
	var clans []Clan

	queryPart := "SELECT * from clans WHERE game_id=$1 AND public_id=%s"
	queryParts := []string{}
	for i := 0; i < len(publicIDs); i++ {
		paramIndex := fmt.Sprintf("$%d", (i + 2))
		queryParts = append(queryParts, fmt.Sprintf(queryPart, paramIndex))
	}
	query := strings.Join(queryParts, " UNION ALL ")

	params := []interface{}{gameID}
	for _, publicID := range publicIDs {
		params = append(params, publicID)
	}

	_, err := db.Select(&clans, query, params...)
	if err != nil {
		return nil, err
	}

	clanIDs := make([]string, len(clans))
	for i, clan := range clans {
		clanIDs[i] = clan.PublicID
	}
	diff := sliceDiff(publicIDs, clanIDs)
	if len(diff) != 0 {
		if len(clans) == 0 {
			return clans, &CouldNotFindAllClansError{gameID, publicIDs}
		}
		return clans, &CouldNotFindAllClansError{gameID, diff}
	}
	return clans, nil
}

// GetClanByPublicIDAndOwnerPublicID returns a clan by its public id and the owner public id
func GetClanByPublicIDAndOwnerPublicID(db DB, gameID, publicID, ownerPublicID string) (*Clan, error) {
	var clans []*Clan
	_, err := db.Select(&clans, "SELECT clans.* FROM clans, players WHERE clans.game_id=$1 AND clans.public_id=$2 AND clans.owner_id=players.id AND players.public_id=$3", gameID, publicID, ownerPublicID)
	if err != nil {
		return nil, err
	}
	if clans == nil || len(clans) < 1 {
		return nil, &ModelNotFoundError{"Clan", publicID}
	}
	return clans[0], nil
}

// CreateClan creates a new clan
func CreateClan(db DB, gameID, publicID, name, ownerPublicID string, metadata map[string]interface{}, allowApplication, autoJoin bool, maxClansPerPlayer int) (*Clan, error) {
	player, err := GetPlayerByPublicID(db, gameID, ownerPublicID)
	if err != nil {
		return nil, err
	}

	if player.MembershipCount+player.OwnershipCount >= maxClansPerPlayer {
		return nil, &PlayerReachedMaxClansError{ownerPublicID}
	}

	clan := &Clan{
		GameID:           gameID,
		PublicID:         publicID,
		Name:             name,
		OwnerID:          player.ID,
		Metadata:         metadata,
		AllowApplication: allowApplication,
		AutoJoin:         autoJoin,
		MembershipCount:  1,
	}

	err = db.Insert(clan)
	if err != nil {
		return nil, err
	}

	err = UpdatePlayerOwnershipCount(db, player.ID)
	if err != nil {
		return nil, err
	}

	return clan, nil
}

// LeaveClan allows the clan owner to leave the clan and transfer the clan ownership to the next player in line
func LeaveClan(db DB, gameID, publicID string) (*Clan, *Player, *Player, error) {
	clan, err := GetClanByPublicID(db, gameID, publicID)
	if err != nil {
		return nil, nil, nil, err
	}

	oldOwnerID := clan.OwnerID
	oldOwner, err := GetPlayerByID(db, oldOwnerID)
	if err != nil {
		return nil, nil, nil, err
	}

	newOwnerMembership, err := GetOldestMemberWithHighestLevel(db, gameID, publicID)
	if err != nil {
		noMembersError := &ClanHasNoMembersError{publicID}
		if err.Error() == noMembersError.Error() {
			// Clan has no approved members, delete all members and clan
			_, err = db.Exec("DELETE FROM memberships where clan_id=$1", clan.ID)
			if err != nil {
				return nil, nil, nil, err
			}
			_, err = db.Delete(clan)
			if err != nil {
				return nil, nil, nil, err
			}
			err = UpdatePlayerOwnershipCount(db, oldOwnerID)
			if err != nil {
				return nil, nil, nil, err
			}
			return clan, oldOwner, nil, nil
		}
		return nil, nil, nil, err
	}

	newOwner, err := GetPlayerByID(db, newOwnerMembership.PlayerID)
	if err != nil {
		return nil, nil, nil, err
	}

	clan.OwnerID = newOwner.ID

	_, err = db.Update(clan)
	if err != nil {
		return nil, nil, nil, err
	}

	_, err = deleteMembershipHelper(db, newOwnerMembership, newOwnerMembership.PlayerID)
	if err != nil {
		return nil, nil, nil, err
	}

	err = UpdatePlayerOwnershipCount(db, oldOwnerID)
	if err != nil {
		return nil, nil, nil, err
	}
	err = UpdatePlayerOwnershipCount(db, newOwner.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	newOwner.MembershipCount--
	newOwner.OwnershipCount++

	// We update it here on purpose, to avoid making another useless query
	// The actual update to reduce membership count is made in the deleteMembershipHelper
	clan.MembershipCount--

	return clan, oldOwner, newOwner, nil
}

// TransferClanOwnership allows the clan owner to transfer the clan ownership to the a clan member
func TransferClanOwnership(db DB, gameID, clanPublicID, playerPublicID string, levels map[string]interface{}, maxLevel int) (*Clan, *Player, *Player, error) {
	clan, err := GetClanByPublicID(db, gameID, clanPublicID)
	if err != nil {
		return nil, nil, nil, err
	}

	newOwnerMembership, err := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, nil, nil, err
	}

	oldOwnerID := clan.OwnerID
	clan.OwnerID = newOwnerMembership.PlayerID
	_, err = db.Update(clan)
	if err != nil {
		return nil, nil, nil, err
	}

	level := GetLevelByLevelInt(maxLevel, levels)
	if level == "" {
		return nil, nil, nil, &InvalidLevelForGameError{gameID, level}
	}
	oldOwnerMembership, err := GetDeletedMembershipByClanAndPlayerID(db, gameID, clan.ID, oldOwnerID)
	if err != nil {
		err = db.Insert(&Membership{
			GameID:      gameID,
			ClanID:      clan.ID,
			PlayerID:    oldOwnerID,
			RequestorID: oldOwnerID,
			Level:       level,
			Approved:    true,
			Denied:      false,
			Banned:      false,
			CreatedAt:   clan.CreatedAt,
			UpdatedAt:   util.NowMilli(),
		})
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		oldOwnerMembership.Approved = true
		oldOwnerMembership.Denied = false
		oldOwnerMembership.Banned = false
		oldOwnerMembership.DeletedBy = 0
		oldOwnerMembership.DeletedAt = 0
		oldOwnerMembership.Level = level
		oldOwnerMembership.RequestorID = oldOwnerID

		_, err = db.Update(oldOwnerMembership)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	//Update new owner memberships
	_, err = deleteMembershipHelper(db, newOwnerMembership, newOwnerMembership.PlayerID)
	if err != nil {
		return nil, nil, nil, err
	}

	err = UpdatePlayerOwnershipCount(db, newOwnerMembership.PlayerID)
	if err != nil {
		return nil, nil, nil, err
	}

	newOwner, err := GetPlayerByID(db, newOwnerMembership.PlayerID)
	if err != nil {
		return nil, nil, nil, err
	}

	//Update old owner membership
	err = UpdatePlayerOwnershipCount(db, oldOwnerID)
	if err != nil {
		return nil, nil, nil, err
	}

	err = UpdatePlayerMembershipCount(db, oldOwnerID)
	if err != nil {
		return nil, nil, nil, err
	}

	oldOwner, err := GetPlayerByID(db, oldOwnerID)
	if err != nil {
		return nil, nil, nil, err
	}

	return clan, oldOwner, newOwner, nil
}

// UpdateClan updates an existing clan
func UpdateClan(db DB, gameID, publicID, name, ownerPublicID string, metadata map[string]interface{}, allowApplication, autoJoin bool) (*Clan, error) {
	clan, err := GetClanByPublicIDAndOwnerPublicID(db, gameID, publicID, ownerPublicID)

	if err != nil {
		return nil, err
	}

	clan.Name = name
	clan.Metadata = metadata
	clan.AllowApplication = allowApplication
	clan.AutoJoin = autoJoin

	_, err = db.Update(clan)

	if err != nil {
		return nil, err
	}

	return clan, nil
}

// GetAllClans returns a list of all clans in a given game
func GetAllClans(db DB, gameID string) ([]Clan, error) {
	if gameID == "" {
		return nil, &EmptyGameIDError{"Clan"}
	}

	var clans []Clan
	_, err := db.Select(&clans, "select * from clans where game_id=$1 order by name", gameID)
	if err != nil {
		return nil, err
	}

	return clans, nil
}

// GetClanMembers gets only the ids of then clan members
func GetClanMembers(db DB, gameID, publicID string) (map[string]interface{}, error) {
	clan, err := GetClanByPublicID(db, gameID, publicID)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT public_id
	FROM players
	WHERE id IN (
		SELECT player_id
		FROM memberships im
		WHERE im.clan_id=$1
		AND im.deleted_at=0
		AND (im.approved=true)
		UNION
			SELECT owner_id
			FROM clans c
			WHERE c.id=$1
	)
	`
	var res []string
	_, err = db.Select(&res, query, clan.ID)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"members": res,
	}, nil

}

// GetClanDetails returns all details for a given clan by its game id and public id
func GetClanDetails(db DB, gameID string, clan *Clan, maxClansPerPlayer int) (map[string]interface{}, error) {
	query := `
	SELECT
		c.game_id GameID,
		c.public_id ClanPublicID, c.name ClanName, c.metadata ClanMetadata,
		c.allow_application ClanAllowApplication, c.auto_join ClanAutoJoin,
		c.membership_count ClanMembershipCount,
		m.membership_level MembershipLevel, m.approved MembershipApproved, m.denied MembershipDenied,
		m.banned MembershipBanned, m.message MembershipMessage,
		m.created_at MembershipCreatedAt, m.updated_at MembershipUpdatedAt,
		m.approved_at MembershipApprovedAt, m.denied_at MembershipDeniedAt,
		o.public_id OwnerPublicID, o.name OwnerName, o.metadata OwnerMetadata,
		p.public_id PlayerPublicID, p.name PlayerName, p.metadata DBPlayerMetadata,
		r.public_id RequestorPublicID, r.name RequestorName,
		a.public_id ApproverPublicID, a.name ApproverName,
		y.name DenierName, y.public_id DenierPublicID,
		Coalesce(p.membership_count, 0) MembershipCount,
		Coalesce(p.ownership_count, 0) OwnershipCount
	FROM clans c
		INNER JOIN players o ON c.owner_id=o.id
		LEFT OUTER JOIN (
			SELECT *
			FROM memberships im
			WHERE im.clan_id=$2 AND im.deleted_at=0 AND (im.approved=true OR im.denied=true OR im.banned=true)
		) m ON m.clan_id=c.id
		LEFT OUTER JOIN players r ON m.requestor_id=r.id
		LEFT OUTER JOIN players a ON m.approver_id=a.id
		LEFT OUTER JOIN players p ON m.player_id=p.id
		LEFT OUTER JOIN players y ON m.denier_id=y.id
	WHERE
		c.game_id=$1 AND c.id=$2
	`
	var details []clanDetailsDAO
	_, err := db.Select(&details, query, gameID, clan.ID)
	if err != nil {
		return nil, err
	}

	if len(details) == 0 {
		return nil, &ModelNotFoundError{"Clan", clan.PublicID}
	}

	result := make(map[string]interface{})
	result["name"] = details[0].ClanName
	result["metadata"] = details[0].ClanMetadata
	result["allowApplication"] = details[0].ClanAllowApplication
	result["autoJoin"] = details[0].ClanAutoJoin
	result["membershipCount"] = details[0].ClanMembershipCount

	result["owner"] = map[string]interface{}{
		"publicID": details[0].OwnerPublicID,
		"name":     details[0].OwnerName,
		"metadata": details[0].OwnerMetadata,
	}

	// First row player public id is not null, meaning we found players!
	if details[0].PlayerPublicID.Valid {

		result["roster"] = make([]map[string]interface{}, 0)
		result["memberships"] = map[string]interface{}{
			"pendingInvites":      []map[string]interface{}{},
			"pendingApplications": []map[string]interface{}{},
			"banned":              []map[string]interface{}{},
			"denied":              []map[string]interface{}{},
		}
		memberships := result["memberships"].(map[string]interface{})

		for _, member := range details {
			approved := nullOrBool(member.MembershipApproved)
			denied := nullOrBool(member.MembershipDenied)
			banned := nullOrBool(member.MembershipBanned)
			pending := !approved && !denied && !banned

			switch {
			case pending:
				memberData := member.Serialize(true)
				if member.MembershipCount+member.OwnershipCount < maxClansPerPlayer {
					if member.PlayerPublicID == member.RequestorPublicID {
						memberData["message"] = nullOrString(member.MembershipMessage)
						memberships["pendingApplications"] = append(memberships["pendingApplications"].([]map[string]interface{}), memberData)
					} else {
						memberships["pendingInvites"] = append(memberships["pendingInvites"].([]map[string]interface{}), memberData)
					}
				}
			case banned:
				memberData := member.Serialize(false)
				memberships["banned"] = append(memberships["banned"].([]map[string]interface{}), memberData)
			case denied:
				memberData := member.Serialize(false)
				memberships["denied"] = append(memberships["denied"].([]map[string]interface{}), memberData)
			case approved:
				memberData := member.Serialize(true)
				result["roster"] = append(result["roster"].([]map[string]interface{}), memberData)
			}
		}
	} else {
		//Otherwise return empty array of object
		result["roster"] = []map[string]interface{}{}
		result["memberships"] = map[string]interface{}{
			"pendingApplications": []map[string]interface{}{},
			"pendingInvites":      []map[string]interface{}{},
			"banned":              []map[string]interface{}{},
			"denied":              []map[string]interface{}{},
		}
	}

	return result, nil
}

// GetClanSummary returns a summary of the clan details for a given clan by its game id and public id
func GetClanSummary(db DB, gameID, publicID string) (map[string]interface{}, error) {
	clan, err := GetClanByPublicID(db, gameID, publicID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["membershipCount"] = clan.MembershipCount
	result["publicID"] = clan.PublicID
	result["metadata"] = clan.Metadata
	result["name"] = clan.Name
	result["allowApplication"] = clan.AllowApplication
	result["autoJoin"] = clan.AutoJoin
	return result, nil
}

// GetClansSummaries returns a summary of the clans details for a given list of clans by their game
// id and public ids
func GetClansSummaries(db DB, gameID string, publicIDs []string) ([]map[string]interface{}, error) {
	clans, err := GetClansByPublicIDs(db, gameID, publicIDs)
	resultClans := make([]map[string]interface{}, len(clans))

	for i := range clans {
		result := map[string]interface{}{
			"membershipCount":  clans[i].MembershipCount,
			"publicID":         clans[i].PublicID,
			"metadata":         clans[i].Metadata,
			"name":             clans[i].Name,
			"allowApplication": clans[i].AllowApplication,
			"autoJoin":         clans[i].AutoJoin,
		}
		resultClans[i] = result
	}

	if err != nil {
		return resultClans, err
	}

	return resultClans, nil
}

// SearchClan returns a list of clans for a given term (by name or publicID)
func SearchClan(db DB, gameID, term string) ([]Clan, error) {
	if term == "" {
		return nil, &EmptySearchTermError{}
	}

	query := `SELECT * FROM clans WHERE game_id=$1 AND (lower(name) like $2) LIMIT 50`

	termStmt := fmt.Sprintf("%%%s%%", strings.ToLower(term))

	var clans []Clan
	_, err := db.Select(&clans, query, gameID, termStmt)
	if err != nil {
		return nil, err
	}

	return clans, nil
}

// GetClanAndOwnerByPublicID returns the clan as well as the owner of a clan by clan's public id
func GetClanAndOwnerByPublicID(db DB, gameID, publicID string) (*Clan, *Player, error) {
	clan, err := GetClanByPublicID(db, gameID, publicID)
	if err != nil {
		return nil, nil, err
	}
	newOwner, err := GetPlayerByID(db, clan.OwnerID)
	if err != nil {
		return nil, nil, err
	}

	return clan, newOwner, nil
}
