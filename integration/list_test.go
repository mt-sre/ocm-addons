/* #nosec */

package integration

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var _ = Describe("list subcommand", func() {
	var ocm *OCMEnvironment

	BeforeEach(func() {
		var err error

		ocm, err = setupEnv()
		Expect(err).ToNot(HaveOccurred())

		DeferCleanup(ocm.CleanUp)

		ocmCommand := exec.Command(
			_ocmBinary, "login",
			"--client-id", "my-client",
			"--client-secret", "my-secret",
			"--token-url", ocm.SSOServerURL(),
			"--url", ocm.APIServerURL(),
		)

		ocmCommand.Env = append(ocmCommand.Env, fmt.Sprintf("OCM_CONFIG=%s", ocm.Config()))

		session, err := Start(ocmCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(Exit(0))
	})

	Describe("using optional flags", func() {
		When("passed the no-headers flag", func() {
			It("should not write headers", func() {
				pluginCommand := exec.Command(
					_pluginPath, "list",
					"--no-headers",
				)

				pluginCommand.Env = append(pluginCommand.Env, fmt.Sprintf("OCM_CONFIG=%s", ocm.Config()))

				session, err := Start(pluginCommand, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(0))

				for _, addon := range ocm.Addons() {
					Expect(session.Out).To(MatchRow(addon.ID(), addon.Name(), fmt.Sprint(addon.Enabled())))
				}
			})
		})

		When("passed the columns flag", func() {
			It("should only write the specified columns", func() {
				pluginCommand := exec.Command(
					_pluginPath, "list",
					"--columns", "id, enabled",
				)

				pluginCommand.Env = append(pluginCommand.Env, fmt.Sprintf("OCM_CONFIG=%s", ocm.Config()))

				session, err := Start(pluginCommand, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(MatchRow("ID", "ENABLED"))

				for _, addon := range ocm.Addons() {
					Expect(session.Out).To(MatchRow(addon.ID(), fmt.Sprint(addon.Enabled())))
				}
			})
		})
	})

	Describe("using default behavior", func() {
		Context("with 100 addons", func() {
			It("should write multiple pages of addons as a table", func() {
				pluginCommand := exec.Command(
					_pluginPath, "list",
				)

				pluginCommand.Env = append(pluginCommand.Env, fmt.Sprintf("OCM_CONFIG=%s", ocm.Config()))

				session, err := Start(pluginCommand, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(MatchRow("ID", "NAME", "ENABLED"))

				for _, addon := range ocm.Addons() {
					Expect(session.Out).To(MatchRow(addon.ID(), addon.Name(), fmt.Sprint(addon.Enabled())))
				}
			})
		})
	})
})

func setupEnv() (*OCMEnvironment, error) {
	addons, err := generateAddOns(100)
	if err != nil {
		return nil, err
	}

	return NewOCMEnvironment(
		OCMEnvironmentAddons(addons...),
	)
}

func generateAddOns(num int) ([]*cmv1.AddOn, error) {
	addons := make([]*cmv1.AddOn, 0, num)

	for i := 1; i <= num; i++ {
		addon, err := GenerateAddOn(
			AddOnID(fmt.Sprintf("test-addon-%d", i)),
			AddOnName(fmt.Sprintf("Test Addon %d", i)),
			AddOnEnabled,
		)
		if err != nil {
			return nil, err
		}

		addons = append(addons, addon)
	}

	return addons, nil
}
