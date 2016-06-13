// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

//GameIDNotFoundError identifies a request with a null or empty game id
type GameIDNotFoundError struct {
}

func (e *GameIDNotFoundError) Error() string {
	return "The game ID is required for this request."
}
