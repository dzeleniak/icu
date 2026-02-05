package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dzeleniak/icu/internal/propagate"
	"github.com/dzeleniak/icu/internal/storage"
	"github.com/dzeleniak/icu/internal/types"
	"github.com/spf13/cobra"
)

var (
	noradID  int
	satName  string
	verbose  bool
	position bool
)

var getCmd = &cobra.Command{
	Use:   "get [NORAD_ID]",
	Short: "Get satellite information by NORAD ID or name",
	Long: `Retrieve and display satellite TLE, current position, and catalog information.
Provide a NORAD ID as a positional argument, or use --name to search by satellite name.
The default view shows TLE, current position (if observer is configured), and metadata.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runGet(args)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().IntVarP(&noradID, "norad", "n", 0, "Filter by NORAD catalog number")
	getCmd.Flags().StringVarP(&satName, "name", "m", "", "Filter by satellite name (case-insensitive, exact match)")
	getCmd.Flags().BoolVarP(&position, "position", "p", false, "Display TLE and current position")
	getCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose satellite information")
}

func runGet(args []string) {
	// Parse positional argument for NORAD ID if provided
	if len(args) > 0 && noradID == 0 && satName == "" {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid NORAD ID: %s", args[0])
		}
		noradID = id
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

	// Filter satellites
	filtered := filterSatellites(catalog.Satellites, noradID, satName)

	if len(filtered) == 0 {
		fmt.Println("No satellites found matching the criteria.")
		return
	}

	// Display results
	if verbose {
		displaySatellitesVerbose(filtered)
	} else if position {
		displaySatellitesWithPosition(filtered)
	} else {
		displaySatellitesDefault(filtered)
	}
}

// filterSatellites filters satellites based on NORAD ID and/or name
func filterSatellites(satellites []*types.Satellite, noradID int, name string) []*types.Satellite {
	if noradID == 0 && name == "" {
		return satellites
	}

	filtered := make([]*types.Satellite, 0)
	nameLower := strings.ToLower(name)

	for _, sat := range satellites {
		// Filter by NORAD ID if specified
		if noradID > 0 && sat.NoradID != noradID {
			continue
		}

		// Filter by name if specified (exact match, case-insensitive)
		if name != "" && strings.ToLower(sat.Name) != nameLower {
			continue
		}

		filtered = append(filtered, sat)
	}

	return filtered
}

// displaySatellitesDefault shows just the 3-line TLE format
func displaySatellitesDefault(satellites []*types.Satellite) {
	for _, sat := range satellites {
		if sat.TLE != nil {
			fmt.Printf("0 %s\n", sat.Name)
			fmt.Println(sat.TLE.Line1)
			fmt.Println(sat.TLE.Line2)
		}
	}
}

// displaySatellitesWithPosition shows TLE and current position
func displaySatellitesWithPosition(satellites []*types.Satellite) {
	// Check if observer is configured
	observerConfigured := config.ObserverLatitude != 0.0 || config.ObserverLongitude != 0.0
	var observer *propagate.ObserverPosition
	if observerConfigured {
		observer = &propagate.ObserverPosition{
			Latitude:  config.ObserverLatitude,
			Longitude: config.ObserverLongitude,
			Altitude:  config.ObserverAltitude,
		}
	}

	now := time.Now()

	for i, sat := range satellites {
		if i > 0 {
			fmt.Println()
		}

		// TLE at the top
		if sat.TLE != nil {
			fmt.Printf("0 %s\n", sat.Name)
			fmt.Println(sat.TLE.Line1)
			fmt.Println(sat.TLE.Line2)
			fmt.Println()
		}

		// Current position if observer is configured
		if observerConfigured && sat.TLE != nil {
			pos, err := propagate.PropagateSatellite(sat.TLE, now)
			if err == nil {
				angles := propagate.CalculateObservationAngles(pos, observer)
				fmt.Printf("Current Position (as of %s):\n", now.Format("2006-01-02 15:04:05 MST"))
				fmt.Printf("  Elevation:    %7.2f°\n", angles.Elevation)
				fmt.Printf("  Azimuth:      %7.2f°\n", angles.Azimuth)
				fmt.Printf("  Range:        %10.0f km\n", angles.Range)
				fmt.Printf("  Range Rate:   %8.2f km/s\n", angles.RangeRate)
			}
		} else if !observerConfigured {
			fmt.Println("Observer location not configured. Set observer_latitude, observer_longitude, and observer_altitude in config.")
		}
	}
}

// displaySatellitesVerbose shows TLE, current position, and all metadata
func displaySatellitesVerbose(satellites []*types.Satellite) {
	// Check if observer is configured
	observerConfigured := config.ObserverLatitude != 0.0 || config.ObserverLongitude != 0.0
	var observer *propagate.ObserverPosition
	if observerConfigured {
		observer = &propagate.ObserverPosition{
			Latitude:  config.ObserverLatitude,
			Longitude: config.ObserverLongitude,
			Altitude:  config.ObserverAltitude,
		}
	}

	now := time.Now()

	for i, sat := range satellites {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("=", 70))
			fmt.Println()
		}

		// TLE at the top
		if sat.TLE != nil {
			fmt.Printf("0 %s\n", sat.Name)
			fmt.Println(sat.TLE.Line1)
			fmt.Println(sat.TLE.Line2)
			fmt.Println()
		}

		// Current position if observer is configured
		if observerConfigured && sat.TLE != nil {
			pos, err := propagate.PropagateSatellite(sat.TLE, now)
			if err == nil {
				angles := propagate.CalculateObservationAngles(pos, observer)
				fmt.Printf("Current Position (as of %s):\n", now.Format("2006-01-02 15:04:05 MST"))
				fmt.Printf("  Elevation:    %7.2f°\n", angles.Elevation)
				fmt.Printf("  Azimuth:      %7.2f°\n", angles.Azimuth)
				fmt.Printf("  Range:        %10.0f km\n", angles.Range)
				fmt.Printf("  Range Rate:   %8.2f km/s\n", angles.RangeRate)
				fmt.Println()
			}
		}

		// Satellite metadata
		fmt.Printf("Name:           %s\n", sat.Name)
		fmt.Printf("NORAD ID:       %d\n", sat.NoradID)
		if sat.IntlID != "" {
			fmt.Printf("International:  %s\n", sat.IntlID)
		}
		if sat.ObjectType != "" {
			fmt.Printf("Type:           %s\n", sat.ObjectType)
		}
		if sat.Owner != "" {
			fmt.Printf("Owner:          %s\n", sat.Owner)
		}
		if sat.OrbitRegime != "" {
			fmt.Printf("Orbit Regime:   %s\n", sat.OrbitRegime)
		}
		if sat.LaunchDate != "" {
			fmt.Printf("Launch Date:    %s\n", sat.LaunchDate)
		}
		if sat.DecayDate != "" {
			fmt.Printf("Decay Date:     %s\n", sat.DecayDate)
		}
		if sat.LaunchSite != "" {
			fmt.Printf("Launch Site:    %s\n", sat.LaunchSite)
		}

		// Orbital parameters
		if sat.Period > 0 || sat.Inclination > 0 || sat.Apogee > 0 || sat.Perigee > 0 {
			fmt.Printf("\nOrbital Parameters:\n")
			if sat.Period > 0 {
				fmt.Printf("  Period:       %.2f minutes\n", sat.Period)
			}
			if sat.Inclination > 0 {
				fmt.Printf("  Inclination:  %.2f°\n", sat.Inclination)
			}
			if sat.Apogee > 0 {
				fmt.Printf("  Apogee:       %.0f km\n", sat.Apogee)
			}
			if sat.Perigee > 0 {
				fmt.Printf("  Perigee:      %.0f km\n", sat.Perigee)
			}
			if sat.RCSSize != "" {
				fmt.Printf("  RCS Size:     %s\n", sat.RCSSize)
			}
		}
	}
}
