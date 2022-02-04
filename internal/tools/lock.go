package tools

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/magefile/mage/sh"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

func LoadLock(path string) (*Lock, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening lock: %w", err)
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading lock: %w", err)
	}

	var cfg Lock

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling lock: %w", err)
	}

	return &cfg, nil
}

func DumpLock(path string, cfg Lock) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening lock: %w", err)
	}

	defer file.Close()

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling lock: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("writing lock: %w", err)
	}

	return nil
}

func NewLock(opts ...LockOption) Lock {
	var lock Lock

	for _, opt := range opts {
		opt(&lock)
	}

	return lock
}

type Lock struct {
	Dependencies DependencyManifest
}

type LockOption func(*Lock)

func LockDependencies(man DependencyManifest) LockOption {
	return func(l *Lock) {
		l.Dependencies = man
	}
}

type DependencyManifest struct {
	Items []Dependency
}

func (m *DependencyManifest) InstallAll(depRoot string) error {
	deps := make([]string, 0, len(m.Items))

	for _, dep := range m.Items {
		deps = append(deps, dep.Name)
	}

	return m.Install(depRoot, deps...)
}

func (m *DependencyManifest) Install(binDir string, names ...string) error {
	var errCollector error

	for _, dep := range m.Items {
		if dep.Exists(binDir) && !Contains(dep.Name, names) {
			continue
		}

		if err := dep.Install(binDir); err != nil {
			multierr.AppendInto(&errCollector, err)
		}
	}

	return errCollector
}

func (m *DependencyManifest) Difference(other DependencyManifest) []string {
	result := make([]string, 0, len(m.Items))

	for _, dep := range m.Items {
		if other.Contains(dep) {
			continue
		}

		result = append(result, dep.Name)
	}

	return result
}

func (m *DependencyManifest) Contains(dep Dependency) bool {
	for _, d := range m.Items {
		if d.Equals(dep) {
			return true
		}
	}

	return false
}

type Dependency struct {
	Name    string
	Version string
	Module  string
}

func (d *Dependency) Exists(binDir string) bool {
	_, err := os.Stat(path.Join(binDir, d.Name))

	return !errors.Is(err, os.ErrNotExist)
}

func (d *Dependency) Install(binDir string) error {
	env := map[string]string{
		"GOBIN": binDir,
	}

	return sh.RunWith(env,
		"go", "install", "-mod=readonly", fmt.Sprintf("%s@%s", d.Module, d.Version),
	)
}

func (d *Dependency) Equals(other Dependency) bool {
	return d.Name == other.Name &&
		d.Version == other.Version &&
		d.Module == other.Module
}

func Contains(elem string, slice []string) bool {
	for _, s := range slice {
		if elem == s {
			return true
		}
	}

	return false
}
