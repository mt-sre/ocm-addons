// Package session wraps ocm session parameters to be accessed during cli
// commands. For example both the active connection and logger can be
// retrieved from the current session.
package cli

import (
	"errors"

	"github.com/apex/log"
	"github.com/openshift-online/ocm-cli/pkg/ocm"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

var ErrNoConfigurationLoaded = errors.New("no configuration loaded")

// NewSession loads an ocm configuration created after the user runs 'ocm login'
// and starts a connection to ocm. A logger with session context is also made
// available to callers of NewSession. An error is returned if a configuration
// cannot be loaded or if a connection cannot be made to OCM. Otherwise a
// pointer to a session object is returned.
func NewSession() (Session, error) {
	config, err := LoadConfig()
	if err != nil {
		return Session{}, err
	}

	if config.IsEmpty() {
		return Session{}, ErrNoConfigurationLoaded
	}

	conn, err := ocm.NewConnection().Build()
	if err != nil {
		return Session{}, err
	}

	return Session{
		config: config,
		conn:   conn,
		logger: log.WithFields(log.Fields{
			"ocm_url": config.URL(),
		}),
	}, nil
}

// Session provides access to the session-bound parameters
// for an invocation of this plug-in.
type Session struct {
	config Config
	conn   *sdk.Connection
	logger log.Interface
}

// Config returns the config loaded for the current session.
func (s *Session) Config() Config {
	return s.config
}

// Conn returns the OCM connection for the current session.
func (s *Session) Conn() *sdk.Connection {
	return s.conn
}

// Logger returns a log instance with session context added.
func (s *Session) Logger() log.Interface {
	return s.logger
}

// End releases any open session resources and logs any errors
// that may occur during this process.
func (s *Session) End() {
	if err := s.conn.Close(); err != nil {
		s.logger.
			WithError(err).
			Error("unable to release connection to OCM")
	}
}
