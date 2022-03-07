#!/bin/bash

set -exvo pipefail -o nounset

source "${PWD}/cicd/jenkins_env.sh"

./mage release
