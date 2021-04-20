package events

import (
	"fmt"
)

type LossyFeedEvent struct {
	Feed string
}

func (e LossyFeedEvent) GetComponent() string {
	return FeedsComponentType
}

func (e LossyFeedEvent) GetType() string {
	return LossyFeedEventType
}

func (e LossyFeedEvent) GetMessage() string {
	return fmt.Sprintf("detected potential missing package data when polling %v feed", e.Feed)
}
