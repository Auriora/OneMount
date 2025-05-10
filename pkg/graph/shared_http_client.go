package graph

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

// getSharedHTTPClient returns the shared HTTP client with connection pooling
func getSharedHTTPClient() HTTPClient {
	var sharedHTTPClient HTTPClient
	clientOnce.Do(func() {
		// Create a custom transport with connection pooling settings
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		}

		// Create the shared client with the custom transport and timeout
		sharedHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   defaultRequestTimeout,
		}

		log.Info().Msg("Initialized shared HTTP client with connection pooling")
	})

	return sharedHTTPClient
}
