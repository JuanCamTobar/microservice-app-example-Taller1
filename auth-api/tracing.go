package main

import (
	"net/http"
	"time"

	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	zipkinhttpreporter "github.com/openzipkin/zipkin-go/reporter/http"
)

type TracedClient struct {
	client *http.Client
}

// Implementa HTTPDoer
func (c *TracedClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func initTracing(zipkinURL string) (func(http.Handler) http.Handler, *TracedClient, error) {
	reporter := zipkinhttpreporter.NewReporter(zipkinURL)

	endpoint, err := zipkin.NewEndpoint("auth-api", "")
	if err != nil {
		return nil, nil, err
	}

	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSharedSpans(false),
	)
	if err != nil {
		return nil, nil, err
	}

	serverMiddleware := zipkinhttp.NewServerMiddleware(tracer, zipkinhttp.TagResponseSize(true))

	// IMPORTANTE: esta función devuelve (RoundTripper, error)
	transport, err := zipkinhttp.NewTransport(tracer)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}

	return serverMiddleware, &TracedClient{client: client}, nil
}
