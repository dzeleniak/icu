package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dzeleniak/icu/pkg/satellite"
	"github.com/spf13/viper"
)

// InitConfig initializes the configuration using Viper and returns a satellite.Config.
// This function handles CLI-specific configuration loading from files.
func InitConfig() (*satellite.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".icu")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set config file details
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Get defaults from library
	defaults := satellite.DefaultConfig()

	// Set Viper defaults
	viper.SetDefault("data_dir", configDir)
	viper.SetDefault("auto_fetch", defaults.AutoFetch)
	viper.SetDefault("api_timeout", defaults.APITimeout)
	viper.SetDefault("max_catalog_age", defaults.MaxCatalogAge)
	viper.SetDefault("tle_endpoint", defaults.TLEEndpoint)
	viper.SetDefault("satcat_endpoint", defaults.SATCATEndpoint)
	viper.SetDefault("observer_latitude", defaults.ObserverLatitude)
	viper.SetDefault("observer_longitude", defaults.ObserverLongitude)
	viper.SetDefault("observer_altitude", defaults.ObserverAltitude)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create it with defaults
			configPath := filepath.Join(configDir, "config.yaml")
			if err := viper.SafeWriteConfigAs(configPath); err != nil {
				return nil, fmt.Errorf("failed to create config file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg satellite.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
