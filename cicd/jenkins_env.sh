#!/bin/bash

# SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

# utilize local go 1.18 version if available
GO_1_19="/opt/go/1.19.3/bin"

if [ -d  "${GO_1_19}" ]; then
     PATH="${GO_1_19}:${PATH}"
fi
