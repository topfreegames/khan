// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"testing"
)

func TestApi(t *testing.T) {
	wl := logrus.New()
	wl.Level = logrus.FatalLevel
	workers.SetLogger(wl)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Khan - API Suite")
}
