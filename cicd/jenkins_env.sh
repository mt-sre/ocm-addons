#!/bin/bash

# SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

# utilize local go 1.21 version if available
GO_1_21="/opt/go/1.21.3/bin"

if [ -d  "${GO_1_21}" ]; then
     PATH="${GO_1_21}:${PATH}"
fi
