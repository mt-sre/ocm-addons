package tools

import (
	"context"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// ApplyEnv makes environment variable application a distinct step.
func ApplyEnv(env map[string]string) func(string, ...string) error {
	return func(cmd string, args ...string) error {
		return sh.RunWith(env, cmd, args...)
	}
}

// GoTimeoutFlag returns a timeout flag and duration corresponding
// to the given context deadline as a slice of strings. If the context
// has no deadline then an empty slice of strings is returned.
func GoTimeoutFlag(ctx context.Context) []string {
	deadline, ok := ctx.Deadline()
	if !ok {
		return []string{}
	}

	timeout := time.Until(deadline)

	return []string{"--timeout", timeout.String()}
}

// GoVerboseFlag checks if 'Mage' was run with the verbose flag and
// if so returns the Go verbose flag '-v' as a slice of strings.
// An empty slice of strings is returned otherwise.
func GoVerboseFlag() []string {
	if !mg.Verbose() {
		return []string{}
	}

	return []string{"-v"}
}
