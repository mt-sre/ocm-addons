package notify

import (
	"fmt"
	"testing"

	"github.com/mt-sre/ocm-addons/internal/testutil"
	"github.com/spf13/cobra"
)

func TestCmdArguments(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		command     *cobra.Command
		args        []string
		expectation string
		reports     []interface{}
	}{
		"less than 2 arguments": {
			command:     mockCommand(),
			expectation: fmt.Sprintf("requires at least %d arg(s), only received 0", _numArgs),
			reports:     []interface{}{fmt.Sprintf("should fail expecting %d args", _numArgs)},
		},
		"two or more arguments": {
			command: mockCommand(),
			args:    []string{"fake-team", "fake-notification-id"},
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
	return generateCommand(testutil.NoOp)
}
