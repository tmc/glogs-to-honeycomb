package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/honeycombio/libhoney-go"
)

// ErrEventSkippedDueToSampling is a sentinel error that represents a skipped message.
var ErrEventSkippedDueToSampling = errors.New("skipped due to sampling")

// Subscriber subscribes to a pubsub topic and ships the paylaods to honeycomb.
type Subscriber struct {
	config           Config
	honeycombClient  *libhoney.Client
	honeycombDataset string

	Stats *SubscriberStats
}

// Config describes the configuration for a Subscriber.
type Config struct {
	Verbosity int

	ProjectID      string
	SubscriptionID string

	HoneycombDataset string
	HoneycombHost    string
	HoneycombAPIKey  string

	DefaultSampleRate int

	ConcurrencyLevel int
}

// NewSubscriber returns a prepared server.
func NewSubscriber(ctx context.Context, config Config) (*Subscriber, error) {
	s := &Subscriber{
		config: config,
		Stats:  &SubscriberStats{},
	}
	if config.HoneycombAPIKey == "" {
		return nil, fmt.Errorf("missing Honeycomb API Key (HC_API_KEY)")
	}
	if s.config.DefaultSampleRate == 0 {
		s.config.DefaultSampleRate = 10
	}
	// Initialize Honeycomb.
	s.honeycombDataset = os.Getenv("HC_DATASET_NAME")
	if s.honeycombDataset == "" {
		s.honeycombDataset = defaultHoneycombDataset
	}
	if s.config.Verbosity > 0 {
		fmt.Println("honeycomb dataset:", s.honeycombDataset)
	}

	hcConfig := libhoney.ClientConfig{
		APIKey:  config.HoneycombAPIKey,
		Dataset: s.honeycombDataset,
		APIHost: config.HoneycombHost,
		// Transmission: hcTransmission,
	}
	if s.config.Verbosity > 0 {
		hcConfig.Logger = &libhoney.DefaultLogger{}
	}
	var err error
	if s.honeycombClient, err = libhoney.NewClient(hcConfig); err != nil {
		return nil, err
	}

	return s, nil
}

// Subscribe initiates the polling loop to fetch and process messages.
func (s *Subscriber) Subscribe(ctx context.Context) error {
	client, err := pubsub.NewClient(ctx, s.config.ProjectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	sub := client.Subscription(s.config.SubscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = s.config.ConcurrencyLevel
	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		s.Stats.AddTotal(1)
		if err := s.ShipToHoneycomb(ctx, msg); err != nil {
			// If we got an unexpected error, record stat, print, and nack.
			if err != ErrEventSkippedDueToSampling {
				s.Stats.AddError(1)
				fmt.Fprintln(os.Stderr, err)
				msg.Nack()
				return
			}
		} else {
			s.Stats.AddSampled(1) // record that we kept this message.
		}
		msg.Ack()
	})
}

// ShipToHoneycomb initiates the polling loop to fetch and process messages.
func (s *Subscriber) ShipToHoneycomb(ctx context.Context, msg *pubsub.Message) error {
	payload := CloudLoggingPayload{}

	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		fmt.Fprintln(os.Stderr, "issue unmarshaling pubsub payload:", err)
		return err
	}
	e := s.honeycombClient.NewEvent()
	if t, err := time.Parse(time.RFC3339Nano, msg.Attributes["logging.googleapis.com/timestamp"]); err == nil {
		e.Timestamp = t
	} else {
		e.Timestamp = msg.PublishTime
	}
	payload = unnestFields(payload)
	fixFieldTypes(payload)
	if s.config.Verbosity > 1 {
		json.NewEncoder(os.Stdout).Encode(payload)
	}
	e.Add(payload)
	e.AddField("glogs-to-honeycomb-version", 2)
	e.SampleRate = ApplySamplingPolicy(payload, s.config.DefaultSampleRate)

	sampled := shouldKeep(e.SampleRate)
	if !sampled {
		return ErrEventSkippedDueToSampling
	}
	return e.SendPresampled()
}

// returns true if the sample should be kept
func shouldKeep(rate uint) bool {
	if rate <= 1 {
		return true
	}

	return rand.Intn(int(rate)) == 0
}

// GetStats returns a snapshot of current subscriber statistics.
func (s *Subscriber) GetStats() SubscriberStats {
	if s.Stats == nil {
		return SubscriberStats{}
	}
	return s.Stats.Copy()
}

func unnestFields(l map[string]interface{}) map[string]interface{} {
	return unnestFieldsN(l, 2)
}

// unnestFieldsN unnests the given map to the specified depth.
func unnestFieldsN(l map[string]interface{}, depth int) map[string]interface{} {
	result := map[string]interface{}{}
	marshaledJSON, _ := json.Marshal(l)
	json.Unmarshal(marshaledJSON, &result)

	if depth == 0 {
		return result
	}
	for k, v := range result {
		v, ok := v.(map[string]interface{})
		if ok {
			v := unnestFieldsN(v, depth-1)
			for kk, vv := range v {
				newK := fmt.Sprintf("%s.%s", k, kk)
				result[newK] = vv
			}
			delete(result, k)
		}
	}
	return result
}

var numericFieldPaths = map[string]bool{
	"jsonPayload.bytes_received":        true,
	"jsonPayload.bytes_sent":            true,
	"jsonPayload.duration":              true,
	"jsonPayload.response_code":         true,
	"jsonPayload.upstream_service_time": true,
}

// fixFieldTypes changes the type of a set of fields to be numeric in the payload.
func fixFieldTypes(p CloudLoggingPayload) {
	for numericField := range numericFieldPaths {
		v, ok := p[numericField].(string)
		if !ok {
			continue
		}
		intVal, err := strconv.Atoi(v)
		p[numericField] = intVal
		if err != nil {
			delete(p, numericField)
		}
	}
}
