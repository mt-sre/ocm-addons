#!/bin/bash

# SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

function main() {
     curl -L "$(download_url)" | tar xvz ocm-addons
}

function download_url() {
     tag=$(latest_release)
     kernel=$(system_kernel)
     arch=$(system_arch)

     echo "https://github.com/mt-sre/ocm-addons/releases/download/v${tag}/ocm-addons_${tag}_${kernel}_${arch}.tar.gz"
}

function latest_release() {
     curl -s "https://api.github.com/repos/mt-sre/ocm-addons/releases/latest" \
          | jq .tag_name \
          | grep -o '[[:digit:]]\+\.[[:digit:]]\+\.[[:digit:]]\+'
}

function system_kernel() {
     case "$(uname -s)" in
          "Linux")
          echo "linux"
          ;;
          "Darwin")
          echo "darwin"
          ;;
          *)
          echo ""
          ;;
     esac
}

function system_arch() {
     arch="$(uname -m)"

     case "${arch}" in
          "x86_64")
          echo "amd64"
          ;;
          .*386.*)
          echo "386"
          ;;
          "armv8")
          echo "arm64"
          ;;
          *)
          "${arch}"
          ;;
     esac
}

main
