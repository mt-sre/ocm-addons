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

Create a fork of this repository and clone that fork to your local machine.

All development tooling can be accessed through _mage_ and simply running
`./mage` will list all targets with additional information for each target
available by running `./mage -h [target]`.

Additionally you should install [pre-commit](https://pre-commit.com/#install) to
ensure code quality with every commit. Once installed run `pre-commit install`
in your local copy of this repository to initialize the pre-commit hooks.

## Configuration

### Adding New Customer Notifications

See [this](internal/notification/data/README.md) document for information
regarding the addition of new customer notifications.

## Known Issues

### Pager support

The `ocm-cli` offers the ability to configure a pager to send output
to by default. This can be done with the command `ocm config set pager [pager]`.
Currently only the `ocm addons installations` and `ocm addons cluster events`
subcommands support this feature.

## License

Apache License 2.0, see [LICENSE](LICENSE).
