/* #nosec */

package integration

import (
	"bytes"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update subcommand", func() {
	Describe("using no flags", func() {
		It("should prompt for update", func() {
			pluginCommand := exec.Command(
				_pluginPath, "update",
			)

			// reject prompt to update binary
			pluginCommand.Stdin = bytes.NewBufferString("n")

			session, err := Start(pluginCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(Exit(0))

			Expect(session.Out).To(Say(`Would you like to update to version v\d+\.\d+\.\d+\?\s+\(y/n\)\:`))
		})
	})
})
