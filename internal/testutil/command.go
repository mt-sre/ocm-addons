package testutil

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func NoOp(cmd *cobra.Command, argv []string) error {
	return nil
}

func NewCommandAssertion(t *testing.T, opts ...CommandAssertionOption) {
	t.Helper()

	var assertion CommandAssertion

	for _, opt := range opts {
		opt(&assertion)
	}

	assertion.asserter = require.New(t)

	if assertion.command == nil {
		panic("a target command must be defined")
	}

	if len(assertion.args) > 0 {
		assertion.command.SetArgs(assertion.args)
	}

	// To avoid polluting test output with usage info when tests pass
	assertion.command.SetOut(io.Discard)
	assertion.command.SetErr(io.Discard)

	err := assertion.command.Execute()

	if assertion.expectation == "" {
		assertion.asserter.Nil(err, assertion.report...)

		return
	}

	assertion.asserter.Contains(err.Error(), assertion.expectation, assertion.report...)
}

type CommandAssertion struct {
	asserter    *require.Assertions
	command     *cobra.Command
	args        []string
	expectation string
	report      []interface{}
}

type CommandAssertionOption func(c *CommandAssertion)

func CommandAssertionCommand(command *cobra.Command) CommandAssertionOption {
	return func(c *CommandAssertion) {
		c.command = command
	}
}

func CommandAssertionArgs(args ...string) CommandAssertionOption {
	return func(c *CommandAssertion) {
		c.args = args
	}
}

func CommandAssertionExpectation(expectation string) CommandAssertionOption {
	return func(c *CommandAssertion) {
		c.expectation = expectation
	}
}

func CommandAssertionReports(report ...interface{}) CommandAssertionOption {
	return func(c *CommandAssertion) {
		c.report = report
	}
}
