// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"strings"
)

func PromptYesOrNo(out io.Writer, in io.Reader, msg string) bool {
	fmt.Fprintf(out, "%s (y/n): ", msg)

	for {
		var answer string

		fmt.Fscanln(in, &answer)

		switch strings.ToLower(answer) {
		case "n", "no":
			return false
		case "y", "yes":
			return true
		default:
			fmt.Fprintf(out, "%s (y/n): ", msg)
		}
	}
}
