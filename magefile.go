//go:build mage
// +build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mt-sre/go-ci/command"
	"github.com/mt-sre/ocm-addons/internal/tools"
)

var Aliases = map[string]interface{}{
	"check":       All.Check,
	"clean":       All.Clean,
	"install":     Build.Install,
	"release":     Release.Full,
	"run-hooks":   Hooks.Run,
	"test":        All.Test,
	"update-deps": All.UpdateDependencies,
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

func (All) UpdateDependencies(ctx context.Context) {
	mg.CtxDeps(
		ctx,
		Deps.UpdateLichen,
		Deps.UpdateGinkgo,
		Deps.UpdateGolangCILint,
		Deps.UpdateGoReleaser,
		Deps.UpdateOCMCLI,
		Deps.UpdatePreCommit,
	)
}

var _depBin = filepath.Join(_dependencyDir, "bin")

var _dependencyDir = func() string {
	if dir, ok := os.LookupEnv("DEPENDENCY_DIR"); ok {
		return dir
	}

	return filepath.Join(_projectRoot, ".cache", "dependencies")
}()

var _projectRoot = func() string {
	if root, ok := os.LookupEnv("PROJECT_ROOT"); ok {
		return root
	}

	topLevel := git(command.WithArgs{"rev-parse", "--show-toplevel"})

	if err := topLevel.Run(); err != nil || !topLevel.Success() {
		panic("failed to get working directory")
	}

	return strings.TrimSpace(topLevel.Stdout())
}()

var git = command.NewCommandAlias("git")

type Deps mg.Namespace

func (Deps) UpdateGinkgo(ctx context.Context) error {
	return updateGODependency(ctx, "github.com/onsi/ginkgo/v2/ginkgo")
}

func (Deps) UpdateGolangCILint(ctx context.Context) error {
	return updateGODependency(ctx, "github.com/golangci/golangci-lint/cmd/golangci-lint")
}

func (Deps) UpdateGoReleaser(ctx context.Context) error {
	return updateGODependency(ctx, "github.com/goreleaser/goreleaser")
}

func (Deps) UpdateLichen(ctx context.Context) error {
	return updateGODependency(ctx, "github.com/uw-labs/lichen")
}

func (Deps) UpdateOCMCLI(ctx context.Context) error {
	return updateGODependency(ctx, "github.com/openshift-online/ocm-cli/cmd/ocm")
}

func updateGODependency(ctx context.Context, src string) error {
	if err := setupDepsBin(); err != nil {
		return fmt.Errorf("creating dependencies bin directory: %w", err)
	}

	toolsDir := filepath.Join(_projectRoot, "tools")

	tidy := gocmd(
		command.WithArgs{"mod", "tidy"},
		command.WithWorkingDirectory(toolsDir),
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := tidy.Run(); err != nil {
		return fmt.Errorf("starting to tidy tools dir: %w", err)
	}

	if !tidy.Success() {
		return fmt.Errorf("tidying tools dir: %w", tidy.Error())
	}

	install := gocmd(
		command.WithArgs{"install", src},
		command.WithWorkingDirectory(toolsDir),
		command.WithCurrentEnv(true),
		command.WithEnv{"GOBIN": _depBin},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := install.Run(); err != nil {
		return fmt.Errorf("starting to install command from source %q: %w", src, err)
	}

	if !install.Success() {
		return fmt.Errorf("installing command from source %q: %w", src, install.Error())
	}

	return nil
}

var gocmd = command.NewCommandAlias(mg.GoCmd())

func (Deps) UpdatePreCommit(ctx context.Context) error {
	if err := setupDepsBin(); err != nil {
		return fmt.Errorf("creating dependencies bin directory: %w", err)
	}

	const urlPrefix = "https://github.com/pre-commit/pre-commit/releases/download"

	// pinning to version 2.17.0 since 2.18.0+ requires python>=3.7
	const version = "2.17.0"

	out := filepath.Join(_depBin, "pre-commit")

	if _, err := os.Stat(out); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspecting output location %q: %w", out, err)
		}

		if err := tools.DownloadFile(ctx, urlPrefix+fmt.Sprintf("/v%s/pre-commit-%s.pyz", version, version), out); err != nil {
			return fmt.Errorf("downloading pre-commit: %w", err)
		}
	}

	return os.Chmod(out, 0775)
}

func setupDepsBin() error {
	return os.MkdirAll(_depBin, 0o774)
}

// Removes any existing dependency binaries
func (Deps) Clean() error {
	return sh.Rm(_depBin)
}

type Check mg.Namespace

// Runs linter against source code.
func (Check) Lint(ctx context.Context) error {
	mg.CtxDeps(
		ctx,
		Deps.UpdateGolangCILint,
	)

	run := golancilint(
		command.WithArgs{"run"},
		command.WithArgs(tools.GoVerboseFlag()),
		command.WithContext{Context: ctx},
	)

	if err := run.Run(); err != nil {
		return fmt.Errorf("starting linter: %w", err)
	}

	if run.Success() {
		return nil
	}

	fmt.Fprint(os.Stdout, run.CombinedOutput())

	return fmt.Errorf("running linter: %w", run.Error())
}

var golancilint = command.NewCommandAlias(filepath.Join(_depBin, "golangci-lint"))

const binOut = "ocm-addons"

var _binDir = filepath.Join(_projectRoot, "bin")

// Scans imported go packages and ensures they are compatible with
// this repository's license (Apache 2.0).
func (Check) License(ctx context.Context) error {
	mg.CtxDeps(
		ctx,
		Build.Plugin,
		Deps.UpdateLichen,
	)

	lichenConfig := ".lichen.yaml"

	licenseCheck := lichen(
		command.WithArgs{
			"-c", filepath.Join(_projectRoot, lichenConfig),
			filepath.Join(_binDir, binOut),
		},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := licenseCheck.Run(); err != nil {
		return fmt.Errorf("starting license check: %w", err)
	}

	if licenseCheck.Success() {
		return nil
	}

	return fmt.Errorf("running license check: %w", licenseCheck.Error())
}

var lichen = command.NewCommandAlias(filepath.Join(_depBin, "lichen"))

// Ensures dependencies are correctly updated in the 'go.mod'
// and 'go.sum' files.
func (Check) Tidy(ctx context.Context) error {
	tidy := gocmd(
		command.WithArgs{"mod", "tidy"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := tidy.Run(); err != nil {
		return fmt.Errorf("starting to tidy go dependencies: %w", err)
	}

	if tidy.Success() {
		return nil
	}

	return fmt.Errorf("tidying go dependencies: %w", tidy.Error())
}

// Ensures package dependencies have not been tampered with since download.
func (Check) Verify(ctx context.Context) error {
	verify := gocmd(
		command.WithArgs{"mod", "verify"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := verify.Run(); err != nil {
		return fmt.Errorf("starting to verify go dependencies: %w", err)
	}

	if verify.Success() {
		return nil
	}

	return fmt.Errorf("verifying go dependencies: %w", verify.Error())
}

type Build mg.Namespace

// Copies plug-in binary to "$GOPATH/bin".
// If "$GOPATH/bin" is in the PATH then the plug-in
// can be invoked using "ocm addons"
func (Build) Install(ctx context.Context) error {
	install := gocmd(
		command.WithArgs{
			"install", filepath.Join(_projectRoot, "cmd", "ocm-addons"),
		},
	)

	if err := install.Run(); err != nil {
		return fmt.Errorf("starting to install plugin: %w", err)
	}

	if install.Success() {
		return nil
	}

	return fmt.Errorf("installing plugin: %w", install.Error())
}

// Compiles top-level 'ocm-addons' command as an executable binary.
// The binary can be used stand-alone or added to a directory in
// the system PATH to work as a plug-in with 'ocm'.
func (Build) Plugin(ctx context.Context) error {
	mg.Deps(Build.Clean)

	build := gocmd(
		command.WithArgs{
			"build",
			"-o", filepath.Join(_binDir, binOut),
			filepath.Join(_projectRoot, "cmd", "ocm-addons"),
		},
		command.WithCurrentEnv(true),
		command.WithEnv{"CGO_ENABLED": "0"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := build.Run(); err != nil {
		return fmt.Errorf("starting to build plugin: %w", err)
	}

	if build.Success() {
		return nil
	}

	return fmt.Errorf("building plugin: %w", build.Error())
}

// Removes built binaries if they already exist.
func (Build) Clean() error {
	return sh.Rm(filepath.Join(_binDir, binOut))
}

type Release mg.Namespace

// Generates release artifacts and pushes to SCM.
func (Release) Full(ctx context.Context) error {
	mg.CtxDeps(
		ctx,
		Deps.UpdateGoReleaser,
		Release.Clean,
	)

	release := goreleaser(
		command.WithArgs{"release", "--rm-dist"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := release.Run(); err != nil {
		return fmt.Errorf("starting release: %w", err)
	}

	if release.Success() {
		return nil
	}

	return fmt.Errorf("releasing plugin: %w", release.Error())
}

// Generates release artifacts locally.
func (Release) Snapshot(ctx context.Context) error {
	mg.CtxDeps(
		ctx,
		Deps.UpdateGoReleaser,
		Release.Clean,
	)

	release := goreleaser(
		command.WithArgs{"release", "--snapshot"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := release.Run(); err != nil {
		return fmt.Errorf("starting release snapshot: %w", err)
	}

	if release.Success() {
		return nil
	}

	return fmt.Errorf("releasing snapshot: %w", release.Error())
}

var goreleaser = command.NewCommandAlias(filepath.Join(_depBin, "goreleaser"))

func (Release) Clean() error {
	return sh.Rm(filepath.Join(_projectRoot, "dist"))
}

type Test mg.Namespace

// Runs unit tests.
func (Test) Unit(ctx context.Context) error {
	test := gocmd(
		command.WithArgs{"test", "-race"},
		command.WithArgs(tools.GoVerboseFlag()),
		command.WithArgs{"./cmd/...", "./internal/..."},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := test.Run(); err != nil {
		return fmt.Errorf("starting unit tests: %w", err)
	}

	if test.Success() {
		return nil
	}

	return fmt.Errorf("running unit tests: %w", test.Error())
}

// Runs integration tests.
func (Test) Integration(ctx context.Context) error {
	mg.CtxDeps(
		ctx,
		Deps.UpdateGinkgo,
		Deps.UpdateOCMCLI,
	)

	test := ginkgo(
		command.WithArgs{
			"-r",
			"--randomize-all",
			"--randomize-suites",
			"--fail-on-pending",
			"--keep-going",
			"--race",
			"--trace",
		},
		command.WithArgs(tools.GoVerboseFlag()),
		command.WithArgs{"integration"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := test.Run(); err != nil {
		return fmt.Errorf("starting integration tests: %w", err)
	}

	if test.Success() {
		return nil
	}

	return fmt.Errorf("running integration tests: %w", test.Error())
}

var ginkgo = command.NewCommandAlias(filepath.Join(_depBin, "ginkgo"))

type Hooks mg.Namespace

func (Hooks) Enable(ctx context.Context) error {
	mg.CtxDeps(ctx, Deps.UpdatePreCommit)

	install := precommit(
		command.WithArgs{
			"install",
			"--hook-type", "pre-commit",
			"--hook-type", "pre-push",
		},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := install.Run(); err != nil {
		return fmt.Errorf("starting to install pre-commit hooks: %w", err)
	}

	if install.Success() {
		return nil
	}

	return fmt.Errorf("installing pre-commit hooks: %w", install.Error())
}

func (Hooks) Disable(ctx context.Context) error {
	mg.CtxDeps(ctx, Deps.UpdatePreCommit)

	uninstall := precommit(
		command.WithArgs{"uninstall"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := uninstall.Run(); err != nil {
		return fmt.Errorf("starting to disable hooks: %w", err)
	}

	if uninstall.Success() {
		return nil
	}

	return fmt.Errorf("disabling hooks: %w", uninstall.Error())
}

func (Hooks) Run(ctx context.Context) error {
	mg.CtxDeps(ctx, Deps.UpdatePreCommit)

	run := precommit(
		command.WithArgs{"run",
			"--show-diff-on-failure",
			"--from-ref", "origin/main", "--to-ref", "HEAD",
		},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := run.Run(); err != nil {
		return fmt.Errorf("starting to run hooks: %w", err)
	}

	if run.Success() {
		return nil
	}

	return fmt.Errorf("running hooks: %w", run.Error())
}

func (Hooks) RunAllFiles(ctx context.Context) error {
	mg.CtxDeps(ctx, Deps.UpdatePreCommit)

	runall := precommit(
		command.WithArgs{"run", "--all-files"},
		command.WithConsoleOut(mg.Verbose()),
		command.WithContext{Context: ctx},
	)

	if err := runall.Run(); err != nil {
		return fmt.Errorf("starting to run hooks for all files: %w", err)
	}

	if runall.Success() {
		return nil
	}

	return fmt.Errorf("running hooks for all files: %w", runall.Error())
}

var precommit = command.NewCommandAlias(filepath.Join(_depBin, "pre-commit"))
