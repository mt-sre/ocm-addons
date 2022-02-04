package tools

import (
	"fmt"
	"io"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type ContainerError string

func (e ContainerError) Error() string {
	return fmt.Sprintf("command: %s", string(e))
}

type Command func(...string) error

func runCommand(env map[string]string, out io.Writer, runtime string, args ...string) Command {
	return func(innerArgs ...string) error {
		fullArgs := args
		fullArgs = append(fullArgs, innerArgs...)

		if out == nil && mg.Verbose() {
			out = os.Stdout
		}

		_, err := sh.Exec(
			env,
			out, os.Stderr,
			runtime, fullArgs...,
		)

		return err
	}
}

// NewBuildCommand returns a function which will build a container image with the given arguments.

// contextDirectory is the "build" directory from which the container runtime will observe
// as the relative root location.
// options is a variadic slice of additional options to apply to the container build command.
// An error will be returned if any invalid parameters are supplied or there is
// no available container runtime on the system.
func NewBuildCommand(contextDirectory string, options ...BuildOption) (Command, error) {
	if contextDirectory == "" {
		return nil, ContainerError("context directory not defined")
	}

	cmd := &BuildCommand{}

	for _, opt := range options {
		err := opt(cmd)
		if err != nil {
			return nil, ContainerError(err.Error())
		}
	}

	args := cmd.compileArgs()
	args = append(args, contextDirectory)

	if cmd.runtime == "" {
		runtime, ok := Runtime()
		if !ok {
			return nil, ContainerError("no container runtime found")
		}

		cmd.runtime = runtime
	}

	return runCommand(cmd.env, nil, cmd.runtime, args...), nil
}

// BuildCommand is used to generate a command which will
// build a container image.
type BuildCommand struct {
	containerFilePath string
	env               map[string]string
	buildArgs         []string
	runtime           string
	tags              []string
	io.Writer
}

type BuildOption func(c *BuildCommand) error

func (c *BuildCommand) compileArgs() []string {
	args := []string{"build"}

	if len(c.buildArgs) > 0 {
		for _, arg := range c.buildArgs {
			args = append(args, "--build-arg", arg)
		}
	}

	if len(c.tags) > 0 {
		for _, tag := range c.tags {
			args = append(args, "-t", tag)
		}
	}

	if c.containerFilePath != "" {
		args = append(args, "-f", c.containerFilePath)
	}

	return args
}

// BuildArgs applies build arguments for variable substitution within a container file.
func BuildArgs(args ...string) BuildOption {
	return func(c *BuildCommand) error {
		c.buildArgs = args

		return nil
	}
}

// BuildEnv adds the supplied map of variable/value pairs to the commands
// run environment.
func BuildEnv(env map[string]string) BuildOption {
	return func(c *BuildCommand) error {
		c.env = env

		return nil
	}
}

// BuildRuntime manually sets the desired container runtime.
func BuildRuntime(runtime string) BuildOption {
	return func(c *BuildCommand) error {
		c.runtime = runtime

		return nil
	}
}

// BuildTags applies tags to the build image.
func BuildTags(tags ...string) BuildOption {
	return func(c *BuildCommand) error {
		c.tags = tags

		return nil
	}
}

// BuildContainerFilepath is a path to a container file containing build
// instructions.
func BuildContainerFilePath(path string) BuildOption {
	return func(c *BuildCommand) error {
		c.containerFilePath = path

		return nil
	}
}

// NewRunCommand returns a function which will run a container with optional arguments.
//
// image is a mandatory image name for the container to run in.
// options is a variadic slice of additional options to apply to the container run command.
// An error will be returned if any invalid parameters are supplied or there is
// no available container runtime on the system.
func NewRunCommand(image string, options ...RunOption) (Command, error) {
	if image == "" {
		return nil, ContainerError("no container image was set")
	}

	cmd := &RunCommand{}

	for _, opt := range options {
		err := opt(cmd)
		if err != nil {
			return nil, ContainerError(err.Error())
		}
	}

	if cmd.runtime == "" {
		runtime, ok := Runtime()
		if !ok {
			return nil, ContainerError("no container runtime found")
		}

		cmd.runtime = runtime
	}

	args := cmd.compileArgs()
	args = append(args, image)

	return runCommand(cmd.env, cmd.out, cmd.runtime, args...), nil
}

type RunOption func(*RunCommand) error

// RunCommand wraps a command which runs a container.
type RunCommand struct {
	disableSELinux bool
	env            map[string]string
	mounts         []Mount
	out            io.Writer
	removeAfter    bool
	runtime        string
	workingDir     string
}

func (c *RunCommand) compileArgs() []string {
	args := []string{"run"}

	if c.removeAfter {
		args = append(args, "--rm")
	}

	if len(c.mounts) > 0 {
		for _, m := range c.mounts {
			args = append(args, "--mount", m.String())
		}
	}

	if c.disableSELinux {
		args = append(args, "--security-opt", "label=disable")
	}

	if c.workingDir != "" {
		args = append(args, "-w", c.workingDir)
	}

	return args
}

// RunDisableSELinux sets "--security-opt label=disable" to bypass setting
// a container policy.
var RunDisableSELinux RunOption = func(c *RunCommand) error {
	c.disableSELinux = true

	return nil
}

// RunEnv adds the supplied map of variable/value pairs to the commands
// run environment.
func RunEnv(env map[string]string) RunOption {
	return func(c *RunCommand) error {
		c.env = env

		return nil
	}
}

// RunMounts adds the supplied mounts to the container when run.
func RunMounts(mounts ...Mount) RunOption {
	return func(c *RunCommand) error {
		c.mounts = mounts

		return nil
	}
}

// Out redirects output from the run command to the supplied io.Writer.
func RunOut(out io.Writer) RunOption {
	return func(c *RunCommand) error {
		c.out = out

		return nil
	}
}

// RunRemoveAfter ensures the container is removed after it exits.
var RunRemoveAfter RunOption = func(c *RunCommand) error {
	c.removeAfter = true

	return nil
}

// RunRuntime manually sets the desired container runtime.
func RunRuntime(runtime string) RunOption {
	return func(c *RunCommand) error {
		c.runtime = runtime

		return nil
	}
}

// RunWorkingDir sets a working directory within the container.
func RunWorkingDir(workingDir string) RunOption {
	return func(c *RunCommand) error {
		c.workingDir = workingDir

		return nil
	}
}
