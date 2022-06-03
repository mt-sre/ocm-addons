#!/bin/bash

set -exvo pipefail -o nounset

source "${PWD}/cicd/jenkins_env.sh"

./mage -t 10m run-hooks && ./mage -t 10m test
