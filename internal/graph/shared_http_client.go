package graph

import (
	"github.com/auriora/onemount/internal/logging"
	"net/http"
	"time"
)

var defaultHTTPClient HTTPClient

// getSharedHTTPClient returns the shared HTTP client with connection pooling
func getSharedHTTPClient() HTTPClient {
	clientOnce.Do(func() {
		// Create a custom transport with connection pooling settings
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		}

		// Create the shared client with the custom transport and timeout
		defaultHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   defaultRequestTimeout,
		}

		logging.Info().Msg("Initialized shared HTTP client with connection pooling")
	})

	return defaultHTTPClient
}
