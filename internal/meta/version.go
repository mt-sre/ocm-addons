package meta

import "fmt"

var (
	version = ""
	commit  = "n/a"
	date    = "n/a"
	builtBy = "n/a"
)

func Version() string {
	if version == "" {
		return "dev"
	}

	return "v" + version
}

func LongVersion() string {
	ver := "v" + version

	if version == "" {
		ver = "dev"
	}

	return fmt.Sprintf("version: %s\ncommit: %s\ndate: %s\nbuilt by: %s", ver, commit, date, builtBy)
}
