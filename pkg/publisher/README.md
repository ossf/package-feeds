# Publishers

Various publishers are available for use publishing packages, each of these can be configured for use as seen in examples below.

## Configuration examples

### stdout

```
publisher:
    type: stdout
```

### GCP Pub Sub

```
publisher:
    type: gcp_pubsub
    config:
        url: gcppubsub://foo.bar
```

### stdout

```
publisher:
    type: kafka
    config:
        brokers:
            - 127.0.0.1:9092
        topic: packagefeeds
```

### HTTP client

```
publisher:
    type: http-client
    config:
      url: "http://target-server:8000/package_feeds_hook"
```
