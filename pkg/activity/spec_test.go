package activity_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestActivity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Activity Suite")
}
