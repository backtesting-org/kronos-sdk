package position_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPositionStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Position Store Suite")
}
