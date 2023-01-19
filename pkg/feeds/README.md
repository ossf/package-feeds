# Feeds

Each of the feeds have their own implementation and support their own set of configuration options.

## Configuration options

`packages` this configuration option is only available on certain feeds, check the README of the feed you're interested in for information on this.

`poll_rate` this allows for setting the frequency of polling for this specific feed. This is supported by all feeds. The value should be a string formatted for [duration parser](https://golang.org/pkg/time/#ParseDuration). Setting this value will enable the scheduled polling regardless of the value of `timer` in the root of the configuration.

## Example

### Poll Pypi every 5 minutes

```
feeds:
- type: pypi
  options:
    poll_rate: "5m"
```

### Poll npm every 10 minutes and crates every hour

```
feeds:
- type: npm
  options:
    poll_rate: "10m"
- type: crates
  options:
    poll_rate: "1h"
```

### Poll a subset of pypi every 10 minutes

```
feeds:
- type: pypi
  options:
    packages:
    - numpy
    - django
    poll_rate: "10m"
```
