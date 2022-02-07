#!/bin/bash

set -exvo pipefail -o nounset

IMAGE=ocm-addons-ci

docker build -t ${IMAGE} -f Dockerfile.ci .
docker run --rm -e "GITHUB_TOKEN=${GITHUB_TOKEN}" "${IMAGE}" release:full
