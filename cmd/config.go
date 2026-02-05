package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dzeleniak/icu/internal/types"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	DataDir           string  `mapstructure:"data_dir"`
	AutoFetch         bool    `mapstructure:"auto_fetch"`
	APITimeout        int     `mapstructure:"api_timeout"`
	MaxCatalogAge     int     `mapstructure:"max_catalog_age"` // in hours, 0 = no auto-refresh
	TLEEndpoint       string  `mapstructure:"tle_endpoint"`
	SATCATEndpoint    string  `mapstructure:"satcat_endpoint"`
	ObserverLatitude  float64 `mapstructure:"observer_latitude"`  // in degrees
	ObserverLongitude float64 `mapstructure:"observer_longitude"` // in degrees
	ObserverAltitude  float64 `mapstructure:"observer_altitude"`  // in meters above sea level
}

// InitConfig initializes the configuration using Viper
func InitConfig() (*Config, error) {
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

	// Set defaults
	viper.SetDefault("data_dir", configDir)
	viper.SetDefault("auto_fetch", true)
	viper.SetDefault("api_timeout", 30)
	viper.SetDefault("max_catalog_age", 24) // 24 hours default
	viper.SetDefault("tle_endpoint", "https://spacebook.com/api/entity/tle")
	viper.SetDefault("satcat_endpoint", "https://spacebook.com/api/entity/satcat")
	viper.SetDefault("observer_latitude", 0.0)   // degrees
	viper.SetDefault("observer_longitude", 0.0)  // degrees
	viper.SetDefault("observer_altitude", 0.0)   // meters

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

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// IsCatalogStale checks if the catalog is older than the configured max age
// Returns true if the catalog should be refreshed, false otherwise
func (c *Config) IsCatalogStale(catalog *types.Catalog) bool {
	// If max_catalog_age is 0, never auto-refresh
	if c.MaxCatalogAge == 0 {
		return false
	}

	// If catalog is nil, it's considered stale
	if catalog == nil {
		return true
	}

	maxAge := time.Duration(c.MaxCatalogAge) * time.Hour
	age := time.Since(catalog.FetchedAt)

	return age > maxAge
}
