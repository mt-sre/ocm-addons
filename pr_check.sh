#!/bin/bash

# SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

set -exvo pipefail -o nounset

source "${PWD}/cicd/jenkins_env.sh"

./mage -t 10m run-hooks && ./mage -t 10m test
