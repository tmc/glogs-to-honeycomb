// Command glogs-to-honeycomb acts as a cloud logging sink and relays sidecar logs to Honeycomb.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

var (
	defaultProjectID        = "your-gcp-project-id"
	defaultSubscriptionID   = "istio-sidecar-log-sink"
	defaultHoneycombDataset = "test-istio-sidecar-logs"
	defaultHoneycombHost    = "https://api.honeycomb.io/"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "issue running:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Prep configuration.
	config := getConfig()
	flag.IntVar(&config.Verbosity, "v", 0, "verbosity level")
	flag.IntVar(&config.ConcurrencyLevel, "concurrency", 1000, "concurrency level")
	flag.Parse()

	// Create subscriber.
	subscriber, err := NewSubscriber(ctx, config)
	if err != nil {
		return err
	}

	// Create debug http server.
	srv, err := NewDebugService()
	if err != nil {
		return err
	}

	// Start Subscription loop.
	go func() {
		defer stop()
		if err := subscriber.Subscribe(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "issue subscribing:", err)
		}
	}()

	// Start http server.
	go func() {
		defer stop()
		// Add custom expvar to expose current counters.
		srv.Publish("stats", Func(func() interface{} { return subscriber.GetStats() }))
		// Implement healthcheck by exposing current stats.
		srv.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(subscriber.GetStats())
		})
		port := "8080"
		if p := os.Getenv("PORT"); p != "" {
			port = p
		}
		if err := http.ListenAndServe(":"+port, srv); err != nil {
			fmt.Fprintln(os.Stderr, "issue starting http server:", err)
		}
	}()

	fmt.Println("servers started.")
	<-ctx.Done()
	err = ctx.Err()
	if err == context.Canceled {
		fmt.Println("shutting down.")
		return nil
	}
	return err
}

func getConfig() Config {
	c := Config{
		ProjectID:        os.Getenv("PROJECT_ID"),
		SubscriptionID:   os.Getenv("PUBSUB_SUBSCRIPTION_ID"),
		HoneycombDataset: os.Getenv("HC_DATASET_NAME"),
		HoneycombHost:    os.Getenv("HC_API_URL"),
		HoneycombAPIKey:  os.Getenv("HC_API_KEY"),
	}
	if c.ProjectID == "" {
		c.ProjectID = defaultProjectID
	}
	if c.SubscriptionID == "" {
		c.SubscriptionID = defaultSubscriptionID
	}
	if c.HoneycombDataset == "" {
		c.HoneycombDataset = defaultHoneycombDataset
	}
	if c.HoneycombHost == "" {
		c.HoneycombHost = defaultHoneycombHost
	}

	return c
}
