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

# Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss this project in the [OSSF Securing Critical Projects Working Group](https://github.com/ossf/wg-securing-critical-projects) meetings.

See the [Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for the schedule and meeting invitations.
