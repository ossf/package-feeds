package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/feeds/crates"
	"github.com/ossf/package-feeds/feeds/goproxy"
	"github.com/ossf/package-feeds/feeds/npm"
	"github.com/ossf/package-feeds/feeds/nuget"
	"github.com/ossf/package-feeds/feeds/packagist"
	"github.com/ossf/package-feeds/feeds/pypi"
	"github.com/ossf/package-feeds/feeds/rubygems"
	"github.com/ossf/package-feeds/publisher"
	"github.com/ossf/package-feeds/publisher/gcppubsub"
	"github.com/ossf/package-feeds/publisher/kafkapubsub"
	"github.com/ossf/package-feeds/publisher/stdout"
)

var (
	errUnknownFeed = errors.New("unknown feed type")
	errUnknownPub  = errors.New("unknown publisher type")
)

// Loads a ScheduledFeedConfig struct from a yaml config file.
func FromFile(configPath string) (*ScheduledFeedConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	return NewConfigFromBytes(data)
}

// Loads a ScheduledFeedConfig struct from a yaml bytes.
func NewConfigFromBytes(yamlBytes []byte) (*ScheduledFeedConfig, error) {
	config := Default()

	err := unmarshalStrict(yamlBytes, config)
	if err != nil {
		return nil, err
	}
	config.applyEnvVars()

	return config, nil
}

// Applies environment variables to the configuration.
func (config *ScheduledFeedConfig) applyEnvVars() {
	// Support legacy env var definition for gcp pub sub.
	pubURL := os.Getenv("OSSMALWARE_TOPIC_URL")
	if pubURL != "" {
		config.PubConfig = PublisherConfig{
			Type: gcppubsub.PublisherType,
			Config: map[string]interface{}{
				"url": pubURL,
			},
		}
	}

	portStr, portProvided := os.LookupEnv("PORT")
	port, err := strconv.Atoi(portStr)

	if portProvided && err == nil {
		config.HTTPPort = port
	}
}

func AddTo(ls *[]int, value int) {
	*ls = append(*ls, value)
}

// Constructs a map of ScheduledFeeds to enable based on the EnabledFeeds provided
// from configuration, indexed by the feed type.
func (config *ScheduledFeedConfig) GetScheduledFeeds() (map[string]feeds.ScheduledFeed, error) {
	var err error
	scheduledFeeds := map[string]feeds.ScheduledFeed{}
	for _, entry := range config.EnabledFeeds {
		switch entry {
		case crates.FeedName:
			scheduledFeeds[entry] = crates.Feed{}
		case goproxy.FeedName:
			scheduledFeeds[entry] = goproxy.Feed{}
		case npm.FeedName:
			scheduledFeeds[entry] = npm.Feed{}
		case nuget.FeedName:
			scheduledFeeds[entry] = nuget.Feed{}
		case pypi.FeedName:
			scheduledFeeds[entry] = pypi.Feed{}
		case packagist.FeedName:
			scheduledFeeds[entry] = packagist.Feed{}
		case rubygems.FeedName:
			scheduledFeeds[entry] = rubygems.Feed{}
		default:
			err = fmt.Errorf("%w : %v", errUnknownFeed, entry)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse enabled_feeds entries: %w", err)
	}
	return scheduledFeeds, nil
}

// Produces a Publisher object from the provided PublisherConfig
// The PublisherConfig.Type value is evaluated and the appropriate Publisher is
// constructed from the Config field.
func (pc PublisherConfig) ToPublisher(ctx context.Context) (publisher.Publisher, error) {
	var err error
	switch pc.Type {
	case gcppubsub.PublisherType:
		var gcpConfig gcppubsub.Config
		err = strictDecode(pc.Config, &gcpConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode gcppubsub config: %w", err)
		}
		return gcppubsub.FromConfig(ctx, gcpConfig)
	case kafkapubsub.PublisherType:
		var kafkaConfig kafkapubsub.Config
		err = strictDecode(pc.Config, &kafkaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode kafkapubsub config: %w", err)
		}
		return kafkapubsub.FromConfig(ctx, kafkaConfig)
	case stdout.PublisherType:
		return stdout.New(), nil
	default:
		err = fmt.Errorf("%w : %v", errUnknownPub, pc.Type)
	}
	return nil, err
}

// Decode an input using mapstruct decoder with strictness enabled, errors will be returned in
// the case of unused fields.
func strictDecode(input, out interface{}) error {
	strictDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      out,
	})
	if err != nil {
		return err
	}
	return strictDecoder.Decode(input)
}

func Default() *ScheduledFeedConfig {
	config := &ScheduledFeedConfig{
		EnabledFeeds: []string{
			crates.FeedName,
			goproxy.FeedName,
			npm.FeedName,
			nuget.FeedName,
			packagist.FeedName,
			pypi.FeedName,
			rubygems.FeedName,
		},
		PubConfig: PublisherConfig{
			Type: stdout.PublisherType,
		},
		HTTPPort:    8080,
		CutoffDelta: "5m",
		Timer:       false,
	}
	config.applyEnvVars()
	return config
}

// Unmarshals configuration data from bytes into the provided interface, strictness is
// enabled which returns an error in the case that an unknown field is provided.
func unmarshalStrict(data []byte, out interface{}) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
