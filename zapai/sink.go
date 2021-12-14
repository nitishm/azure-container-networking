package zapai

import (
	"context"
	"encoding/gob"
	"net/url"
	"strconv"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// SinkScheme is the registerable Sink Scheme for appinsights.
	SinkScheme            = "appinsights"
	paramEndpointURL      = "endpointURL"
	paramMaxBatchInterval = "maxBatchInterval"
	paramMaxBatchSize     = "maxBatchSize"
	paramGracePeriod      = "gracePeriod"
)

func init() {
	// register the appinsights sink factory
	_ = zap.RegisterSink(SinkScheme, sinkbuilder)
	gob.Register(appinsights.TraceTelemetry{})
}

// sinkbuilder builds an appinsights Sink for zap from the passed URL.
// The URL is expected to be parsed and passed by zap.Open, which should be passed a URI generated by the
// SinkConfig.URI() method.
//
// This is effectively a convenience to plug in to the zap constructor machinery. By accepting a URI, we can set up a
// SinkConfig, and let zap.Open "build" the Sink from the URI representing that SinkConfig. "Build" here really means
// that zap.Open will call our registered factory method, but it performs some decorating on the result that is useful.
func sinkbuilder(u *url.URL) (zap.Sink, error) {
	cfg, err := fromURI(u)
	if err != nil {
		return nil, err
	}
	return newSink(cfg), nil
}

// SinkConfig is a container struct for an appinsights Sink configuration.
type SinkConfig struct {
	GracePeriod time.Duration
	appinsights.TelemetryConfiguration
}

// fromURI parses a URL in to a SinkConfig.
func fromURI(u *url.URL) (*SinkConfig, error) {
	q := u.Query()
	gracePeriod, err := time.ParseDuration(q.Get(paramGracePeriod))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration parameter")
	}
	interval, err := time.ParseDuration(q.Get(paramMaxBatchInterval))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration parameter")
	}
	size, err := strconv.Atoi(q.Get(paramMaxBatchSize))
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert size parameter to int")
	}
	endpoint, err := url.QueryUnescape(q.Get(paramEndpointURL))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unescape url parameter")
	}
	return &SinkConfig{
		GracePeriod: gracePeriod,
		TelemetryConfiguration: appinsights.TelemetryConfiguration{
			InstrumentationKey: u.User.Username(),
			EndpointUrl:        endpoint,
			MaxBatchInterval:   interval,
			MaxBatchSize:       size,
		},
	}, nil
}

// toURI generates a URL from a SinkConfig.
func toURI(sc *SinkConfig) *url.URL {
	u := &url.URL{
		Scheme: SinkScheme,
		User:   url.User(sc.InstrumentationKey),
	}
	q := u.Query()
	q.Add(paramEndpointURL, url.QueryEscape(sc.EndpointUrl))
	q.Add(paramGracePeriod, url.QueryEscape(sc.GracePeriod.String()))
	q.Add(paramMaxBatchInterval, url.QueryEscape(sc.MaxBatchInterval.String()))
	q.Add(paramMaxBatchSize, url.QueryEscape(strconv.Itoa(sc.MaxBatchSize)))
	u.RawQuery = q.Encode()
	return u
}

// URI builds an appinsights Sink URI string suitable for passing to zap.Open.
func (sc *SinkConfig) URI() string {
	return toURI(sc).String()
}

// telemetryTracker is a client interface that the appinsights.TelemetryClient implements implicitly.
// It is defined here to restrict the available methods in the Sink from that TelemetryClient, and for testing.
type telemetryTracker interface {
	Track(appinsights.Telemetry)
	Channel() appinsights.TelemetryChannel
}

var _ zap.Sink = (*Sink)(nil)

// Sink implements zap.Sink for appinsights.
// Sink is not inherently safe for concurrent use - the zap constructors will wrap it for concurrency.
//
// To conform to the zap.Sink interface, Sink implements Write([]byte), where it expects that the passed []byte is
// an encoded gob that can be decoded in to a valid appinsights.TraceTelemetry. Passing any other []byte input is
// an error.
type Sink struct {
	*SinkConfig
	cli telemetryTracker
	dec traceDecoder
}

// newSink constructs a Sink from the passed SinkConfig
func newSink(cfg *SinkConfig) *Sink {
	return &Sink{
		SinkConfig: cfg,
		cli:        appinsights.NewTelemetryClientFromConfig(&cfg.TelemetryConfiguration),
		dec:        newTraceDecoder(),
	}
}

// Write accepts a gob []byte that must be Decodable to an appinsights.TraceTelemetry{}, which is then sent
// to appinsights via the telemetryTracker.
func (s *Sink) Write(b []byte) (int, error) {
	t, err := s.dec.decode(b)
	if err != nil {
		return 0, errors.Wrap(err, "sink failed to decode trace")
	}
	s.cli.Track(t)
	return 0, nil
}

// Sync flushes the current channel queue.
func (s *Sink) Sync() error {
	s.cli.Channel().Flush()
	return nil
}

// Close flushes and tears down the appinsights channel.
// Waits up to the GracePeriod duration for sends and retries to complete.
func (s *Sink) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.GracePeriod)
	defer cancel()
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "sink close context timeout")
	case <-s.cli.Channel().Close(s.GracePeriod):
		return nil
	}
}
