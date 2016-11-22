// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"

	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
)

// PruneStats show stats about what has been pruned
type PruneStats struct {
	PendingApplicationsPruned int
	PendingInvitesPruned      int
	DeniedMembershipsPruned   int
	DeletedMembershipsPruned  int
}

//GetStats returns a formatted message
func (ps *PruneStats) GetStats() string {
	return fmt.Sprintf(
		"-Pending Applications: %d\n-Pending Invites: %d\nDenied Memberships: %d\nDeleted Memberships: %d\n",
		ps.PendingApplicationsPruned,
		ps.PendingInvitesPruned,
		ps.DeniedMembershipsPruned,
		ps.DeletedMembershipsPruned,
	)
}

// PruneOptions has all the prunable memberships TTL
type PruneOptions struct {
	GameID                        string
	PendingApplicationsExpiration int
	PendingInvitesExpiration      int
	DeniedMembershipsExpiration   int
	DeletedMembershipsExpiration  int
}

func runAndReturnRowsAffected(query string, db DB, args ...interface{}) (int, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	return int(rows), err
}

func prunePendingApplications(options *PruneOptions, db DB, logger zap.Logger) (int, error) {
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

func prunePendingInvites(options *PruneOptions, db DB, logger zap.Logger) (int, error) {
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

func pruneDeniedMemberships(options *PruneOptions, db DB, logger zap.Logger) (int, error) {
	query := `DELETE FROM memberships m WHERE
		m.game_id=$1 AND
		m.denied=TRUE AND
		m.updated_at < $2`

	updatedAt := util.NowMilli() - int64(options.DeniedMembershipsExpiration*1000)
	return runAndReturnRowsAffected(query, db, options.GameID, updatedAt)
}

func pruneDeletedMemberships(options *PruneOptions, db DB, logger zap.Logger) (int, error) {
	query := `DELETE FROM memberships m WHERE
		m.game_id=$1 AND
		m.deleted_at > 0 AND
		m.updated_at < $2`

	updatedAt := util.NowMilli() - int64(options.DeletedMembershipsExpiration*1000)
	return runAndReturnRowsAffected(query, db, options.GameID, updatedAt)
}

// PruneStaleData off of Khan's database
func PruneStaleData(options *PruneOptions, db DB, logger zap.Logger) (*PruneStats, error) {
	log.I(logger, "Pruning stale data...", func(cm log.CM) {
		cm.Write(zap.String("GameID", options.GameID))
		cm.Write(zap.Int("PendingApplicationsExpiration", options.PendingApplicationsExpiration))
		cm.Write(zap.Int("PendingInvitesExpiration", options.PendingInvitesExpiration))
		cm.Write(zap.Int("DeniedMembershipsExpiration", options.DeniedMembershipsExpiration))
		cm.Write(zap.Int("DeletedMembershipsExpiration", options.DeletedMembershipsExpiration))
	})

	pendingApplicationsPruned, err := prunePendingApplications(options, db, logger)
	if err != nil {
		log.E(logger, "Failed to prune stale pending applications.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	pendingInvitesPruned, err := prunePendingInvites(options, db, logger)
	if err != nil {
		log.E(logger, "Failed to prune stale pending invites.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	deniedMembershipsPruned, err := pruneDeniedMemberships(options, db, logger)
	if err != nil {
		log.E(logger, "Failed to prune stale denied memberships.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	deletedMembershipsPruned, err := pruneDeletedMemberships(options, db, logger)
	if err != nil {
		log.E(logger, "Failed to prune stale deleted memberships.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, err
	}

	stats := &PruneStats{
		PendingApplicationsPruned: pendingApplicationsPruned,
		PendingInvitesPruned:      pendingInvitesPruned,
		DeniedMembershipsPruned:   deniedMembershipsPruned,
		DeletedMembershipsPruned:  deletedMembershipsPruned,
	}

	log.I(logger, "Pruned stale data succesfully.", func(cm log.CM) {
		cm.Write(zap.String("GameID", options.GameID))
		cm.Write(zap.Int("PendingApplicationsPruned", stats.PendingApplicationsPruned))
		cm.Write(zap.Int("PendingInvitesPruned", stats.PendingInvitesPruned))
		cm.Write(zap.Int("DeniedMembershipsPruned", stats.DeniedMembershipsPruned))
		cm.Write(zap.Int("DeletedMembershipsPruned", stats.DeletedMembershipsPruned))
	})
	return stats, nil
}
