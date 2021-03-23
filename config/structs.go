package config

type ScheduledFeedConfig struct {
	PubConfig    PublisherConfig `yaml:"publisher"`
	EnabledFeeds []string        `yaml:"enabled_feeds"`
	HttpPort     int             `yaml:"http_port,omitempty"`
}

type PublisherConfig struct {
	Type   string      `mapstructure:"type"`
	Config interface{} `mapstructure:"config"`
}
