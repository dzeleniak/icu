// Package satellite provides satellite catalog data management and orbital propagation.
//
// This package offers functionality for fetching, storing, searching, and analyzing satellite
// catalog data including TLE (Two-Line Element) sets and SATCAT (Satellite Catalog) information.
// It includes SGP4 orbital propagation for calculating satellite positions and visibility.
//
// # Basic Usage
//
// Fetch and merge satellite catalog data:
//
//	client := satellite.NewClient(
//	    "https://spacebook.com/api/entity/tle",
//	    "https://spacebook.com/api/entity/satcat",
//	    30*time.Second,
//	)
//
//	catalog, err := satellite.FetchAndMergeCatalog(client)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Search for satellites:
//
//	results := satellite.SearchSatellites(catalog.Satellites, satellite.SearchCriteria{
//	    Name:   "starlink",
//	    Type:   "payload",
//	    Regime: "LEO",
//	})
//
// Propagate satellite position:
//
//	if len(results) > 0 && results[0].TLE != nil {
//	    pos, err := satellite.PropagateSatellite(results[0].TLE, time.Now())
//	    if err == nil {
//	        fmt.Printf("Position: X=%.2f, Y=%.2f, Z=%.2f km\n", pos.X, pos.Y, pos.Z)
//	    }
//	}
//
// Calculate observation angles:
//
//	observer := &satellite.ObserverPosition{
//	    Latitude:  40.7128,  // New York City
//	    Longitude: -74.0060,
//	    Altitude:  10.0,     // meters
//	}
//
//	angles := satellite.CalculateObservationAngles(pos, observer)
//	fmt.Printf("Azimuth: %.2f°, Elevation: %.2f°\n", angles.Azimuth, angles.Elevation)
//
// Find visible satellites:
//
//	visible, err := satellite.FindVisibleSatellites(
//	    catalog.Satellites,
//	    observer,
//	    time.Now(),
//	    satellite.VisibilityCriteria{
//	        MinElevation: 10.0,  // Above 10° elevation
//	        MaxElevation: 90.0,
//	    },
//	)
//
// # Persistence
//
// Save and load catalog data:
//
//	storage, err := satellite.NewStorage("/path/to/data")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save catalog
//	if err := storage.Save(catalog); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load catalog
//	catalog, err := storage.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
package satellite
