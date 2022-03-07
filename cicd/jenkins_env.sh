#!/bin/bash

# utilize local go 1.17 version if available
GO_1_17="/opt/go/1.17.7/bin"

if [ -d  "${GO_1_17}" ]; then
     PATH="${GO_1_17}:${PATH}"
fi
