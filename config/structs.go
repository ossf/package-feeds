package config

type ScheduledFeedConfig struct {
	PubConfig    PublisherConfig `yaml:"publisher"`
	EnabledFeeds []string        `yaml:"enabled_feeds"`
	HTTPPort     int             `yaml:"http_port,omitempty"`
	PollRate     string          `yaml:"poll_rate"`
	Timer        bool            `yaml:"timer"`
}

type PublisherConfig struct {
	Type   string      `mapstructure:"type"`
	Config interface{} `mapstructure:"config"`
}
