# Github Feed

This feed allows polling of releases from specific github repositories. This feed **requires** a list of repositories to be specified.

## Configuration options

The `packages` field is required by the github feed.

```
feeds:
- type: github
  options:
    packages:
    - "ossf/package-feeds"
```