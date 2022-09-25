// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/testutil"
	"github.com/spf13/cobra"
)

func TestCmdArgs(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		command     *cobra.Command
		args        []string
		expectation string
		reports     []interface{}
	}{
		"no arguments": {
			command: mockCommand(),
			reports: []interface{}{"should execute successfully"},
		},
		"one or more arguments": {
			command:     mockCommand(),
			args:        []string{"fake-arg"},
			expectation: "unknown command",
			reports:     []interface{}{"should fail expecting no args"},
		},
	}

	for name, test := range testcases {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testutil.NewCommandAssertion(
				t,
				testutil.CommandAssertionCommand(test.command),
				testutil.CommandAssertionArgs(test.args...),
				testutil.CommandAssertionExpectation(test.expectation),
				testutil.CommandAssertionReports(test.reports...),
			)
		})
	}
}

func TestCmdOptions(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		command     *cobra.Command
		args        []string
		expectation string
		reports     []interface{}
	}{
		"long flag": {
			command: mockCommand(),
			args: []string{
				"--long",
			},
			reports: []interface{}{"should execute successfully"},
		},
	}

	for name, test := range testcases {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testutil.NewCommandAssertion(
				t,
				testutil.CommandAssertionCommand(test.command),
				testutil.CommandAssertionArgs(test.args...),
				testutil.CommandAssertionExpectation(test.expectation),
				testutil.CommandAssertionReports(test.reports...),
			)
		})
	}
}

func mockCommand() *cobra.Command {
	return generateCommand(new(options), testutil.NoOp)
}
