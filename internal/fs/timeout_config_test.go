package fs

import (
	"testing"
	"time"
)

// TestDefaultTimeoutConfig verifies that default timeout configuration is valid
func TestDefaultTimeoutConfig(t *testing.T) {
	config := DefaultTimeoutConfig()

	// Verify all timeouts are positive
	if config.DownloadWorkerShutdown <= 0 {
		t.Error("DownloadWorkerShutdown must be positive")
	}
	if config.UploadGracefulShutdown <= 0 {
		t.Error("UploadGracefulShutdown must be positive")
	}
	if config.FilesystemShutdown <= 0 {
		t.Error("FilesystemShutdown must be positive")
	}
	if config.NetworkCallbackShutdown <= 0 {
		t.Error("NetworkCallbackShutdown must be positive")
	}
	if config.MetadataRequestTimeout <= 0 {
		t.Error("MetadataRequestTimeout must be positive")
	}
	if config.ContentStatsTimeout <= 0 {
		t.Error("ContentStatsTimeout must be positive")
	}

	// Verify default values match expected values
	if config.DownloadWorkerShutdown != 5*time.Second {
		t.Errorf("Expected DownloadWorkerShutdown to be 5s, got %v", config.DownloadWorkerShutdown)
	}
	if config.UploadGracefulShutdown != 30*time.Second {
		t.Errorf("Expected UploadGracefulShutdown to be 30s, got %v", config.UploadGracefulShutdown)
	}
	if config.FilesystemShutdown != 10*time.Second {
		t.Errorf("Expected FilesystemShutdown to be 10s, got %v", config.FilesystemShutdown)
	}
	if config.NetworkCallbackShutdown != 5*time.Second {
		t.Errorf("Expected NetworkCallbackShutdown to be 5s, got %v", config.NetworkCallbackShutdown)
	}
	if config.MetadataRequestTimeout != 30*time.Second {
		t.Errorf("Expected MetadataRequestTimeout to be 30s, got %v", config.MetadataRequestTimeout)
	}
	if config.ContentStatsTimeout != 5*time.Second {
		t.Errorf("Expected ContentStatsTimeout to be 5s, got %v", config.ContentStatsTimeout)
	}

	// Verify configuration is valid
	if err := config.Validate(); err != nil {
		t.Errorf("Default configuration should be valid, got error: %v", err)
	}
}

// TestTimeoutConfigValidation verifies that timeout validation works correctly
func TestTimeoutConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *TimeoutConfig
		expectError bool
		errorField  string
	}{
		{
			name:        "Valid configuration",
			config:      DefaultTimeoutConfig(),
			expectError: false,
		},
		{
			name: "Zero download timeout",
			config: &TimeoutConfig{
				DownloadWorkerShutdown:  0,
				UploadGracefulShutdown:  30 * time.Second,
				FilesystemShutdown:      10 * time.Second,
				NetworkCallbackShutdown: 5 * time.Second,
				MetadataRequestTimeout:  30 * time.Second,
				ContentStatsTimeout:     5 * time.Second,
			},
			expectError: true,
			errorField:  "DownloadWorkerShutdown",
		},
		{
			name: "Negative upload timeout",
			config: &TimeoutConfig{
				DownloadWorkerShutdown:  5 * time.Second,
				UploadGracefulShutdown:  -1 * time.Second,
				FilesystemShutdown:      10 * time.Second,
				NetworkCallbackShutdown: 5 * time.Second,
				MetadataRequestTimeout:  30 * time.Second,
				ContentStatsTimeout:     5 * time.Second,
			},
			expectError: true,
			errorField:  "UploadGracefulShutdown",
		},
		{
			name: "Too short download timeout",
			config: &TimeoutConfig{
				DownloadWorkerShutdown:  500 * time.Millisecond,
				UploadGracefulShutdown:  30 * time.Second,
				FilesystemShutdown:      10 * time.Second,
				NetworkCallbackShutdown: 5 * time.Second,
				MetadataRequestTimeout:  30 * time.Second,
				ContentStatsTimeout:     5 * time.Second,
			},
			expectError: true,
			errorField:  "DownloadWorkerShutdown",
		},
		{
			name: "Too long filesystem timeout",
			config: &TimeoutConfig{
				DownloadWorkerShutdown:  5 * time.Second,
				UploadGracefulShutdown:  30 * time.Second,
				FilesystemShutdown:      10 * time.Minute,
				NetworkCallbackShutdown: 5 * time.Second,
				MetadataRequestTimeout:  30 * time.Second,
				ContentStatsTimeout:     5 * time.Second,
			},
			expectError: true,
			errorField:  "FilesystemShutdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s, got nil", tt.errorField)
				} else {
					// Check that error mentions the expected field
					if invalidErr, ok := err.(*InvalidConfigError); ok {
						if invalidErr.Field != tt.errorField {
							t.Errorf("Expected error for field %s, got error for field %s", tt.errorField, invalidErr.Field)
						}
					} else {
						t.Errorf("Expected InvalidConfigError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

// TestTimeoutConfigInFilesystem verifies that filesystem uses timeout configuration
func TestTimeoutConfigInFilesystem(t *testing.T) {
	// This test verifies that the filesystem is initialized with a timeout configuration
	// We can't easily test the full filesystem initialization here, but we can verify
	// that the default configuration is created correctly

	config := DefaultTimeoutConfig()
	if config == nil {
		t.Fatal("DefaultTimeoutConfig() returned nil")
	}

	// Verify that the configuration can be validated
	if err := config.Validate(); err != nil {
		t.Errorf("Default configuration validation failed: %v", err)
	}
}

// TestInvalidConfigError verifies that InvalidConfigError formats correctly
func TestInvalidConfigError(t *testing.T) {
	err := &InvalidConfigError{
		Field:  "TestField",
		Reason: "test reason",
	}

	expected := "invalid timeout configuration for TestField: test reason"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}
