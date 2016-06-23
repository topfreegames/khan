// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"

	"github.com/topfreegames/khan/util"
)

type clanDetailsDAO struct {
	// Clan general information
	GameID               string
	ClanPublicID         string
	ClanName             string
	ClanMetadata         string
	ClanAllowApplication bool
	ClanAutoJoin         bool

	//Membership Information
	MembershipLevel     sql.NullInt64
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipBanned    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64

	// Clan Owner Information
	OwnerPublicID string
	OwnerName     string
	OwnerMetadata string

	// Member Information
	PlayerPublicID sql.NullString
	PlayerName     sql.NullString
	PlayerMetadata sql.NullString

	// Requestor Information
	RequestorPublicID sql.NullString
	RequestorName     sql.NullString
}

func (member *clanDetailsDAO) Serialize() util.JSON {
	return util.JSON{
		// No need to include clan information as that will be available in the payload already
		"membershipLevel":     nullOrInt(member.MembershipLevel),
		"membershipApproved":  nullOrBool(member.MembershipApproved),
		"membershipDenied":    nullOrBool(member.MembershipDenied),
		"membershipBanned":    nullOrBool(member.MembershipBanned),
		"membershipCreatedAt": nullOrInt(member.MembershipCreatedAt),
		"membershipUpdatedAt": nullOrInt(member.MembershipUpdatedAt),
		"playerPublicID":      nullOrString(member.PlayerPublicID),
		"playerName":          nullOrString(member.PlayerName),
		"playerMetadata":      nullOrString(member.PlayerMetadata),
		"requestorPublicID":   nullOrString(member.RequestorPublicID),
		"requestorName":       nullOrString(member.RequestorName),
	}
}

type playerDetailsDAO struct {
	// Player Details
	PlayerName      string
	PlayerMetadata  string
	PlayerPublicID  string
	PlayerCreatedAt int64
	PlayerUpdatedAt int64

	// Membership Details
	MembershipLevel     sql.NullInt64
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipBanned    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64
	MembershipDeletedAt sql.NullInt64

	// Clan Details
	ClanPublicID sql.NullString
	ClanName     sql.NullString
	ClanMetadata sql.NullString

	// Membership Requestor Details
	RequestorName     sql.NullString
	RequestorPublicID sql.NullString
	RequestorMetadata sql.NullString

	// Deleted by Details
	DeletedByName     sql.NullString
	DeletedByPublicID sql.NullString
}

func (p *playerDetailsDAO) Serialize() util.JSON {
	result := util.JSON{
		"level":     nullOrInt(p.MembershipLevel),
		"approved":  nullOrBool(p.MembershipApproved),
		"denied":    nullOrBool(p.MembershipDenied),
		"banned":    nullOrBool(p.MembershipBanned),
		"createdAt": nullOrInt(p.MembershipCreatedAt),
		"updatedAt": nullOrInt(p.MembershipUpdatedAt),
		"deletedAt": nullOrInt(p.MembershipDeletedAt),
		"clan": util.JSON{
			"publicID": nullOrString(p.ClanPublicID),
			"name":     nullOrString(p.ClanName),
			"metadata": nullOrString(p.ClanMetadata),
		},
		"requestor": util.JSON{
			"publicID": nullOrString(p.RequestorPublicID),
			"name":     nullOrString(p.RequestorName),
			"metadata": nullOrString(p.RequestorMetadata),
		},
	}
	if p.DeletedByPublicID.Valid {
		result["deletedBy"] = util.JSON{
			"publicID": nullOrString(p.DeletedByPublicID),
			"name":     nullOrString(p.DeletedByName),
		}
	}
	return result
}
