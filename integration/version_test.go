/* #nosec */

package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("version subcommand", func() {
	Describe("using no flags", func() {
		It("should print the current version", func() {
			pluginCommand := exec.Command(
				_pluginPath, "version",
			)

			session, err := Start(pluginCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(Exit(0))

			Expect(session.Out).To(Say("v0.0.0"))
		})
	})

	Describe("using optional flags", func() {
		When("passed the long flag", func() {
			It("should print the current version in long format", func() {
				pluginCommand := exec.Command(
					_pluginPath, "version", "--long",
				)

				session, err := Start(pluginCommand, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(Say("version: v0.0.0"))
				Expect(session.Out).To(Say("commit: abcdefg"))
				Expect(session.Out).To(Say("date: 0000-00-00T00:00:00"))
				Expect(session.Out).To(Say("built by: test-suite"))
			})
		})
	})
})
