#!/bin/bash

set -exvo pipefail -o nounset

docker login quay.io -u ${QUAY_USER} -p ${QUAY_TOKEN}

IMAGE=ocm-addons-ci

docker build -t ${IMAGE} -f Dockerfile.ci .
docker run --rm ${IMAGE} check test
