# SPDX-FileCopyrightText: 2025 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

{
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = github:NixOS/nixpkgs/nixos-unstable;
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
    let
        version = "0.7.18";
        pkgs = import nixpkgs { inherit system; };
        ocm-cli = pkgs.buildGoModule rec {
          pname = "ocm-cli";
          version = "1.0.3";

          src = pkgs.fetchFromGitHub {
            owner = "openshift-online";
            repo = "ocm-cli";
            rev = "v${version}";
            hash = "sha256-RuGUIG58cyyWvHD/0T7xwtzFy9XJUmavkQg4MRAHQqQ=";
          };
          subPackages = [
            "cmd/ocm"
          ];
          vendorHash = "sha256-qkTh+tkU6MXBJkX0XwktRCMjoySe1/9uWHFGTc7ozRM=";
        };
        lichen = pkgs.buildGoModule rec {
          pname = "lichen";
          version = "0.1.7";

          src = pkgs.fetchFromGitHub {
            owner = "uw-labs";
            repo = "lichen";
            rev = "v${version}";
            hash = "sha256-6ZLfYFjbrzDYqiaRlRQnbCjrAzSRyCB2TqpMpN8p9mo=";
          };
          vendorHash = "sha256-UHqJlzimBVffU7eZ3bvV6ZJ+40svajyR29Eq0/ClEyI=";
        };
        devDeps = with pkgs; [
          git
          go_1_23
          goreleaser
          pre-commit
          golangci-lint
          ginkgo
          ocm-cli
          lichen
          reuse
        ];
    in
    {
      devShells.default = pkgs.mkShell {
        buildInputs = devDeps;
      };
  });
}
