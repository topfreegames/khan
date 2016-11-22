// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
)

// PruneStats show stats about what has been pruned
type PruneStats struct {
	PendingApplicationsPruned int64
	PendingInvitesPruned      int64
	DeniedMembershipsPruned   int64
	DeletedMembershipsPruned  int64
}

// PruneOptions has all the prunable memberships TTL
type PruneOptions struct {
	GameID                        string
	PendingApplicationsExpiration int
	PendingInvitesExpiration      int
	DeniedMembershipsExpiration   int
	DeletedMembershipsExpiration  int
}

func runAndReturnRowsAffected(query string, db DB, args ...interface{}) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	return rows, err
}

func prunePendingApplications(options *PruneOptions, db DB, logger zap.Logger) (int64, error) {
	query := `DELETE FROM memberships m WHERE
		m.game_id=$1 AND
		m.deleted_at=0 AND
		m.approved=FALSE AND
		m.denied=FALSE AND
		m.requestor_id=m.player_id AND
		m.updated_at < $2`

	updatedAt := util.NowMilli() - int64(options.PendingApplicationsExpiration*1000)
	return runAndReturnRowsAffected(query, db, options.GameID, updatedAt)
}

func prunePendingInvites(options *PruneOptions, db DB, logger zap.Logger) (int64, error) {
	query := `DELETE FROM memberships m WHERE
		m.game_id=$1 AND
		m.deleted_at=0 AND
		m.approved=FALSE AND
		m.denied=FALSE AND
		m.requestor_id != m.player_id AND
		m.updated_at < $2`

	updatedAt := util.NowMilli() - int64(options.PendingInvitesExpiration*1000)
	return runAndReturnRowsAffected(query, db, options.GameID, updatedAt)
}

func pruneDeniedMemberships(options *PruneOptions, db DB, logger zap.Logger) (int64, error) {
	query := `DELETE FROM memberships m WHERE
		m.game_id=$1 AND
		m.denied=TRUE AND
		m.updated_at < $2`

	updatedAt := util.NowMilli() - int64(options.DeniedMembershipsExpiration*1000)
	return runAndReturnRowsAffected(query, db, options.GameID, updatedAt)
}

func pruneDeletedMemberships(options *PruneOptions, db DB, logger zap.Logger) (int64, error) {
	query := `DELETE FROM memberships m WHERE
		m.game_id=$1 AND
		m.deleted_at > 0 AND
		m.updated_at < $2`

	updatedAt := util.NowMilli() - int64(options.DeletedMembershipsExpiration*1000)
	return runAndReturnRowsAffected(query, db, options.GameID, updatedAt)
}

// PruneStaleData off of Khan's database
func PruneStaleData(options *PruneOptions, db DB, logger zap.Logger) (*PruneStats, error) {
	pendingApplicationsPruned, err := prunePendingApplications(options, db, logger)
	if err != nil {
		return nil, err
	}

	pendingInvitesPruned, err := prunePendingInvites(options, db, logger)
	if err != nil {
		return nil, err
	}

	deniedMembershipsPruned, err := pruneDeniedMemberships(options, db, logger)
	if err != nil {
		return nil, err
	}

	deletedMembershipsPruned, err := pruneDeletedMemberships(options, db, logger)
	if err != nil {
		return nil, err
	}

	stats := &PruneStats{
		PendingApplicationsPruned: pendingApplicationsPruned,
		PendingInvitesPruned:      pendingInvitesPruned,
		DeniedMembershipsPruned:   deniedMembershipsPruned,
		DeletedMembershipsPruned:  deletedMembershipsPruned,
	}
	return stats, nil
}
