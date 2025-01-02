#!/bin/bash

# SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

# utilize local go 1.23 version if available
GO_1_23="/opt/go/1.23.1/bin"

if [ -d  "${GO_1_23}" ]; then
     PATH="${GO_1_23}:${PATH}"
fi
