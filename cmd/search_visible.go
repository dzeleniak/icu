package cmd

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/dzeleniak/icu/internal/propagate"
	"github.com/dzeleniak/icu/internal/storage"
	"github.com/dzeleniak/icu/internal/types"
	"github.com/spf13/cobra"
)

var (
	visibleName         string
	visibleOwner        string
	visibleType         string
	visibleRegime       string
	visibleMinElevation float64
	visibleMaxElevation float64
	visibleLimit        int
	visibleVerbose      bool
)

var visibleCmd = &cobra.Command{
	Use:   "visible",
	Short: "Search for satellites currently visible from observer location",
	Long: `Search for satellites currently overhead based on observer location from config.
Propagates satellites to current time and checks if they are visible (above minimum elevation).
Supports all standard search filters (name, owner, type, regime) plus elevation constraints.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSearchVisible()
	},
}

func init() {
	searchCmd.AddCommand(visibleCmd)
	visibleCmd.Flags().StringVarP(&visibleName, "name", "n", "", "Search by satellite name (partial match, case-insensitive)")
	visibleCmd.Flags().StringVarP(&visibleOwner, "owner", "o", "", "Filter by owner/country code")
	visibleCmd.Flags().StringVarP(&visibleType, "type", "t", "", "Filter by object type (PAYLOAD, ROCKET BODY, DEBRIS)")
	visibleCmd.Flags().StringVarP(&visibleRegime, "regime", "r", "", "Filter by orbital regime (LEO, MEO, GEO, HEO)")
	visibleCmd.Flags().Float64Var(&visibleMinElevation, "min-elevation", 10.0, "Minimum elevation angle in degrees")
	visibleCmd.Flags().Float64Var(&visibleMaxElevation, "max-elevation", 90.0, "Maximum elevation angle in degrees")
	visibleCmd.Flags().IntVarP(&visibleLimit, "limit", "l", 0, "Maximum number of results to display (0 = no limit)")
	visibleCmd.Flags().BoolVarP(&visibleVerbose, "verbose", "v", false, "Display verbose satellite information")
}

// VisibleSatellite holds a satellite and its current observation angles
type VisibleSatellite struct {
	Satellite *types.Satellite
	Angles    *propagate.ObservationAngles
}

func runSearchVisible() {
	// Check observer configuration
	if config.ObserverLatitude == 0.0 && config.ObserverLongitude == 0.0 {
		fmt.Println("Observer location not configured.")
		fmt.Println("Please set observer_latitude, observer_longitude, and observer_altitude in ~/.icu/config.yaml")
		return
	}

	observer := &propagate.ObserverPosition{
		Latitude:  config.ObserverLatitude,
		Longitude: config.ObserverLongitude,
		Altitude:  config.ObserverAltitude,
	}

	// Load catalog
	store, err := storage.NewStorage(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	catalog, err := store.Load()
	if err != nil {
		log.Fatalf("Error loading catalog: %v", err)
	}

	if catalog == nil {
		fmt.Println("No catalog found. Run 'icu fetch' to download data.")
		return
	}

	// Apply search filters to narrow down candidates
	candidates := searchSatellites(catalog.Satellites, visibleName, visibleOwner, visibleType, visibleRegime)

	if len(candidates) == 0 {
		fmt.Println("No satellites found matching the search criteria.")
		return
	}

	fmt.Printf("Checking %d satellites for visibility...\n", len(candidates))

	// Get current time
	now := time.Now()

	// Check visibility for each candidate
	visible := make([]*VisibleSatellite, 0)

	for _, sat := range candidates {
		if sat.TLE == nil {
			continue // Skip satellites without TLE data
		}

		// Propagate satellite to current time
		pos, err := propagate.PropagateSatellite(sat.TLE, now)
		if err != nil {
			// Skip satellites that fail to propagate
			continue
		}

		// Calculate observation angles
		angles := propagate.CalculateObservationAngles(pos, observer)

		// Check if within elevation bounds
		if angles.Elevation >= visibleMinElevation && angles.Elevation <= visibleMaxElevation {
			visible = append(visible, &VisibleSatellite{
				Satellite: sat,
				Angles:    angles,
			})
		}
	}

	if len(visible) == 0 {
		fmt.Printf("\nNo satellites currently visible (elevation between %.1f° and %.1f°).\n",
			visibleMinElevation, visibleMaxElevation)
		return
	}

	// Sort by elevation (highest first)
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].Angles.Elevation > visible[j].Angles.Elevation
	})

	// Limit results
	displayCount := len(visible)
	if visibleLimit > 0 && displayCount > visibleLimit {
		displayCount = visibleLimit
	}

	// Display results
	fmt.Printf("\nFound %d visible satellites", len(visible))
	if visibleLimit > 0 && len(visible) > visibleLimit {
		fmt.Printf(" (showing first %d)", visibleLimit)
	}
	fmt.Printf("\nObserver: %.4f°N, %.4f°E, %.0fm\n", observer.Latitude, observer.Longitude, observer.Altitude)
	fmt.Printf("Time: %s\n\n", now.Format("2006-01-02 15:04:05 MST"))

	if visibleVerbose {
		displayVisibleSatellitesVerbose(visible[:displayCount])
	} else {
		displayVisibleSatellitesList(visible[:displayCount])
	}

	if visibleLimit > 0 && len(visible) > visibleLimit {
		fmt.Printf("\n... %d more visible satellites. Use --limit to show more.\n", len(visible)-visibleLimit)
	}
}

func displayVisibleSatellitesList(visible []*VisibleSatellite) {
	fmt.Printf("%-8s  %-40s  %-7s  %-7s  %-11s\n", "NORAD", "Name", "El (°)", "Az (°)", "Range (km)")
	fmt.Println(strings.Repeat("-", 80))

	for _, v := range visible {
		fmt.Printf("%-8d  %-40s  %7.2f  %7.2f  %11.0f\n",
			v.Satellite.NoradID,
			v.Satellite.Name,
			v.Angles.Elevation,
			v.Angles.Azimuth,
			v.Angles.Range)
	}
}

func displayVisibleSatellitesVerbose(visible []*VisibleSatellite) {
	for i, v := range visible {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("-", 60))
		}

		sat := v.Satellite
		fmt.Printf("Name:           %s\n", sat.Name)
		fmt.Printf("NORAD ID:       %d\n", sat.NoradID)
		fmt.Printf("Type:           %s\n", sat.ObjectType)
		fmt.Printf("Owner:          %s\n", sat.Owner)
		fmt.Printf("Orbit Regime:   %s\n", sat.OrbitRegime)

		fmt.Printf("\nCurrent Position:\n")
		fmt.Printf("  Elevation:    %.2f°\n", v.Angles.Elevation)
		fmt.Printf("  Azimuth:      %.2f°\n", v.Angles.Azimuth)
		fmt.Printf("  Range:        %.0f km\n", v.Angles.Range)
		fmt.Printf("  Range Rate:   %.2f km/s\n", v.Angles.RangeRate)

		fmt.Printf("\nOrbital Parameters:\n")
		fmt.Printf("  Period:       %.2f minutes\n", sat.Period)
		fmt.Printf("  Inclination:  %.2f°\n", sat.Inclination)
		fmt.Printf("  Apogee:       %.0f km\n", sat.Apogee)
		fmt.Printf("  Perigee:      %.0f km\n", sat.Perigee)
	}
}
