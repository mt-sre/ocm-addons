<!--
SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>

SPDX-License-Identifier: Apache-2.0
-->

# Contributing

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

## Before Opening a Pull Request

Thank you for considering making a contribution to the `ocm-addons` plug-in.
Before opening a pull request please check that the issue/feature request
you are addressing is not already being worked on in any of the currently
open [pull requests](https://github.com/mt-sre/ocm-addons/pulls). If it is
then please feel free to contribute by testing and reviewing the changes.

Please also check for any open [issues](https://github.com/mt-sre/ocm-addons/issues)
that your PR may close and be sure to link those issues in your PR description.

### Testing Pull Requests Locally

To test PR's locally you must first clone this repository:

`git clone git@github.com:mt-sre/ocm-addons.git`

Then execute the following with the correct `<PULL REQUEST NUMBER>`:

```bash
PULL_REQUEST=<PULL REQUEST NUMBER>
git fetch $(curl -s https://api.github.com/repos/mt-sre/ocm-addons/pulls/${PULL_REQUEST} \
     | jq -r '(.head.repo.ssh_url) + " " + (.head.ref) + ":" + (.head.ref)')
```

## Development

### Installing `pre-commit` hooks

Run `nix develop --command pre-commit install` to install `pre-commit` and any configured _git_ hooks.
This will ensure that code quality is checked after every commit and also that
tests are run before each push.

### Running tests

To run tests on an ad-hoc basis run `nix develop --command make test` and when adding new packages
run `nix develop --command make lint` to tidy modules and verfiy that they are compliant with
this project's license.

### Building the Plug-in

Run `nix develop --command make` to build the plug-in. Binaries will be located under the
`dist` folder.

## Submitting Pull Requests

First fork this repository and create a branch off of `main` to commit to.
Commits should be named as per
[convential commits](https://www.conventionalcommits.org/en/v1.0.0/).
When submitting your PR fill in the pull request template to the best of
your abilities. If you are not a member of the _mt-sre_ organization a
member will have to leave a comment for CI to run checks against your
PR so please be patient until a member can review.

## Releasing

> Note: releases can only be performed by members of the _mt-sre_ organization.

Pull the latest changes from this repository to your local `main` branch.

```bash
git checkout main
git pull upstream main
```

Determine the next release version as per [semver](https://semver.org).

Then, tag the latest commit with the next release version:

```bash
git tag -a vX.Y.Z -m "release vX.Y.Z"
```

Finally push the tags upstream:

```bash
git push upstream --follow-tags
```
