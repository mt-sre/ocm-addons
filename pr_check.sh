#!/bin/bash

set -exvo pipefail -o nounset

# utilize local go 1.17 version if available
GO_1_17="/opt/go/1.17.5/bin"

if [ -d  "${GO_1_17}" ]; then
     PATH="${GO_1_17}:${PATH}"
fi

./mage -t 10m check
./mage -t 10m test
