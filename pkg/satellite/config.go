package satellite

import "time"

// Config represents satellite catalog configuration.
// This struct can be instantiated programmatically or loaded from a configuration file.
type Config struct {
	DataDir           string  // Directory for storing catalog data
	AutoFetch         bool    // Automatically fetch data if stale or missing
	APITimeout        int     // API request timeout in seconds
	MaxCatalogAge     int     // Maximum catalog age in hours before considered stale (0 = never stale)
	TLEEndpoint       string  // URL for TLE data endpoint
	SATCATEndpoint    string  // URL for SATCAT data endpoint
	ObserverLatitude  float64 // Observer latitude in degrees
	ObserverLongitude float64 // Observer longitude in degrees
	ObserverAltitude  float64 // Observer altitude in meters above sea level
}

// DefaultConfig returns a Config with sensible defaults.
// Users can modify the returned config as needed before use.
func DefaultConfig() *Config {
	return &Config{
		AutoFetch:         true,
		APITimeout:        30,
		MaxCatalogAge:     24,
		TLEEndpoint:       "https://spacebook.com/api/entity/tle",
		SATCATEndpoint:    "https://spacebook.com/api/entity/satcat",
		ObserverLatitude:  0.0,
		ObserverLongitude: 0.0,
		ObserverAltitude:  0.0,
	}
}

// IsCatalogStale checks if the catalog needs refreshing based on age.
// Returns true if the catalog is nil, or if it exceeds MaxCatalogAge.
// Returns false if MaxCatalogAge is 0 (no age limit) or if catalog is fresh.
func (c *Config) IsCatalogStale(catalog *Catalog) bool {
	if c.MaxCatalogAge == 0 {
		return false
	}
	if catalog == nil {
		return true
	}
	maxAge := time.Duration(c.MaxCatalogAge) * time.Hour
	age := time.Since(catalog.FetchedAt)
	return age > maxAge
}
