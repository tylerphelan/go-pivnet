package eula_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/eula"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EULA commands suite")
}

var _ = BeforeSuite(func() {
	eula.Format = printer.PrintAsJSON
})
