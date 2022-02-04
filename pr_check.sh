#!/bin/bash

set -exvo pipefail -o nounset

IMAGE=ocm-addons-ci

docker build -t ${IMAGE} -f Dockerfile.ci .
docker run --rm --timeout 900 ${IMAGE} check test
