package cli

import (
	"strings"

	"github.com/openshift-online/ocm-cli/pkg/config"
)

// LoadConfig loads an existing 'ocm.json' file and returns a Config object
// if possible.
func LoadConfig() (Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return Config{}, err
	}

	return Config{
		cfg: cfg,
	}, nil
}

// Config wraps an 'ocm-cli' config object and it's fields/methods.
type Config struct {
	cfg *config.Config
}

// IsEmpty returns true if there is no configuration file or if no
// credentials are present with which to start a session.
func (c Config) IsEmpty() bool {
	armed, reason, err := c.cfg.Armed()
	if err != nil {
		return true
	}

	if armed || strings.Contains(reason, "expired") {
		return false
	}

	return true
}

// ClientID returns the user's client id as configured.
func (c Config) ClientID() string {
	return c.cfg.ClientID
}

// Pager returns the configured paging application to pipe output to.
func (c Config) Pager() string {
	return c.cfg.Pager
}

// URL returns the configured OCM URL.
func (c Config) URL() string {
	return c.cfg.URL
}
