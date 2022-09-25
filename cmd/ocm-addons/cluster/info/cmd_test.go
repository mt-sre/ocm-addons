// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/testutil"
	"github.com/spf13/cobra"
)

func TestCmdArguments(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		command     *cobra.Command
		args        []string
		expectation string
		reports     []interface{}
	}{
		"no arguments": {
			command:     mockCommand(),
			expectation: "requires at least 1 arg(s), only received 0",
			reports:     []interface{}{"should report missing argument"},
		},
		"single argument": {
			command: mockCommand(),
			args:    []string{"fake-cluster-name"},
			reports: []interface{}{"should execute successfully"},
		},
	}

	for name, test := range testCases {
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
		"no headers flag": {
			command: mockCommand(),
			args:    []string{"--no-headers", "fake-cluster-name"},
			reports: []interface{}{"should execute successfully"},
		},
		"no color flag": {
			command: mockCommand(),
			args:    []string{"--no-color", "fake-cluster-name"},
			reports: []interface{}{"should execute successfully"},
		},
		"columns flag with no arguments": {
			command:     mockCommand(),
			args:        []string{"--columns"},
			expectation: "flag needs an argument: --columns",
			reports:     []interface{}{"should report missing command argument"},
		},
		"columns flag with single argument": {
			command: mockCommand(),
			args: []string{
				"--columns", "column1,column2,column3",
				"fake-cluster-name",
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
