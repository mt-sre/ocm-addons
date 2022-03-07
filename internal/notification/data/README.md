# Customer Notifications

## Overview

Issues related to OSD addons may require that a notification be sent to the
cluster owner in the event that a SRE team supporting the addon is required
to perform some manual action.

This directory hosts configurations for such notifications which may be sent
using the `ocm addons notify` command.

## Adding New Teams

Each team is required to create their own subdirectory with the same name
as the owning team. For example, the team directory for the _MTSRE_ team
is named `mtsre`.

## Adding New Notification Configurations

Within each team directory are any number of `yaml` files each containing
notifications for a particular addon. The name of the file, excluding the
extension, will be used as the product name.

The general structure of the config is a top-level dictionary where the keys
are the id of each configuration.

Example:

```yaml
---
important-notification:
  summary: CriticalServiceFailure
  description: The service has critically failed. Service will be reinstalled
  severity: Error
```

The full list of configurable fields are as follows:

|Field       |Description                                                                          |Default          |
|------------|-------------------------------------------------------------------------------------|-----------------|
|description |A complete description of the alert which will also be sent to the customer via email|N/A              |
|internalOnly|Whether the log entry will be visible to customers                                   |false            |
|serviceName |The service which created the log entry                                              |"SREManualAction"|
|severity    |The severity level (debug, error, fatal, info, warning)                              |N/A              |
|summary     |Brief description of the alert                                                       |N/A              |

## FAQ

### Descriptions

The description field is sent to the cluster owner in the body of a notification email.
When formatted as an email a trailing period is automatically added so the configured
description should not end with any punctuation.
