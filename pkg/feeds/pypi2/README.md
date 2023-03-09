# Pypi Feed

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