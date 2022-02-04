//go:build mage
// +build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mt-sre/ocm-addons/internal/tools"
)

var Aliases = map[string]interface{}{
	"check":   All.Check,
	"clean":   All.Clean,
	"test":    All.Test,
	"install": Build.Install,
}

type All mg.Namespace

// Runs all checks.
func (All) Check(ctx context.Context) {
	mg.SerialCtxDeps(
		ctx,
		Check.Tidy,
		Check.Verify,
		Check.Lint,
		Check.License,
	)
}

// Cleans all artifacts.
func (All) Clean() {
	mg.Deps(
		Deps.Clean,
		Build.Clean,
		Release.Clean,
	)
}

// Runs all tests.
func (All) Test(ctx context.Context) {
	mg.CtxDeps(
		ctx,
		Test.Unit,
		Test.Integration,
	)
}

var _depBin = path.Join(_dependencyDir, "bin")

var _dependencyDir = func() string {
	if dir, ok := os.LookupEnv("DEPENDENCY_DIR"); ok {
		return dir
	}

	return path.Join(_projectRoot, ".cache", "dependencies")
}()

var _projectRoot = func() string {
	if root, ok := os.LookupEnv("PROJECT_ROOT"); ok {
		return root
	}

	root, err := sh.Output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		panic("failed to get working directory")
	}

	return root
}()

const lockFile = ".mage.lock"

func dependencies() tools.DependencyManifest {
	return tools.DependencyManifest{
		Items: []tools.Dependency{
			{
				Name:    "lichen",
				Version: "v0.1.4",
				Module:  "github.com/uw-labs/lichen",
			}, {
				Name:    "golangci-lint",
				Version: "v1.43.0",
				Module:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
			}, {
				Name:    "ginkgo",
				Version: "v1.16.4",
				Module:  "github.com/onsi/ginkgo/ginkgo",
			}, {
				Name:    "ocm",
				Version: "latest",
				Module:  "github.com/openshift-online/ocm-cli/cmd/ocm",
			}, {
				Name:    "goreleaser",
				Version: "v1.4.1",
				Module:  "github.com/goreleaser/goreleaser",
			},
		},
	}
}

type Deps mg.Namespace

// Updates any dependent binaries in the cache.
func (Deps) Update() error {
	if err := os.MkdirAll(_dependencyDir, 0o774); err != nil {
		return err
	}

	deps := dependencies()

	cfg, err := tools.LoadLock(lockFile)
	if errors.Is(err, os.ErrNotExist) {
		if err := deps.InstallAll(_depBin); err != nil {
			return err
		}

		return tools.DumpLock(
			lockFile, tools.NewLock(tools.LockDependencies(dependencies())),
		)
	} else if err != nil {
		return err
	}

	if err := cfg.Dependencies.Install(_depBin, deps.Difference(cfg.Dependencies)...); err != nil {
		return err
	}

	return tools.DumpLock(lockFile, tools.NewLock(tools.LockDependencies(dependencies())))
}

// Removes any existing dependency binaries
func (Deps) Clean() error {
	return sh.Rm(_depBin)
}

type Check mg.Namespace

// Runs linter against source code.
func (Check) Lint(ctx context.Context) error {
	mg.Deps(Deps.Update)

	args := []string{"run"}
	args = append(args, tools.GoVerboseFlag()...)
	args = append(args, tools.GoTimeoutFlag(ctx)...)

	out, err := sh.Output(path.Join(_depBin, "golangci-lint"), args...)

	fmt.Print(out)

	return err
}

const binOut = "ocm-addons"

var _binDir = path.Join(_projectRoot, "bin")

// Scans imported go packages and ensures they are compatible with
// this repository's license (Apache 2.0).
func (Check) License() error {
	mg.Deps(
		Build.Install,
		Deps.Update,
	)

	lichenConfig := ".lichen.yaml"

	return sh.Run(
		path.Join(_depBin, "lichen"),
		"-c", path.Join(_projectRoot, lichenConfig),
		path.Join(_binDir, binOut))
}

// Ensures dependencies are correctly updated in the 'go.mod'
// and 'go.sum' files.
func (Check) Tidy() error {
	return sh.Run("go", "mod", "tidy")
}

// Ensures package dependencies have not been tampered with since download.
func (Check) Verify() error {
	return sh.Run("go", "mod", "verify")
}

type Build mg.Namespace

// Copies plug-in binary to "$GOPATH/bin".
// If "$GOPATH/bin" is in the PATH then the plug-in
// can be invoked using "ocm addons"
func (Build) Install() error {
	mg.Deps(Build.Plugin)

	gopath, err := sh.Output(mg.GoCmd(), "env", "GOPATH")
	if err != nil {
		return fmt.Errorf("GOPATH cannot be found: %w", err)
	}

	return sh.Copy(path.Join(gopath, "bin", binOut), path.Join(_binDir, binOut))
}

// Compiles top-level 'ocm-addons' command as an executable binary.
// The binary can be used stand-alone or added to a directory in
// the system PATH to work as a plug-in with 'ocm'.
func (Build) Plugin() error {
	mg.Deps(Build.Clean)

	var goVars = map[string]string{
		"CGO_ENABLED": "0",
	}

	runWithGoVars := tools.ApplyEnv(goVars)

	return runWithGoVars(
		"go", "build",
		"-o", path.Join(_binDir, binOut),
		"./cmd/ocm-addons",
	)
}

// Removes built binaries if they already exist.
func (Build) Clean() error {
	return sh.Rm(path.Join(_binDir, binOut))
}

type Release mg.Namespace

// Generates release artifacts and pushes to SCM.
func (Release) Full() error {
	mg.Deps(
		Deps.Update,
		Release.Clean,
	)

	return sh.Run(path.Join(_depBin, "goreleaser"), "release")
}

// Generates release artifacts locally.
func (Release) Snapshot() error {
	mg.Deps(
		Deps.Update,
		Release.Clean,
	)

	return sh.Run(path.Join(_depBin, "goreleaser"), "release", "--snapshot")
}

func (Release) Clean() error {
	return sh.Rm(path.Join(_projectRoot, "dist"))
}

type Test mg.Namespace

// Runs unit tests.
func (Test) Unit(ctx context.Context) error {
	args := []string{"test", "-race"}

	args = append(args, tools.GoVerboseFlag()...)
	args = append(args, tools.GoTimeoutFlag(ctx)...)

	targetDirs := []string{
		"./cmd",
		"./internal",
	}
	for _, dir := range targetDirs {
		args = append(args, fmt.Sprintf("%s/...", dir))
	}

	return sh.Run("go", args...)
}

// Runs integration tests.
func (Test) Integration(ctx context.Context) error {
	mg.Deps(Deps.Update)

	args := []string{
		"-r",
		"--randomizeAllSpecs",
		"--randomizeSuites",
		"--failOnPending",
		"--keepGoing",
		"--race",
		"--trace",
	}
	args = append(args, tools.GoVerboseFlag()...)
	args = append(args, tools.GoTimeoutFlag(ctx)...)
	args = append(args, "integration")

	return sh.Run(path.Join(_depBin, "ginkgo"), args...)
}
