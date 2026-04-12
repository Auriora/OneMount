package common

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_CMD_02_01_Config_ValidConfigFile_LoadsCorrectValues verifies that configuration can be loaded from a file.
//
//	Test Case ID    UT-CMD-02-01
//	Title           Configuration Loading
//	Description     Tests loading configuration from a file
//	Preconditions   None
//	Steps           1. Load configuration from a test config file
//	                2. Get the user's home directory
//	                3. Check if the loaded configuration matches expected values
//	Expected Result The configuration values match the expected values from the config file
//	Notes: This test verifies the functionality for loading configuration from a file.
func TestUT_CMD_02_01_Config_ValidConfigFile_LoadsCorrectValues(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("ConfigLoadingFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "onemount-test-*")
		if err != nil {
			return nil, err
		}

		// Create a test config file
		configPath := filepath.Join(tempDir, "config.json")

		return map[string]interface{}{
			"tempDir":    tempDir,
			"configPath": configPath,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		return os.RemoveAll(tempDir)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, f interface{}) {
		fixtureData := f.(*framework.UnitTestFixture)
		data := fixtureData.SetupData.(map[string]interface{})
		tempDir := data["tempDir"].(string)

		// Create a valid YAML config file with specific values
		configPath := filepath.Join(tempDir, "config.yml")
		configContent := `cacheDir: /tmp/onemount-test-cache
log: info
logOutput: STDOUT
deltaInterval: 600
cacheExpiration: 14
mountTimeout: 120
overlay:
  defaultPolicy: REMOTE_WINS
`
		if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		// Load the configuration
		config := LoadConfig(configPath)
		if config == nil {
			t.Fatal("LoadConfig returned nil")
		}

		// Verify loaded values
		if config.LogLevel != "info" {
			t.Errorf("Expected logLevel 'info', got '%s'", config.LogLevel)
		}
		if config.DeltaInterval != 600 {
			t.Errorf("Expected deltaInterval 600, got %d", config.DeltaInterval)
		}
		if config.CacheExpiration != 14 {
			t.Errorf("Expected cacheExpiration 14, got %d", config.CacheExpiration)
		}
		if config.MountTimeout != 120 {
			t.Errorf("Expected mountTimeout 120, got %d", config.MountTimeout)
		}
	})
}

// TestUT_CMD_03_01_Config_MergedSettings_ContainsMergedValues verifies that configuration settings can be merged.
//
//	Test Case ID    UT-CMD-03-01
//	Title           Configuration Merging
//	Description     Tests merging configuration settings
//	Preconditions   None
//	Steps           1. Load configuration from a test config file with merged settings
//	                2. Check if the loaded configuration contains the merged values
//	Expected Result The configuration contains the merged values
//	Notes: This test verifies the functionality for merging configuration settings.
func TestUT_CMD_03_01_Config_MergedSettings_ContainsMergedValues(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("ConfigMergingFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "onemount-test-*")
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"tempDir": tempDir,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		return os.RemoveAll(tempDir)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, f interface{}) {
		fixtureData := f.(*framework.UnitTestFixture)
		data := fixtureData.SetupData.(map[string]interface{})
		tempDir := data["tempDir"].(string)

		// Create a config file with only partial settings (others should get defaults)
		configPath := filepath.Join(tempDir, "partial-config.yml")
		configContent := `log: warn
deltaInterval: 900
overlay:
  defaultPolicy: LOCAL_WINS
`
		if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config := LoadConfig(configPath)
		if config == nil {
			t.Fatal("LoadConfig returned nil")
		}

		// Verify explicitly set values
		if config.LogLevel != "warn" {
			t.Errorf("Expected logLevel 'warn', got '%s'", config.LogLevel)
		}
		if config.DeltaInterval != 900 {
			t.Errorf("Expected deltaInterval 900, got %d", config.DeltaInterval)
		}
		if config.Overlay.DefaultPolicy != "LOCAL_WINS" {
			t.Errorf("Expected overlay policy 'LOCAL_WINS', got '%s'", config.Overlay.DefaultPolicy)
		}

		// Verify defaults were merged for unset values
		defaults := createDefaultConfig()
		if config.CacheExpiration != defaults.CacheExpiration {
			t.Errorf("Expected default cacheExpiration %d, got %d", defaults.CacheExpiration, config.CacheExpiration)
		}
		if config.MountTimeout != defaults.MountTimeout {
			t.Errorf("Expected default mountTimeout %d, got %d", defaults.MountTimeout, config.MountTimeout)
		}
		if config.ActiveDeltaInterval != defaults.ActiveDeltaInterval {
			t.Errorf("Expected default activeDeltaInterval %d, got %d", defaults.ActiveDeltaInterval, config.ActiveDeltaInterval)
		}
	})
}

