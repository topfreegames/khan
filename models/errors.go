// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import "fmt"

//ModelNotFoundError identifies that a given model was not found in the Database with the given ID
type ModelNotFoundError struct {
	Type string
	ID   interface{}
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("%s was not found with id: %v", e.Type, e.ID)
}

//EmptyGameIDError identifies that a request was made for a model without the proper game id
type EmptyGameIDError struct {
	Type string
}

func (e *EmptyGameIDError) Error() string {
	return fmt.Sprintf("Game ID is required to retrieve %s!", e.Type)
}
