package loadtest

// Operation represents a Khan operation to be tested in the load tests
type Operation struct {
	probability float64
	canExecute  func() (bool, error)
	execute     func() error
}
