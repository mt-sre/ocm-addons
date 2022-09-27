// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

/* #nosec */

package integration

import (
	"strings"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
)

func MatchRow(values ...interface{}) types.GomegaMatcher {
	var subs []string

	for i := 0; i < len(values); i++ {
		subs = append(subs, `%s`)
	}

	row := strings.Join(subs, `\s*|\s*`)

	return gbytes.Say(row, values...)
}

var (
	dotJoin   = joinBy(".")
	slashJoin = joinBy("/")
)

func joinBy(sep string) func(...string) string {
	return func(words ...string) string {
		return strings.Join(words, sep)
	}
}
