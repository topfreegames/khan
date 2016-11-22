// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import "github.com/uber-go/zap"

//PruneStats show stats about what has been pruned
type PruneStats struct {
	PendingApplicationsPruned int
	PendingInvitesPruned      int
	DeniedMembershipsPruned   int
	DeletedMembershipsPruned  int
}

type PruneOptions struct {
	PendingApplicationsExpiration int
}

func prunePendingApplications(options *PruneOptions, db DB, logger zap.Logger) (int, error) {
	return 0, nil
}

//PruneStaleData off of Khan's database
func PruneStaleData(options *PruneOptions, db DB, logger zap.Logger) (*PruneStats, error) {
	pendingApplicationsPruned, err := prunePendingApplications(options, db, logger)
	if err != nil {
		return nil, err
	}
	stats := &PruneStats{
		PendingApplicationsPruned: pendingApplicationsPruned,
	}
	return stats, nil
}
