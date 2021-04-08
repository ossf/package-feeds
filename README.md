# Package Feeds

The binary produced by [cmd/scheduled-feed/main.go](cmd/scheduled-feed/main.go) can be used to monitor various
package repositories for changes and publish data to external services for further processing.

Additionally, the repo contains a few subprojects to aid in the analysis of these open source packages, in particular to look for malicious software.

These are:

[Feeds](./feeds/) to watch package registries (PyPI, NPM, etc.) for changes to packages
and to make that data available via a single standard interface.

[Publisher](./publisher/) provides the functionality to push package details from feeds towards
external services such as GCP Pub/Sub.

This repo used to contain several other projects, which have since been split out into
[github.com/ossf/package-analysis](https://github.com/ossf/package-analysis).

The goal is for all of these components to work together and provide extensible, community-run
infrastructure to study behavior of open source packages and to look for malicious software.
We also hope that the components can be used independently, to provide package feeds or runtime
behavior data for anyone interested.

# Configuration

A YAML configuration file can be provided with the following format:

```
enabled_feeds:
- pypi
- npm
- goproxy
- rubygems
- crates

publisher:
  type: 'gcp_pubsub'
  config:
    url: "gcppubsub://foobar.com"

http_port: 8080

cutoff_delta: 5m

timer: false
```

`cutoff_delta` string formatted for [duration parser](https://golang.org/pkg/time/#ParseDuration).
`timer` will configure interal polling of the `enabled_feeds` at the given `cutoff_delta` period. To specify this configuration file, define its path in your environment under the `PACKAGE_FEEDS_CONFIG_PATH` variable.

## Legacy Configuration

Legacy configuration methods are still supported. By default, without a configuration file all feeds will be enabled. The environment variable `OSSMALWARE_TOPIC_URL` can be used to select the GCP pubsub publisher and `PORT` will configure the port for the HTTP server.
The default `cutoff_delta` is 5 minutes, this is used to generate a cutoff point for feed events relative to the given time at execution.

# Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss this project in the [OSSF Securing Critical Projects Working Group](https://github.com/ossf/wg-securing-critical-projects) meetings.

See the [Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for the schedule and meeting invitations.
