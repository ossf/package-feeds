# PyPI Feed

This feed allows polling of package updates from the PyPI package repository.

## Configuration options

The `packages` Field can be supplied to the PyPI feed options to enable polling of package specific apis.
This is less effective with large lists of packages as it polls the RSS feed for each package individually,
but it is much less likely to miss package updates between polling.


```
feeds:
- type: pypi
  options:
    packages:
    - numpy
    - scipy
```

# PyPI Artifacts Feed

This feed allows polling of PyPI package updates using the
[XML-RPC feed](https://warehouse.pypa.io/api-reference/xml-rpc.html#mirroring-support).
This feed contains extra information compared to the other PyPI feed in this project.
In particular, this avoids missing upstream notifications when platform-specific archives are
uploaded for a package some time after the release was made.

## Configuration

No configuration; all package updates are monitored
```
feeds:
- type: pypi-v2
```
