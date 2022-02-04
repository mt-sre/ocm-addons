package tools

import (
	"fmt"
	"strings"
)

// NewMount returns a mount with the supplied options applied.
// An error is returned if any of the options are invalid.
func NewMount(options ...MountOption) (Mount, error) {
	mount := Mount{
		mountType: "bind",
	}

	err := mount.Options(options...)
	if err != nil {
		return mount, err
	}

	return mount, nil
}

// Mount abstracts a volume mount used with container commands.
type Mount struct {
	readOnly  bool
	source    string
	target    string
	mountType string
}

type MountOption func(m *Mount) error

// Options applies the given options to a mount.
// If any options are invalid an error is returned.
func (m *Mount) Options(options ...MountOption) error {
	for _, opt := range options {
		if err := opt(m); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mount) String() string {
	var pairs []string

	pairs = append(pairs,
		fmt.Sprintf("type=%s", m.mountType),
		fmt.Sprintf("source=%s", m.source),
		fmt.Sprintf("target=%s", m.target),
	)

	if m.readOnly {
		pairs = append(pairs, "readonly=true")
	}

	return strings.Join(pairs, ",")
}

// MountReadOnly sets the mount to be "readonly" within the container.
var MountReadOnly MountOption = func(m *Mount) error {
	m.readOnly = true

	return nil
}

// MountSource sets the mount source location.
// The nature of this value changes with the mount type.
func MountSource(source string) MountOption {
	return func(m *Mount) error {
		m.source = source

		return nil
	}
}

// MountTarget sets the mount target within the container.
func MountTarget(target string) MountOption {
	return func(m *Mount) error {
		m.target = target

		return nil
	}
}

// MountType sets the mount type.
func MountType(mountType string) MountOption {
	return func(m *Mount) error {
		m.mountType = mountType

		return nil
	}
}
