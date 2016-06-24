// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package util

import "sort"

func getIntLevel(level interface{}) int {
	switch level.(type) {
	case float64:
		return int(level.(float64))
	default:
		return level.(int)
	}
}

// SortLevels sorts levels
func SortLevels(levels JSON) LevelsList {
	ll := make(LevelsList, len(levels))
	i := 0
	for k, v := range levels {
		ll[i] = Level{k, getIntLevel(v)}
		i++
	}
	sort.Sort(ll)
	return ll
}

// Level maps levels
type Level struct {
	Key   string
	Value int
}

// LevelsList allows sorting levels by the int value
type LevelsList []Level

func (l LevelsList) Len() int           { return len(l) }
func (l LevelsList) Less(i, j int) bool { return l[i].Value < l[j].Value }
func (l LevelsList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
