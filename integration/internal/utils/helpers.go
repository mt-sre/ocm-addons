/* #nosec */

package utils

import (
	"strings"
)

var (
	DotJoin   = JoinBy(".")
	SlashJoin = JoinBy("/")
)

func JoinBy(sep string) func(...string) string {
	return func(words ...string) string {
		return strings.Join(words, sep)
	}
}
