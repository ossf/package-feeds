# npm Feed

This feed allows polling of package updates from the repository.npmjs.org package repository.

## Configuration options

The `packages` Field can be supplied to the npm feed options to enable polling of package specific apis. This is much slower
with large lists of packages, but it is much less likely to miss package updates between polling.

```
feeds:
- type: npm
  options:
    packages:
    - lodash
    - react
```