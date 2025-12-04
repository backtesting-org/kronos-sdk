package trade_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTradeStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trade Store Suite")
}
