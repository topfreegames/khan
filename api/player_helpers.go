// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

type createPlayerPayload struct {
	PublicID string
	Name     string
	Metadata map[string]interface{}
}

type updatePlayerPayload struct {
	Name     string
	Metadata map[string]interface{}
}