// TestUT_CMD_04_01_Config_NonexistentFile_LoadsDefaultValues verifies that default configuration is loaded when the config file doesn't exist.
//
//	Test Case ID    UT-CMD-04-01
//	Title           Default Configuration Loading
//	Description     Tests loading default configuration when the config file doesn't exist
//	Preconditions   None
//	Steps           1. Load configuration from a nonexistent config file
//	                2. Get the user's home directory
//	                3. Check if the loaded configuration contains default values
//	Expected Result The configuration contains the default values
//	Notes: This test verifies the functionality for loading default configuration when the config file doesn't exist.
func TestUT_CMD_04_01_Config_NonexistentFile_LoadsDefaultValues(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("DefaultConfigLoadingFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "onemount-test-*")
		if err != nil {
			return nil, err
		}

		// Path to a nonexistent config file
		configPath := filepath.Join(tempDir, "nonexistent-config.json")

		return map[string]interface{}{
			"tempDir":    tempDir,
			"configPath": configPath,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		return os.RemoveAll(tempDir)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, f interface{}) {
		fixtureData := f.(*framework.UnitTestFixture)
		data := fixtureData.SetupData.(map[string]interface{})
		configPath := data["configPath"].(string)

		// Load config from a path that doesn't exist yet — should create defaults
		config := LoadConfig(configPath)
		if config == nil {
			t.Fatal("LoadConfig returned nil for nonexistent file")
		}

		defaults := createDefaultConfig()

		// Verify default values are loaded
		if config.DeltaInterval != defaults.DeltaInterval {
			t.Errorf("Expected default deltaInterval %d, got %d", defaults.DeltaInterval, config.DeltaInterval)
		}
		if config.CacheExpiration != defaults.CacheExpiration {
			t.Errorf("Expected default cacheExpiration %d, got %d", defaults.CacheExpiration, config.CacheExpiration)
		}
		if config.MountTimeout != defaults.MountTimeout {
			t.Errorf("Expected default mountTimeout %d, got %d", defaults.MountTimeout, config.MountTimeout)
		}
		if config.LogLevel != defaults.LogLevel {
			t.Errorf("Expected default logLevel '%s', got '%s'", defaults.LogLevel, config.LogLevel)
		}
		if config.ActiveDeltaInterval != defaults.ActiveDeltaInterval {
			t.Errorf("Expected default activeDeltaInterval %d, got %d", defaults.ActiveDeltaInterval, config.ActiveDeltaInterval)
		}
	})
}

