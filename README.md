<!--
SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>

SPDX-License-Identifier: Apache-2.0
-->

# OCM-AddOns plug-in

## Overview

OCM-CLI plugin to be used when working with add-ons.

## Installation

### Option 1: Install using `go install`

Execute the following `go install` command to build and install the `ocm-addons`
command at `$GOPATH/bin`.

```bash
go install github.com/mt-sre/ocm-addons/cmd/ocm-addons@latest
```

See these [instructions](https://go.dev/doc/install) for installing _go_ on your machine.

### Option 2: Build and install from source

Clone this repository to your local machine.

```bash
git clone git@github.com/mt-sre/ocm-addons.git
```

Run `./mage install` under the repository root to build the plug-in binary and
add it to your `$GOPATH/bin`.

If you have not added `$GOPATH/bin` to your `$PATH`, then you may
alternatively run `./mage build` and manually move the `bin/ocm-addons`
binary to any other directory in your `$PATH`.

### Option 3: Install using 'install.sh' (Not supported for Windows)

```bash
curl -L https://raw.githubusercontent.com/mt-sre/ocm-addons/main/scripts/install.sh | bash
```

## Usage

Install the [ocm-cli](https://github.com/openshift-online/ocm-cli#installation)
and ensure that you are able to log in to an active _OCM_ environment.

Run `ocm login ...` to establish a login session with your chosen _OCM_ environment.

Once logged in the plug-in can be accessed by running `ocm addons` which will
display command information.

_Note: the plug-in requires that an `ocm.json` file exist which the OCM-CLI
will generate upon login and will remove upon logout. In the current state
`$OCM_TOKEN` will be ignored if present in the environment._

## Development

See the [contributing](CONTRIBUTING.md) guide for more information.

## Configuration

### Adding New Customer Notifications

See [this](internal/notification/data/README.md) document for information
regarding the addition of new customer notifications.

## Known Issues

No issues have been reported at this time.

## License

Apache License 2.0, see [LICENSES/Apache-2.0.txt](LICENSES/Apache-2.0.txt).

This repository complies with the REUSE [specification](https://reuse.software/spec/)
and any new files added will be checked for compliance. This means that an appropriate
SPDX header must be present for new files which can be added manually or by using
the REUSE [tool](https://git.fsfe.org/reuse/tool).
