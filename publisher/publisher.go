package publisher

import (
	"context"
)

type Publisher interface {
	Send(ctx context.Context, body []byte) error
	Name() string
}