// TestUT_CMD_05_01_Config_ValidSettings_WritesSuccessfully verifies that configuration can be written to a file.
//
//	Test Case ID    UT-CMD-05-01
//	Title           Configuration Writing
//	Description     Tests writing a configuration file
//	Preconditions   None
//	Steps           1. Load configuration from a test config file
//	                2. Write the configuration to a new file
//	                3. Check if the write operation succeeds
//	Expected Result The configuration is successfully written to the file
//	Notes: This test verifies the functionality for writing configuration to a file.
func TestUT_CMD_05_01_Config_ValidSettings_WritesSuccessfully(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("ConfigWritingFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "onemount-test-*")
		if err != nil {
			return nil, err
		}

		// Create a test config file
		configPath := filepath.Join(tempDir, "config.json")
		outputPath := filepath.Join(tempDir, "output-config.json")

		return map[string]interface{}{
			"tempDir":    tempDir,
			"configPath": configPath,
			"outputPath": outputPath,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		return os.RemoveAll(tempDir)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, f interface{}) {
		fixtureData := f.(*framework.UnitTestFixture)
		data := fixtureData.SetupData.(map[string]interface{})
		outputPath := data["outputPath"].(string)

		// Create a config with known values
		config := createDefaultConfig()
		config.LogLevel = "warn"
		config.DeltaInterval = 900
		config.CacheExpiration = 7

		// Write the configuration
		err := config.WriteConfig(outputPath)
		if err != nil {
			t.Fatalf("WriteConfig failed: %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Fatal("Config file was not created")
		}

		// Read back and verify
		reloaded := LoadConfig(outputPath)
		if reloaded == nil {
			t.Fatal("Failed to reload written config")
		}
		if reloaded.LogLevel != "warn" {
			t.Errorf("Expected logLevel 'warn' after reload, got '%s'", reloaded.LogLevel)
		}
		if reloaded.DeltaInterval != 900 {
			t.Errorf("Expected deltaInterval 900 after reload, got %d", reloaded.DeltaInterval)
		}
		if reloaded.CacheExpiration != 7 {
			t.Errorf("Expected cacheExpiration 7 after reload, got %d", reloaded.CacheExpiration)
		}
	})
}

func TestUT_CMD_Config_DefaultDeltaIntervalIsFiveMinutes(t *testing.T) {
	cfg := createDefaultConfig()
	expected := int((5 * time.Minute).Seconds())
	if cfg.DeltaInterval != expected {
		t.Fatalf("expected default delta interval %d seconds, got %d", expected, cfg.DeltaInterval)
	}
}

func TestUT_CMD_Config_ValidateConfigOverlayPolicy(t *testing.T) {
	cfg := createDefaultConfig()
	cfg.Overlay.DefaultPolicy = "local_wins"
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
	if cfg.Overlay.DefaultPolicy != "LOCAL_WINS" {
		t.Fatalf("expected overlay policy normalized to LOCAL_WINS, got %s", cfg.Overlay.DefaultPolicy)
	}

	cfg.Overlay.DefaultPolicy = "invalid"
	if err := validateConfig(&cfg); err == nil {
		t.Fatalf("expected error for invalid overlay policy")
	}
}

func TestUT_CMD_Config_ValidateRealtimeConfigDefaults(t *testing.T) {
	cfg := &RealtimeConfig{
		Enabled:          true,
		Resource:         "",
		FallbackInterval: 0,
	}
	if err := validateRealtimeConfig(cfg); err != nil {
		t.Fatalf("unexpected error validating realtime config: %v", err)
	}
	if cfg.Resource == "" {
		t.Fatalf("expected resource default to be set")
	}
	if cfg.FallbackInterval <= 0 {
		t.Fatalf("expected fallback interval default to be set")
	}
	if cfg.ClientState == "" {
		t.Fatalf("expected client state to be generated when enabled")
	}
}

func TestUT_CMD_Config_DefaultActiveDeltaTuning(t *testing.T) {
	cfg := createDefaultConfig()
	if cfg.ActiveDeltaInterval != 60 {
		t.Fatalf("expected default active delta interval 60 seconds, got %d", cfg.ActiveDeltaInterval)
	}
	if cfg.ActiveDeltaWindow != 120 {
		t.Fatalf("expected default active delta window 120 seconds, got %d", cfg.ActiveDeltaWindow)
	}
}

func TestUT_CMD_Config_ValidateConfigResetsInvalidActiveDeltaTuning(t *testing.T) {
	cfg := createDefaultConfig()
	cfg.ActiveDeltaInterval = -5
	cfg.ActiveDeltaWindow = 0
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
	if cfg.ActiveDeltaInterval != 60 {
		t.Fatalf("active delta interval not reset; got %d", cfg.ActiveDeltaInterval)
	}
	if cfg.ActiveDeltaWindow != 120 {
		t.Fatalf("active delta window not reset; got %d", cfg.ActiveDeltaWindow)
	}
}

func TestUT_CMD_Config_HydrationConfigDefaultsAndValidation(t *testing.T) {
	cfg := createDefaultConfig()
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
	if cfg.Hydration.Workers != 4 || cfg.Hydration.QueueSize != 500 {
		t.Fatalf("unexpected hydration defaults: %+v", cfg.Hydration)
	}

	cfg.Hydration.Workers = -1
	cfg.Hydration.QueueSize = 0
	if err := validateConfig(&cfg); err == nil {
		t.Fatalf("expected error after invalid hydration settings")
	}
}

func TestUT_CMD_Config_MetadataQueueDefaultsAndValidation(t *testing.T) {
	cfg := createDefaultConfig()
	cfg.MetadataQueue.Workers = -1
	cfg.MetadataQueue.HighPrioritySize = 0
	cfg.MetadataQueue.LowPrioritySize = -10
	if err := validateConfig(&cfg); err == nil {
		t.Fatalf("expected error for invalid metadata queue values")
	}
}

func TestUT_CMD_Config_RealtimeFallbackValidationBounds(t *testing.T) {
	cfg := createDefaultConfig()
	cfg.Realtime.FallbackInterval = 10
	if err := validateConfig(&cfg); err == nil {
		t.Fatalf("expected error for fallback interval below minimum")
	}

	cfg = createDefaultConfig()
	cfg.Realtime.FallbackInterval = int((3 * time.Hour).Seconds())
	if err := validateConfig(&cfg); err == nil {
		t.Fatalf("expected error for fallback interval above maximum")
	}

	cfg = createDefaultConfig()
	cfg.Realtime.FallbackInterval = 120
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("validateConfig returned error for valid fallback interval: %v", err)
	}
}
