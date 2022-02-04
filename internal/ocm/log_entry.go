package ocm

import (
	"strings"

	"github.com/mt-sre/ocm-addons/internal/output"
	slv1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
)

type LogEntry struct {
	*slv1.LogEntry
}

func (l *LogEntry) ToRow() output.Row {
	severity := strings.ToUpper(string(l.Severity()))

	return output.Row{
		{Name: "timestamp", Value: l.Timestamp()},
		{Name: "cluster_uuid", Value: l.ClusterUUID()},
		{Name: "description", Value: l.Description()},
		{Name: "id", Value: l.ID()},
		{Name: "service_name", Value: l.ServiceName()},
		{Name: "severity", Value: severity},
		{Name: "summary", Value: l.Summary()},
		{Name: "username", Value: l.Username()},
	}
}

type LogLevel string

const (
	LogLevelNone    = ""
	LogLevelDebug   = "Debug"
	LogLevelInfo    = "Info"
	LogLevelWarning = "Warning"
	LogLevelError   = "Error"
	LogLevelFatal   = "Fatal"
)

func NewLogEntrySorter(size int, sortFunc LogEntrySortFunc) *LogEntrySorter {
	return &LogEntrySorter{
		entries:  make([]LogEntry, 0, size),
		sortFunc: sortFunc,
	}
}

type LogEntrySorter struct {
	entries  []LogEntry
	sortFunc LogEntrySortFunc
}

func (s *LogEntrySorter) Len() int           { return len(s.entries) }
func (s *LogEntrySorter) Swap(i, j int)      { s.entries[i], s.entries[j] = s.entries[j], s.entries[i] }
func (s *LogEntrySorter) Less(i, j int) bool { return s.sortFunc(s.entries[i], s.entries[j]) }

func (s *LogEntrySorter) Append(e LogEntry) {
	s.entries = append(s.entries, e)
}

func (s *LogEntrySorter) Entries() []LogEntry {
	result := make([]LogEntry, len(s.entries))

	copy(result, s.entries)

	return result
}

type LogEntrySortFunc func(LogEntry, LogEntry) bool

func LogEntryByTime(ord Order) LogEntrySortFunc {
	return func(e1, e2 LogEntry) bool {
		if ord == OrderAsc {
			return e1.Timestamp().Before(e2.Timestamp())
		}

		return e1.Timestamp().After(e2.Timestamp())
	}
}

type Order string

const (
	OrderNone = ""
	OrderAsc  = "ascending"
	OrderDesc = "descending"
)
