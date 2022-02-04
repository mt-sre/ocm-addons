package version

import (
	"fmt"
	"os"

	"github.com/mt-sre/ocm-addons/internal/meta"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Cmd() *cobra.Command {
	var opts options

	return generateCommand(&opts, run(&opts))
}

type options struct {
	Long bool
}

func (o *options) AddLongFlag(flags *pflag.FlagSet) {
	flags.BoolVar(
		&o.Long,
		"long",
		o.Long,
		"outputs extended version information",
	)
}

func generateCommand(opts *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "returns the current version of this plug-in",
		Args:  cobra.NoArgs,
		RunE:  run,
	}

	flags := cmd.Flags()

	opts.AddLongFlag(flags)

	return cmd
}

func run(opts *options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		version := meta.Version()

		if opts.Long {
			version = meta.LongVersion()
		}

		_, err := fmt.Fprintln(os.Stdout, version)

		return err
	}
}
