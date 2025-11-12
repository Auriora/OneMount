package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	yaml "gopkg.in/yaml.v3"

	"github.com/auriora/onemount/internal/errors"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/ui"
)

const (
	// DefaultLogOutput is the default log output destination
	DefaultLogOutput = "STDOUT"
)

type Config struct {
	CacheDir         string `yaml:"cacheDir"`
	LogLevel         string `yaml:"log"`
	LogOutput        string `yaml:"logOutput"`
	SyncTree         bool   `yaml:"syncTree"`
	DeltaInterval    int    `yaml:"deltaInterval"`
	CacheExpiration  int    `yaml:"cacheExpiration"`
	MountTimeout     int    `yaml:"mountTimeout"`
	graph.AuthConfig `yaml:"auth"`
}

// DefaultConfigPath returns the default config location for onemount
func DefaultConfigPath() string {
	confDir, err := os.UserConfigDir()
	if err != nil {
		logging.Error().Err(err).Msg("Could not determine configuration directory.")
	}
	return filepath.Join(confDir, "onemount/config.yml")
}

// createDefaultConfig returns a Config struct with default values
func createDefaultConfig() Config {
	xdgCacheDir, _ := os.UserCacheDir()
	return Config{
		CacheDir:        filepath.Join(xdgCacheDir, "onemount"),
		LogLevel:        "debug",
		LogOutput:       DefaultLogOutput, // Default to standard output
		SyncTree:        true,             // Enable tree sync by default for better performance
		DeltaInterval:   1,                // Default to 1 second
		CacheExpiration: 30,               // Default to 30 days
		MountTimeout:    60,               // Default to 60 seconds
	}
}

// readConfigFile reads the configuration file at the given path
func readConfigFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// parseConfig parses the YAML configuration data into a Config struct
func parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	err := yaml.Unmarshal(data, config)
	return config, err
}

// mergeWithDefaults merges the parsed configuration with the defaults
func mergeWithDefaults(config *Config, defaults Config) error {
	return mergo.Merge(config, defaults)
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate LogLevel
	validLogLevels := LogLevels()
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if strings.ToLower(config.LogLevel) == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		logging.Warn().
			Str("logLevel", config.LogLevel).
			Interface("validLevels", validLogLevels).
			Msg("Invalid log level, using default.")
		config.LogLevel = "debug"
	}

	// Validate LogOutput
	if config.LogOutput == "" {
		logging.Warn().Msg("Log output location cannot be empty, using default (STDOUT).")
		config.LogOutput = DefaultLogOutput
	} else {
		// Normalize special values to uppercase
		switch strings.ToUpper(config.LogOutput) {
		case "STDOUT", "STDERR":
			config.LogOutput = strings.ToUpper(config.LogOutput)
		default:
			// For file paths, ensure the directory exists
			logDir := filepath.Dir(config.LogOutput)
			if logDir != "." {
				if err := os.MkdirAll(logDir, 0755); err != nil {
					logging.Warn().
						Err(err).
						Str("logOutput", config.LogOutput).
						Msg("Could not create directory for log file, using STDOUT.")
					config.LogOutput = DefaultLogOutput
				}
			}
		}
	}

	// Validate DeltaInterval
	if config.DeltaInterval <= 0 {
		logging.Warn().
			Int("deltaInterval", config.DeltaInterval).
			Msg("Delta interval must be positive, using default.")
		config.DeltaInterval = 1
	}

	// Validate CacheExpiration
	if config.CacheExpiration < 0 {
		logging.Warn().
			Int("cacheExpiration", config.CacheExpiration).
			Msg("Cache expiration must be non-negative, using default.")
		config.CacheExpiration = 30
	}

	// Validate MountTimeout
	if config.MountTimeout <= 0 {
		logging.Warn().
			Int("mountTimeout", config.MountTimeout).
			Msg("Mount timeout must be positive, using default.")
		config.MountTimeout = 60
	}

	// Validate CacheDir
	if config.CacheDir == "" {
		logging.Warn().Msg("Cache directory cannot be empty, using default.")
		xdgCacheDir, _ := os.UserCacheDir()
		config.CacheDir = filepath.Join(xdgCacheDir, "onemount")
	}

	// Validate AuthConfig if provided
	if config.AuthConfig.ClientID != "" {
		if config.AuthConfig.CodeURL == "" || config.AuthConfig.TokenURL == "" || config.AuthConfig.RedirectURL == "" {
			return errors.New("incomplete auth configuration: all auth fields must be provided if any are set")
		}
	}

	return nil
}

// LoadConfig is the primary way of loading onemount's config
func LoadConfig(path string) *Config {
	// Create default configuration
	defaults := createDefaultConfig()

	// Read configuration file
	conf, err := readConfigFile(path)
	if err != nil {
		logging.Warn().
			Err(err).
			Str("path", path).
			Msg("Configuration file not found, using defaults.")
		return &defaults
	}

	// Parse configuration
	config, err := parseConfig(conf)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", path).
			Msg("Could not parse configuration file, using defaults.")
		return &defaults
	}

	// Merge with defaults
	if err = mergeWithDefaults(config, defaults); err != nil {
		logging.Error().
			Err(err).
			Str("path", path).
			Msg("Could not merge configuration file with defaults, using defaults only.")
		return &defaults
	}

	// Process CacheDir (unescape home directory)
	config.CacheDir = ui.UnescapeHome(config.CacheDir)

	// Validate configuration
	if err = validateConfig(config); err != nil {
		logging.Error().
			Err(err).
			Str("path", path).
			Msg("Invalid configuration, using defaults.")
		return &defaults
	}

	return config
}

// WriteConfig - Write config to a file
func (c Config) WriteConfig(path string) error {
	out, err := yaml.Marshal(c)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", path).
			Msg("Could not marshal config!")
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", path).
			Msg("Could not create directory for config file.")
		return err
	}

	err = os.WriteFile(path, out, 0600)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", path).
			Msg("Could not write config to disk.")
		return err
	}

	logging.Debug().
		Str("path", path).
		Msg("Configuration written to file.")
	return nil
}
