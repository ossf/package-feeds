package main

import (
	"context"

	"github.com/ossf/package-feeds/feeds/pypi"
)

func main() {
	pypi.Poll(context.Background(), pypi.PubSubMessage{})
}
