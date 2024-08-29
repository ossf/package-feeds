package stdout

import (
	"context"
	"fmt"
)

const (
	PublisherType = "stdout"
)

type Stdout struct{}

func New() *Stdout {
	return &Stdout{}
}

func (pub *Stdout) Name() string {
	return PublisherType
}

func (pub *Stdout) Send(_ context.Context, body []byte) error {
	fmt.Printf("%s\n", body)
	return nil
}
