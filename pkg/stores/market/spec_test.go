package market

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKlines(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Market Store Suite")
}
