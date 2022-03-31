#!/bin/bash

set -exvo pipefail -o nounset

source "${PWD}/cicd/jenkins_env.sh"

SKIP=""

if ! [[ (${PYTHON_MAJOR_VERSION} -eq 3 && ${PYTHON_MINOR_VERSION} -ge 8) ]]; then
     SKIP="pymarkdown"
fi

SKIP="${SKIP}" ./mage -t 10m run-hooks && ./mage -t 10m test
