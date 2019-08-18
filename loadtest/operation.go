package loadtest

// operation represents a Khan operation to be tested in the load tests
type operation struct {
	key             string
	wontUpdateCache bool
	probability     float64
	canExecute      func() (bool, error)
	execute         func() error
}
