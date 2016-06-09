package models

import "fmt"

//ModelNotFoundError identifies that a given model was not found in the Database with the given ID
type ModelNotFoundError struct {
	Type string
	ID   int
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("%s was not found with id: %d", e.Type, e.ID)
}
