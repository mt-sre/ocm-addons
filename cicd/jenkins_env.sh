#!/bin/bash

# utilize local go 1.17 version if available
GO_1_17="/opt/go/1.17.7/bin"

if [ -d  "${GO_1_17}" ]; then
     PATH="${GO_1_17}:${PATH}"
fi

PYTHON_VERSION="$(python3 -c 'import sys; print(*sys.version_info[:3])')"
PYTHON_MAJOR_VERSION=$(echo "${PYTHON_VERSION}" | cut -d' ' -f1)
PYTHON_MINOR_VERSION=$(echo "${PYTHON_VERSION}" | cut -d' ' -f2)
PYTHON_PATCH_VERSION=$(echo "${PYTHON_VERSION}" | cut -d' ' -f3)
