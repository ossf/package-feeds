package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ossf/package-feeds/feeds"
	"github.com/ossf/package-feeds/feeds/crates"
	"github.com/ossf/package-feeds/feeds/goproxy"
	"github.com/ossf/package-feeds/feeds/npm"
	"github.com/ossf/package-feeds/feeds/nuget"
	"github.com/ossf/package-feeds/feeds/packagist"
	"github.com/ossf/package-feeds/feeds/pypi"
	"github.com/ossf/package-feeds/feeds/pypi_critical"
	"github.com/ossf/package-feeds/feeds/rubygems"
	"github.com/ossf/package-feeds/publisher"
	"github.com/ossf/package-feeds/publisher/gcppubsub"
	"github.com/ossf/package-feeds/publisher/stdout"
	"gopkg.in/yaml.v3"

	"github.com/mitchellh/mapstructure"
)

// Loads a ScheduledFeedConfig struct from a yaml config file
func FromFile(configPath string) (*ScheduledFeedConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	return NewConfigFromBytes(data)
}

// Loads a ScheduledFeedConfig struct from a yaml bytes
func NewConfigFromBytes(bytes []byte) (*ScheduledFeedConfig, error) {
	config := Default()

	err := unmarshalStrict(bytes, config)
	if err != nil {
		return nil, err
	}
	config.applyEnvVars()

	return config, nil
}

// Applies environment variables to the configuration
func (config *ScheduledFeedConfig) applyEnvVars() {
	// Support legacy env var definition for gcp pub sub.
	pubURL := os.Getenv("OSSMALWARE_TOPIC_URL")
	if pubURL != "" {
		config.PubConfig = TypeConfigPair{
			Type: gcppubsub.PublisherType,
			Config: map[string]interface{}{
				"url": pubURL,
			},
		}
	}

	portStr, portProvided := os.LookupEnv("PORT")
	port, err := strconv.Atoi(portStr)

	if portProvided && err == nil {
		config.HttpPort = port
	}
}

func AddTo(ls *[]int, value int) {
	*ls = append(*ls, value)
}

// Constructs a map of ScheduledFeeds to enable based on the EnabledFeeds provided from configuration, indexed by the feed type.
func (sConfig *ScheduledFeedConfig) GetScheduledFeeds() (map[string]feeds.ScheduledFeed, error) {
	var err error
	scheduledFeeds := map[string]feeds.ScheduledFeed{}

	// Parse feeds with custom configurations
	for _, entry := range sConfig.Feeds {
		feed, err := entry.ToFeed()
		if err != nil {
			return nil, fmt.Errorf("failed to parse feeds entries: %w", err)
		}
		scheduledFeeds[entry.Type] = feed
		fmt.Printf("%#v", feed)
	}

	for _, entry := range sConfig.EnabledFeeds {
		// If these are already added by sConfig.Feeds, we can skip them
		if _, ok := scheduledFeeds[entry]; ok {
			continue
		}
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
			err = fmt.Errorf("unknown feed type %v", entry)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse enabled_feeds entries: %w", err)
	}

	return scheduledFeeds, nil
}

// Produces a Publisher object from the provided TypeConfigPair
// The TypeConfigPair.Type value is evaluated and the appropriate Publisher is
// constructed from the Config field. If the type is not a recognised Publisher type,
// an error is returned.
func (pc TypeConfigPair) ToPublisher(ctx context.Context) (publisher.Publisher, error) {
	var err error
	switch pc.Type {
	case gcppubsub.PublisherType:
		var config gcppubsub.PubSubConfig
		err = strictDecode(pc.Config, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to decode gcppubsub config: %w", err)
		}
		return gcppubsub.FromConfig(ctx, config)
	case stdout.PublisherType:
		return stdout.New(), nil
	default:
		return nil, err
	}
}

// Produces a Feed object from the provided TypeConfigPair
// The TypeConfigPair.Type value is evaluated and the appropriate Feed is
// constructed from the Config field. If the type is not a recognised Feed type,
// an error is returned.
func (fc TypeConfigPair) ToFeed() (feeds.ScheduledFeed, error) {
	var err error
	feedMap := map[string]feeds.ScheduledFeed{
		crates.FeedName:        (*crates.Feed)(nil),
		goproxy.FeedName:       (*goproxy.Feed)(nil),
		npm.FeedName:           (*npm.Feed)(nil),
		nuget.FeedName:         (*nuget.Feed)(nil),
		pypi.FeedName:          (*pypi.Feed)(nil),
		pypi_critical.FeedName: (*pypi_critical.Feed)(nil),
		packagist.FeedName:     (*packagist.Feed)(nil),
		rubygems.FeedName:      (*rubygems.Feed)(nil),
	}

	feed, ok := feedMap[fc.Type]
	if !ok {
		return nil, fmt.Errorf("invalid type provided for feed configuration: %v", fc.Type)
	}
	err = strictDecode(fc.Config, &feed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %v feed config: %w", fc.Type, err)
	}

	return feed, nil
}

// Decode an input using mapstruct decoder with strictness enabled, errors will be returned in
// the case of unused fields.
func strictDecode(input interface{}, out interface{}) error {
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
		PubConfig: TypeConfigPair{
			Type: stdout.PublisherType,
		},
		HttpPort: 8080,
	}
	config.applyEnvVars()
	return config
}

// Unmarshals configuration data from bytes into the provided interface, strictness is
// enabled which returns an error in the case that an unknown field is provided.
func unmarshalStrict(data []byte, out interface{}) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil && err != io.EOF {
		return err
	}
	return nil
}
