package yevna_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestYevna(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Yevna Suite")
}
