#!/bin/bash

# SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

# utilize local go 1.20 version if available
GO_1_20="/opt/go/1.20.6/bin"

if [ -d  "${GO_1_20}" ]; then
     PATH="${GO_1_20}:${PATH}"
fi
