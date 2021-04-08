package config

type ScheduledFeedConfig struct {
	PubConfig    TypeConfigPair   `yaml:"publisher"`
	EnabledFeeds []string         `yaml:"enabled_feeds"`
	Feeds        []TypeConfigPair `yaml:"feeds"`
	HttpPort     int              `yaml:"http_port,omitempty"`
}

type TypeConfigPair struct {
	Type   string      `mapstructure:"type"`
	Config interface{} `mapstructure:"config"`
}
