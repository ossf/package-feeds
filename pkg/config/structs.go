package config

import (
	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/publisher/httpclientpubsub"
)

type ScheduledFeedConfig struct {
	// Configures the publisher for pushing packages after polling.
	PubConfig PublisherConfig `yaml:"publisher"`

	// Configures the feeds to be used for polling from package repositories.
	Feeds []FeedConfig `yaml:"feeds"`

	HTTPPort int    `yaml:"http_port,omitempty"`
	PollRate string `yaml:"poll_rate"`
	Timer    bool   `yaml:"timer"`

	// Configures the EventHandler instance to be used throughout the package-feeds application.
	EventsConfig *EventsConfig `yaml:"events"`

	eventHandler *events.Handler
}

type PublisherConfig struct {
	Type             string                  `mapstructure:"type"`
	Config           interface{}             `mapstructure:"config"`
	HTTPClientConfig httpclientpubsub.Config `mapstructure:"http_client_config"`
}

type FeedConfig struct {
	Type    string            `mapstructure:"type"`
	Options feeds.FeedOptions `mapstructure:"options"`
}

type EventsConfig struct {
	Sink        string        `yaml:"sink"`
	EventFilter events.Filter `yaml:"filter"`
}
