package loadtest

import "fmt"

// GenericError represents a generic error
type GenericError struct {
	Type        string
	Description string
}

func (e *GenericError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Description)
}
