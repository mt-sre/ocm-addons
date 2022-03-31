#!/bin/bash

set -exvo pipefail -o nounset

source "${PWD}/cicd/jenkins_env.sh"

SKIP=""

if [ "${HAS_PYTHON_38}" != "True" ]; then
     SKIP="pymarkdown"
fi

SKIP="${SKIP}" ./mage -t 10m run-hooks && ./mage -t 10m test
