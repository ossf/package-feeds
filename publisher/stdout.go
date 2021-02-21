package publisher

import (
	"context"
	"fmt"
)

type Stdout struct{}

func NewStdoutPublisher() *Stdout {
	return &Stdout{}
}

func (pub *Stdout) Name() string {
	return "stdout"
}

func (pub *Stdout) Send(ctx context.Context, body []byte) error {
	fmt.Printf("%s\n", body)
	return nil
}
