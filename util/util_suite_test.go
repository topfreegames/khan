package util_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKhan(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Khan Suite")
}
