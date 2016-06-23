// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"
	"encoding/json"

	"github.com/topfreegames/khan/util"
)

type clanDetailsDAO struct {
	// Clan general information
	GameID               string
	ClanPublicID         string
	ClanName             string
	ClanMetadata         util.JSON
	ClanAllowApplication bool
	ClanAutoJoin         bool

	//Membership Information
	MembershipLevel     sql.NullString
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipBanned    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64

	// Clan Owner Information
	OwnerPublicID string
	OwnerName     string
	OwnerMetadata util.JSON

	// Member Information
	PlayerPublicID   sql.NullString
	PlayerName       sql.NullString
	DBPlayerMetadata sql.NullString
	PlayerMetadata   util.JSON

	// Requestor Information
	RequestorPublicID sql.NullString
	RequestorName     sql.NullString
}

func (member *clanDetailsDAO) Serialize() util.JSON {
	result := util.JSON{
		// No need to include clan information as that will be available in the payload already
		"membershipLevel":     nullOrString(member.MembershipLevel),
		"membershipApproved":  nullOrBool(member.MembershipApproved),
		"membershipDenied":    nullOrBool(member.MembershipDenied),
		"membershipBanned":    nullOrBool(member.MembershipBanned),
		"membershipCreatedAt": nullOrInt(member.MembershipCreatedAt),
		"membershipUpdatedAt": nullOrInt(member.MembershipUpdatedAt),
		"playerPublicID":      nullOrString(member.PlayerPublicID),
		"playerName":          nullOrString(member.PlayerName),
		"requestorPublicID":   nullOrString(member.RequestorPublicID),
		"requestorName":       nullOrString(member.RequestorName),
	}
	if member.DBPlayerMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(member.DBPlayerMetadata)), &member.PlayerMetadata)
	} else {
		member.PlayerMetadata = util.JSON{}
	}
	result["playerMetadata"] = member.PlayerMetadata
	return result
}

type playerDetailsDAO struct {
	// Player Details
	PlayerName      string
	PlayerMetadata  util.JSON
	PlayerPublicID  string
	PlayerCreatedAt int64
	PlayerUpdatedAt int64

	// Membership Details
	MembershipLevel     sql.NullString
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipBanned    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64
	MembershipDeletedAt sql.NullInt64

	// Clan Details
	ClanPublicID   sql.NullString
	ClanName       sql.NullString
	DBClanMetadata sql.NullString
	ClanMetadata   util.JSON

	// Membership Requestor Details
	RequestorName       sql.NullString
	RequestorPublicID   sql.NullString
	DBRequestorMetadata sql.NullString
	RequestorMetadata   util.JSON

	// Deleted by Details
	DeletedByName     sql.NullString
	DeletedByPublicID sql.NullString
}

func (p *playerDetailsDAO) Serialize() util.JSON {
	result := util.JSON{
		"level":     nullOrString(p.MembershipLevel),
		"approved":  nullOrBool(p.MembershipApproved),
		"denied":    nullOrBool(p.MembershipDenied),
		"banned":    nullOrBool(p.MembershipBanned),
		"createdAt": nullOrInt(p.MembershipCreatedAt),
		"updatedAt": nullOrInt(p.MembershipUpdatedAt),
		"deletedAt": nullOrInt(p.MembershipDeletedAt),
		"clan": util.JSON{
			"publicID": nullOrString(p.ClanPublicID),
			"name":     nullOrString(p.ClanName),
		},
		"requestor": util.JSON{
			"publicID": nullOrString(p.RequestorPublicID),
			"name":     nullOrString(p.RequestorName),
		},
	}

	if p.DBClanMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(p.DBClanMetadata)), &p.ClanMetadata)
	} else {
		p.ClanMetadata = util.JSON{}
	}
	result["clan"].(util.JSON)["metadata"] = p.ClanMetadata

	if p.DBRequestorMetadata.Valid {
		json.Unmarshal([]byte(nullOrString(p.DBRequestorMetadata)), &p.RequestorMetadata)
	} else {
		p.RequestorMetadata = util.JSON{}
	}
	result["requestor"].(util.JSON)["metadata"] = p.RequestorMetadata

	if p.DeletedByPublicID.Valid {
		result["deletedBy"] = util.JSON{
			"publicID": nullOrString(p.DeletedByPublicID),
			"name":     nullOrString(p.DeletedByName),
		}
	}
	return result
}
