package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/ossf/package-feeds/pkg/events"
	"github.com/ossf/package-feeds/pkg/feeds"
	"github.com/ossf/package-feeds/pkg/feeds/crates"
	"github.com/ossf/package-feeds/pkg/feeds/goproxy"
	"github.com/ossf/package-feeds/pkg/feeds/npm"
	"github.com/ossf/package-feeds/pkg/feeds/nuget"
	"github.com/ossf/package-feeds/pkg/feeds/packagist"
	"github.com/ossf/package-feeds/pkg/feeds/pypi"
	"github.com/ossf/package-feeds/pkg/feeds/rubygems"
	"github.com/ossf/package-feeds/pkg/publisher"
	"github.com/ossf/package-feeds/pkg/publisher/gcppubsub"
	"github.com/ossf/package-feeds/pkg/publisher/httpclientpubsub"
	"github.com/ossf/package-feeds/pkg/publisher/kafkapubsub"
	"github.com/ossf/package-feeds/pkg/publisher/stdout"
)

var (
	errUnknownFeed     = errors.New("unknown feed type")
	errUnknownPub      = errors.New("unknown publisher type")
	errUnknownSinkType = errors.New("unknown sink type")

	// feed-specific poll rate is left unspecified, so it can still be
	// configured by the global 'poll_rate' option in the ScheduledFeedConfig YAML.
	defaultFeedOptions    = feeds.FeedOptions{Packages: nil, PollRate: ""}
	npmDefaultFeedOptions = feeds.FeedOptions{Packages: nil, PollRate: "2m"}
)

// Loads a ScheduledFeedConfig struct from a yaml config file.
func FromFile(configPath string) (*ScheduledFeedConfig, error) {
	data, err := os.ReadFile(configPath)
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
func (sc *ScheduledFeedConfig) applyEnvVars() {
	// Support legacy env var definition for gcp pub sub.
	pubURL := os.Getenv("OSSMALWARE_TOPIC_URL")
	if pubURL != "" {
		sc.PubConfig = PublisherConfig{
			Type: gcppubsub.PublisherType,
			Config: map[string]interface{}{
				"url": pubURL,
			},
		}
	}

	portStr, portProvided := os.LookupEnv("PORT")
	port, err := strconv.Atoi(portStr)

	if portProvided && err == nil {
		sc.HTTPPort = port
	}
}

func AddTo(ls *[]int, value int) {
	*ls = append(*ls, value)
}

// Constructs a map of ScheduledFeeds to enable based on the Feeds
// provided from configuration, indexed by the feed type.
func (sc *ScheduledFeedConfig) GetScheduledFeeds() (map[string]feeds.ScheduledFeed, error) {
	scheduledFeeds := map[string]feeds.ScheduledFeed{}
	eventHandler, err := sc.GetEventHandler()
	if err != nil {
		return nil, err
	}

	for _, entry := range sc.Feeds {
		feed, err := entry.ToFeed(eventHandler)
		if err != nil {
			return nil, err
		}
		scheduledFeeds[entry.Type] = feed
	}

	return scheduledFeeds, nil
}

func (sc *ScheduledFeedConfig) GetEventHandler() (*events.Handler, error) {
	var err error
	if sc.EventsConfig == nil {
		sc.eventHandler = events.NewNullHandler()
	} else if sc.eventHandler == nil {
		sc.eventHandler, err = sc.EventsConfig.ToEventHandler()
		if err != nil {
			return nil, err
		}
	}
	return sc.eventHandler, nil
}

func (ec *EventsConfig) ToEventHandler() (*events.Handler, error) {
	var sink events.Sink
	switch ec.Sink {
	case events.LoggingEventSinkType:
		sink = events.NewLoggingEventSink(log.New())
	default:
		return nil, fmt.Errorf("%w : %v", errUnknownSinkType, ec.Sink)
	}
	return events.NewHandler(sink, ec.EventFilter), nil
}

// Produces a Publisher object from the provided PublisherConfig
// The PublisherConfig.Type value is evaluated and the appropriate Publisher is
// constructed from the Config field. If the type is not a recognised Publisher type,
// an error is returned.
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
	case "http-client": // Check for the new publisher type
		return httpclientpubsub.New(pc.HTTPClientConfig)
	default:
		return nil, fmt.Errorf("%w : %v", errUnknownPub, pc.Type)
	}
}

// Constructs the appropriate feed for the given type, providing the
// options to the feed.
func (fc FeedConfig) ToFeed(eventHandler *events.Handler) (feeds.ScheduledFeed, error) {
	switch fc.Type {
	case crates.FeedName:
		return crates.New(fc.Options, eventHandler)
	case goproxy.FeedName:
		return goproxy.New(fc.Options)
	case npm.FeedName:
		return npm.New(fc.Options, eventHandler)
	case nuget.FeedName:
		return nuget.New(fc.Options)
	case pypi.FeedName:
		return pypi.New(fc.Options, eventHandler)
	case packagist.FeedName:
		return packagist.New(fc.Options)
	case rubygems.FeedName:
		return rubygems.New(fc.Options, eventHandler)
	default:
		return nil, fmt.Errorf("%w : %v", errUnknownFeed, fc.Type)
	}
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
		Feeds: []FeedConfig{
			{
				Type:    crates.FeedName,
				Options: defaultFeedOptions,
			},
			{
				Type:    goproxy.FeedName,
				Options: defaultFeedOptions,
			},
			{
				Type:    npm.FeedName,
				Options: npmDefaultFeedOptions,
			},
			{
				Type:    nuget.FeedName,
				Options: defaultFeedOptions,
			},
			{
				Type:    packagist.FeedName,
				Options: defaultFeedOptions,
			},
			{
				Type:    pypi.FeedName,
				Options: defaultFeedOptions,
			},
			{
				Type:    rubygems.FeedName,
				Options: defaultFeedOptions,
			},
		},
		PubConfig: PublisherConfig{
			Type: stdout.PublisherType,
		},
		HTTPPort: 8080,
		PollRate: "5m",
		Timer:    false,
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
