package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dzeleniak/icu/internal/propagate"
	"github.com/dzeleniak/icu/internal/storage"
	"github.com/dzeleniak/icu/internal/types"
	"github.com/spf13/cobra"
)

var (
	noradID  int
	satName  string
	showTLE  bool
	showPos  bool
	showData bool
	verbose  bool
	follow   bool
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
	getCmd.Flags().BoolVarP(&showTLE, "tle", "t", false, "Display TLE")
	getCmd.Flags().BoolVarP(&showPos, "position", "p", false, "Display current position")
	getCmd.Flags().BoolVarP(&showData, "data", "d", false, "Display satellite metadata")
	getCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Display all information (TLE + position + metadata)")
	getCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Continuously update position every second")
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
	if follow {
		// Follow mode: continuously update position (shows TLE + position)
		displaySatellitesFollow(filtered)
	} else if verbose {
		// Verbose is shorthand for --tle --position --data
		displaySatellitesVerbose(filtered)
	} else {
		// Composable flags: show only what's requested
		// If no flags set, default to TLE
		if !showTLE && !showPos && !showData {
			showTLE = true
		}
		displaySatellitesComposed(filtered, showTLE, showPos, showData)
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

// displaySatellitesComposed shows only the requested components based on flags
func displaySatellitesComposed(satellites []*types.Satellite, showTLE, showPos, showData bool) {
	// Check if observer is configured for position display
	observerConfigured := config.ObserverLatitude != 0.0 || config.ObserverLongitude != 0.0
	var observer *propagate.ObserverPosition
	if showPos && observerConfigured {
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

		// Display TLE if requested
		if showTLE && sat.TLE != nil {
			fmt.Printf("0 %s\n", sat.Name)
			fmt.Println(sat.TLE.Line1)
			fmt.Println(sat.TLE.Line2)
			if showPos || showData {
				fmt.Println()
			}
		}

		// Display current position if requested
		if showPos {
			if !observerConfigured {
				fmt.Println("Observer location not configured. Set observer_latitude, observer_longitude, and observer_altitude in config.")
			} else if sat.TLE != nil {
				pos, err := propagate.PropagateSatellite(sat.TLE, now)
				if err == nil {
					angles := propagate.CalculateObservationAngles(pos, observer)
					fmt.Printf("Current Position (as of %s):\n", now.Format("2006-01-02 15:04:05 MST"))
					fmt.Printf("  Elevation:    %7.2f°\n", angles.Elevation)
					fmt.Printf("  Azimuth:      %7.2f°\n", angles.Azimuth)
					fmt.Printf("  Range:        %10.0f km\n", angles.Range)
					fmt.Printf("  Range Rate:   %8.2f km/s\n", angles.RangeRate)
					if showData {
						fmt.Println()
					}
				}
			}
		}

		// Display metadata if requested
		if showData {
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
}

// displaySatellitesFollow continuously updates position every second
func displaySatellitesFollow(satellites []*types.Satellite) {
	// Only support single satellite for follow mode
	if len(satellites) > 1 {
		fmt.Println("Follow mode only supports a single satellite. Please specify one satellite.")
		return
	}
	if len(satellites) == 0 {
		return
	}

	sat := satellites[0]
	if sat.TLE == nil {
		fmt.Println("No TLE data available for this satellite.")
		return
	}

	// Check if observer is configured
	observerConfigured := config.ObserverLatitude != 0.0 || config.ObserverLongitude != 0.0
	if !observerConfigured {
		fmt.Println("Observer location not configured. Set observer_latitude, observer_longitude, and observer_altitude in config.")
		return
	}

	observer := &propagate.ObserverPosition{
		Latitude:  config.ObserverLatitude,
		Longitude: config.ObserverLongitude,
		Altitude:  config.ObserverAltitude,
	}

	// Set up signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create ticker for 1-second updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Display TLE once at the top
	fmt.Printf("0 %s\n", sat.Name)
	fmt.Println(sat.TLE.Line1)
	fmt.Println(sat.TLE.Line2)
	fmt.Println()
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Initial display
	displayCurrentPosition(sat, observer)

	for {
		select {
		case <-ticker.C:
			// Move cursor up to overwrite previous position (6 lines)
			fmt.Print("\033[6A")
			displayCurrentPosition(sat, observer)

		case <-sigChan:
			fmt.Println("\nExiting follow mode...")
			return
		}
	}
}

// displayCurrentPosition shows the current position for a single satellite
func displayCurrentPosition(sat *types.Satellite, observer *propagate.ObserverPosition) {
	now := time.Now()
	pos, err := propagate.PropagateSatellite(sat.TLE, now)
	if err != nil {
		fmt.Printf("Error propagating satellite: %v\n", err)
		return
	}

	angles := propagate.CalculateObservationAngles(pos, observer)
	fmt.Printf("Current Position (as of %s):\r\n", now.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("  Elevation:    %7.2f°%s\r\n", angles.Elevation, strings.Repeat(" ", 20))
	fmt.Printf("  Azimuth:      %7.2f°%s\r\n", angles.Azimuth, strings.Repeat(" ", 20))
	fmt.Printf("  Range:        %10.0f km%s\r\n", angles.Range, strings.Repeat(" ", 20))
	fmt.Printf("  Range Rate:   %8.2f km/s%s\r\n", angles.RangeRate, strings.Repeat(" ", 20))
	fmt.Printf("%s\r\n", strings.Repeat(" ", 70))
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
