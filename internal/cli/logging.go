package cli

import "github.com/apex/log"

const (
	LogLevelError = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// LogLevel converts an integer value to the appropriate log Level.
func LogLevel(verbosity int) log.Level {
	switch verbosity {
	case LogLevelError:
		return log.ErrorLevel
	case LogLevelWarn:
		return log.WarnLevel
	case LogLevelInfo:
		return log.InfoLevel
	default:
		return log.DebugLevel
	}
}
