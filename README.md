# Mattermost Nagios Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-nagios/master)](https://circleci.com/gh/mattermost/mattermost-plugin-nagios)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-nagios/master)](https://codecov.io/gh/mattermost/mattermost-plugin-nagios)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-nagios)](https://github.com/mattermost/mattermost-plugin-nagios/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-nagios/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-nagios/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Help Wanted Tickets**

Visit https://github.com/mattermost/mattermost-plugin-nagios/issues.

# Contents

- [Overview](#overview)
- [Features](#features)
- [Admin Guide](docs/admin-guide.md)
- [End User Guide](#end-user-guide)
- [Contribute](#contribute)
- [License](#license)
- [Security Vulnerability Disclosure](#security-vulnerability-disclosure)
- [Get Help](#get-help)

## Overview

A Mattermost Plugin for Nagios to get logs, alerts, and notifications in Mattermost. Supports Nagios Core >= 4.4.x.

**Maintainers**: [@amwolff](https://github.com/amwolff) & [@DanielSz50](https://github.com/DanielSz50)

## Features

This plugin allows you to:

- Get logs from specific systems without leaving Mattermost.
- Get alerts and notifications resembling the `showlog.cgi` UI instantly delivered.
- Receive system monitoring reports on a subscribed channel.
- Be frequently informed which hosts and services have an abnormal state.
- Receive notifications about changes to the configuration on a subscribed channel.
- Receive a diff between the old and the new version anytime a change has been made to Nagios configuration.

This README provides guidance on installation, configuration, and usage.

## [Admin Guide](docs/admin-guide.md)

## End User Guide

### Get Started

### Use the Plugin

### Slash Commands Overview

- `nagios`
    - `get-logs <alerts|notifications>`
        - `[host|service <host name|service description>]`
    - `set-logs-limit <count>`
    - `set-logs-start-time <seconds>`
    - `subscribe <report|configuration-changes>`
    - `unsubscribe <report|configuration-changes>`
    - `set-report-frequency <minutes>`

### Slash commands

#### nagios

`nagios`

This is the root command.

##### get-logs

`get-logs <alerts|notifications> [<host|service> <host name|service description>]`

This action allows you to get alerts or notifications.

Example: `/nagios get-logs alerts`

###### host

`get-logs <alerts|notifications> host <host name>`

This optional parameter allows you to get alerts or notifications from a specific host.

Example: `/nagios get-logs alerts host bvmhost-p09-02.iad2.fedoraproject.org`

###### service

`get-logs <alerts|notifications> service <service description>`

This optional parameter allows you to get alerts or notifications from a specific service.

Example: `/nagios get-logs alerts service Swap-Is-Low`

##### set-logs-limit

`set-logs-limit <count>`

This action allows you to limit the number of logs `get-logs` fetches.

Example: `/nagios set-logs-limit 10`

##### set-logs-start-time

`set-logs-start-time <seconds>`

This action allows you to specify the age of the oldest log `get-logs` fetches.

Example: `/nagios set-logs-start-time 3600`

##### subscribe

`subscribe <report|configuration-changes>`

This action allows you to subscribe to system monitoring reports or configuration changes on the current channel.

Example: `/nagios subscribe report`

##### unsubscribe

`unsubscribe <report|configuration-changes>`

This action allows you to unsubscribe from system monitoring reports or configuration changes on the current channel.

Example: `/nagios unsubscribe configuration-changes`

##### set-report-frequency

`set-report-frequency <minutes>`

This action allows you to set the frequency of system monitoring reports.

Example: `/nagios set-report-frequency 60`

### FAQ

## Contribute

### Development

This repository uses the [mattermost-plugin-starter-template](https://github.com/mattermost/mattermost-plugin-starter-template). Therefore, developing this plugin is roughly the same as it is with every plugin using the template. All the necessary steps to develop are in the template's repository.

If you are a Nagios admin/user and think there is something this plugin lacks or something that it does could be done another way, let us know! We are trying to develop this plugin based on users' needs. If there is a certain feature you or your team needs, open up an issue, and explain your needs. We will be happy to help.

This plugin only contains a server portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/integrate/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/integrate/plugins/developer-setup/) for more information about developing and extending plugins.

### Running a Nagios server with Docker

There is a [docker-compose.yml](https://github.com/mattermost/mattermost-plugin-naguis/blob/master/dev/docker-compose.yml) in the `dev` folder of the repository, configured to run a Nagios server for development. You can run `make nagios` in the root of the repository to spin up the Nagios server. The Nagios web application will be served at http://localhost:8080.

You can login with these credentials:

- Username: `nagiosadmin`
- Password: `nagios`

### Develop the watcher

To build the watcher, you can use the following command:

```sh
env GOOS=linux GOARCH=amd64 go build -o dist/watcherX.Y.Z.linux-amd64 -a -v cmd/watcher/main.go
```

Of course, you can build the watcher for other operating systems and architectures too.

### Deploy with Local Mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/administration/mmctl-cli-tool.html#local-mode) to streamline deploying your plugin. After configuring it, just run:

```sh
make deploy
```
### Deploy with credentials

Alternatively, you can authenticate with the server's API with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):

```sh
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

See the Development guide section for information about how to contribute to this plugin.

### I saw a bug, I have a feature request or a suggestion

Please file a [GitHub issue](https://github.com/mattermost/mattermost-plugin-nagios/issues), it will be very useful!

Pull Requests are welcome! You can contact us on the [Mattermost Community ~Plugin: Nagios channel](https://community.mattermost.com/core/channels/plugin-nagios).

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options.

### Share Feedback

Feel free to create a [GitHub Issue](https://github.com/mattermost/mattermost-plugin-nagios/issues) or join the [Nagios Plugin channel](https://community.mattermost.com/core/channels/plugin-nagios) on the Mattermost Community server to discuss.

## License

## Security vulnerability disclosure

Please report any security vulnerability to [https://mattermost.com/security-vulnerability-report/](https://mattermost.com/security-vulnerability-report/).

## Get Help
