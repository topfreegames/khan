// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package testing

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
)

//BeforeOnce runs the before each block only once
func BeforeOnce(beforeBlock func()) {
	hasRun := false

	ginkgo.BeforeEach(func() {
		if !hasRun {
			beforeBlock()
			hasRun = true
		}
	})
}

//WaitForFunc waits for a given function to finish without error or a timeout
func WaitForFunc(timeout int, f func() error) error {
	var err error

	start := time.Now()

	for err = f(); err != nil || int(time.Now().Sub(start).Seconds()) > timeout; err = f() {
		time.Sleep(time.Millisecond)
	}
	if int(time.Now().Sub(start).Seconds()) > timeout {
		return fmt.Errorf("Timeout")
	}
	return nil
}
