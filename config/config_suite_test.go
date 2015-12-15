package config_test

import (
	"testing"

	. "github.com/eljuanchosf/gocafier/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config loader test suite")
}

var _ = Describe("Config Loader", func() {
	Describe("ParseConfig", func() {
		Context("called with a empty string", func() {
			It("should return a empty hash", func() {
				expected := map[string]string{}
				Expect(ParseExtraFields("")).To(Equal(expected))
			})
		})

		Context("called with extra events", func() {
			It("should return a hash with the events we want", func() {
				expected := map[string]string{"env": "dev", "kehe": "wakawaka"}
				extraEvents := "env:dev,kehe:wakawaka"
				Expect(ParseExtraFields(extraEvents)).To(Equal(expected))
			})
		})

		Context("called with extra events with weird whitespace", func() {
			It("should return a hash with the events we want", func() {
				expected := map[string]string{"env": "dev", "kehe": "wakawaka"}
				extraEvents := "    env:      \ndev,      kehe:wakawaka   "
				Expect(ParseExtraFields(extraEvents)).To(Equal(expected))
			})
		})

		Context("called with extra events with to many values to a kv pair", func() {
			It("should return a error", func() {
				extraEvents := "to:many:values"
				_, err := ParseExtraFields(extraEvents)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
