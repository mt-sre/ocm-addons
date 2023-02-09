// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"fmt"
	"os"
	"time"

	"github.com/apex/log"

	"github.com/blang/semver/v4"
	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/meta"
	"github.com/mt-sre/ocm-addons/internal/scm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	const (
		organization = "mt-sre"
		repository   = "ocm-addons"
		binary       = "ocm-addons"
	)

	vu := scm.NewGitHubClient(
		scm.WithOrganization(organization),
		scm.WithRepository(repository),
		scm.WithTargetBinary(binary),
	)

	var opts options

	return generateCommand(&opts, run(vu, &opts))
}

type options struct{}

func generateCommand(_ *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "updates the plug-in to the latest version",
		Args:  cobra.NoArgs,
		RunE:  run,
	}

	return cmd
}

func run(vu cli.VersionUpdater, _ *options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var (
			ctx = cmd.Context()
			in  = cmd.InOrStdin()
			out = cmd.OutOrStdout()
		)

		latest, err := vu.GetLatestVersion(ctx)
		if err != nil {
			return fmt.Errorf("getting latest version: %w", err)
		}

		current := meta.Version()

		log.Debug(fmt.Sprintf("Found current version %s\n", current))
		log.Debug(fmt.Sprintf("Found latest version %s\n", latest))

		if upToDate(current, latest) {
			fmt.Fprintf(out, "The current version %s is already up-to-date.\n", current)

			return nil
		}

		if !cli.PromptYesOrNo(out, in, fmt.Sprintf("Would you like to update to version %s?", latest)) {
			return nil
		}

		data, err := vu.GetLatestPluginBinary(ctx)
		if err != nil {
			return fmt.Errorf("retrieving latest plugin binary: %w", err)
		}

		bin, err := os.Executable()
		if err != nil {
			return fmt.Errorf("getting binary path: %w", err)
		}

		backup := bin + "_" + time.Now().Format("2006.01.02_15:04:05")

		if err := os.Rename(bin, backup); err != nil {
			return fmt.Errorf("backing up binary: %w", err)
		}

		const perms = os.FileMode(0o755)

		if err := os.WriteFile(bin, data, perms); err != nil {
			if err := os.Rename(backup, bin); err != nil {
				log.Error(fmt.Sprintf("restoring binary: %v", err))
			}

			return fmt.Errorf("writing to file %q: %w", bin, err)
		}

		if err := os.Remove(backup); err != nil {
			return fmt.Errorf("cleaning up old binary: %w", err)
		}

		return nil
	}
}

func upToDate(current, latest string) bool {
	curVer, err := semver.ParseTolerant(current)
	if err != nil {
		return false
	}

	latestVer, err := semver.ParseTolerant(latest)
	if err != nil {
		return true
	}

	return latestVer.LTE(curVer)
}
