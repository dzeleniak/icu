package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/dzeleniak/icu/internal/storage"
	"github.com/dzeleniak/icu/internal/types"
	"github.com/spf13/cobra"
)

var (
	noradID int
	satName string
	verbose bool
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get satellite information by NORAD ID or name",
	Long: `Retrieve and display merged TLE and SATCAT information for satellites.
You can filter by NORAD ID or satellite name. The command will display
all matching satellites with their orbital parameters and catalog data.`,
	Run: func(cmd *cobra.Command, args []string) {
		runGet()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().IntVarP(&noradID, "norad", "n", 0, "Filter by NORAD catalog number")
	getCmd.Flags().StringVarP(&satName, "name", "m", "", "Filter by satellite name (case-insensitive, exact match)")
	getCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose satellite information")
}

func runGet() {
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

	// Build lookup maps
	tleMap := buildTLEMap(catalog.TLEs)
	satcatMap := buildSATCATMap(catalog.SATCATs)

	// Merge and filter satellites
	satellites := mergeSatellites(tleMap, satcatMap)
	filtered := filterSatellites(satellites, noradID, satName)

	if len(filtered) == 0 {
		fmt.Println("No satellites found matching the criteria.")
		return
	}

	// Display results
	if verbose {
		displaySatellites(filtered)
	} else {
		displayTLEWithName(filtered)
	}
}

// buildTLEMap creates a map of NORAD ID to TLE
func buildTLEMap(tles []types.TLE) map[int]*types.TLE {
	tleMap := make(map[int]*types.TLE)
	for i := range tles {
		noradID := tles[i].GetNoradID()
		if noradID > 0 {
			tleMap[noradID] = &tles[i]
		}
	}
	return tleMap
}

// buildSATCATMap creates a map of NORAD ID to SATCAT
func buildSATCATMap(satcats []types.SATCAT) map[int]*types.SATCAT {
	satcatMap := make(map[int]*types.SATCAT)
	for i := range satcats {
		satcatMap[satcats[i].NoradID] = &satcats[i]
	}
	return satcatMap
}

// mergeSatellites combines TLE and SATCAT data into Satellite objects
func mergeSatellites(tleMap map[int]*types.TLE, satcatMap map[int]*types.SATCAT) []*types.Satellite {
	satellites := make([]*types.Satellite, 0)

	// Start with SATCAT entries as they have the most metadata
	for noradID, satcat := range satcatMap {
		sat := &types.Satellite{
			NoradID:     noradID,
			Name:        satcat.Name,
			IntlID:      satcat.IntlID,
			ObjectType:  satcat.ObjectType,
			Owner:       satcat.Owner,
			LaunchDate:  satcat.LaunchDate,
			DecayDate:   satcat.DecayDate,
			LaunchSite:  satcat.LaunchSite,
			Period:      satcat.Period,
			Inclination: satcat.Inclination,
			Apogee:      satcat.Apogee,
			Perigee:     satcat.Perigee,
			RCSSize:     satcat.RCSSize,
			SATCAT:      satcat,
		}

		// Add TLE if available
		if tle, ok := tleMap[noradID]; ok {
			sat.TLE = tle
		}

		satellites = append(satellites, sat)
	}

	return satellites
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

// displayTLEWithName outputs TLE with line 0 (name) for satellites that have TLE data
func displayTLEWithName(satellites []*types.Satellite) {
	for _, sat := range satellites {
		if sat.TLE != nil {
			fmt.Printf("0 %s\n", sat.Name)
			fmt.Println(sat.TLE.Line1)
			fmt.Println(sat.TLE.Line2)
		}
	}
}

// displaySatellites formats and displays satellite information
func displaySatellites(satellites []*types.Satellite) {
	for i, sat := range satellites {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("-", 60))
		}

		fmt.Printf("Name:           %s\n", sat.Name)
		fmt.Printf("NORAD ID:       %d\n", sat.NoradID)
		fmt.Printf("International:  %s\n", sat.IntlID)
		fmt.Printf("Type:           %s\n", sat.ObjectType)
		fmt.Printf("Owner:          %s\n", sat.Owner)
		fmt.Printf("Launch Date:    %s\n", sat.LaunchDate)

		if sat.DecayDate != "" {
			fmt.Printf("Decay Date:     %s\n", sat.DecayDate)
		}

		fmt.Printf("Launch Site:    %s\n", sat.LaunchSite)
		fmt.Printf("\nOrbital Parameters:\n")
		fmt.Printf("  Period:       %.2f minutes\n", sat.Period)
		fmt.Printf("  Inclination:  %.2f°\n", sat.Inclination)
		fmt.Printf("  Apogee:       %.0f km\n", sat.Apogee)
		fmt.Printf("  Perigee:      %.0f km\n", sat.Perigee)
		fmt.Printf("  RCS Size:     %s\n", sat.RCSSize)

		if sat.TLE != nil {
			fmt.Printf("\nTLE:\n")
			fmt.Printf("  %s\n", sat.TLE.Line1)
			fmt.Printf("  %s\n", sat.TLE.Line2)
		} else {
			fmt.Printf("\nTLE:            Not available\n")
		}
	}
}
