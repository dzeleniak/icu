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
	searchName    string
	searchOwner   string
	searchType    string
	searchLimit   int
	searchVerbose bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for satellites by name or other criteria",
	Long: `Search the satellite catalog using partial name matching and filters.
Returns a list of matching satellites with their NORAD IDs.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSearch()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&searchName, "name", "n", "", "Search by satellite name (partial match, case-insensitive)")
	searchCmd.Flags().StringVarP(&searchOwner, "owner", "o", "", "Filter by owner/country code")
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "Filter by object type (PAYLOAD, ROCKET BODY, DEBRIS)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 0, "Maximum number of results to display (0 = no limit)")
	searchCmd.Flags().BoolVarP(&searchVerbose, "verbose", "v", false, "Display verbose satellite information")
}

func runSearch() {
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

	// Search SATCAT entries
	results := searchSATCAT(catalog.SATCATs, searchName, searchOwner, searchType)

	if len(results) == 0 {
		fmt.Println("No satellites found matching the criteria.")
		return
	}

	// Limit results
	displayCount := len(results)
	if searchLimit > 0 && displayCount > searchLimit {
		displayCount = searchLimit
	}

	if searchVerbose {
		// Build TLE map for merging
		tleMap := buildTLEMap(catalog.TLEs)

		// Convert SATCAT results to Satellite objects with merged TLE data
		satellites := make([]*types.Satellite, displayCount)
		for i := 0; i < displayCount; i++ {
			sat := &types.SATCAT{
				ID:          results[i].ID,
				IntlID:      results[i].IntlID,
				Name:        results[i].Name,
				NoradID:     results[i].NoradID,
				LaunchDate:  results[i].LaunchDate,
				DecayDate:   results[i].DecayDate,
				ObjectType:  results[i].ObjectType,
				Owner:       results[i].Owner,
				LaunchSite:  results[i].LaunchSite,
				Period:      results[i].Period,
				Inclination: results[i].Inclination,
				Apogee:      results[i].Apogee,
				Perigee:     results[i].Perigee,
				RCSSize:     results[i].RCSSize,
			}

			satellites[i] = &types.Satellite{
				NoradID:     sat.NoradID,
				Name:        sat.Name,
				IntlID:      sat.IntlID,
				ObjectType:  sat.ObjectType,
				Owner:       sat.Owner,
				LaunchDate:  sat.LaunchDate,
				DecayDate:   sat.DecayDate,
				LaunchSite:  sat.LaunchSite,
				Period:      sat.Period,
				Inclination: sat.Inclination,
				Apogee:      sat.Apogee,
				Perigee:     sat.Perigee,
				RCSSize:     sat.RCSSize,
				SATCAT:      sat,
			}

			// Add TLE if available
			if tle, ok := tleMap[sat.NoradID]; ok {
				satellites[i].TLE = tle
			}
		}

		fmt.Printf("Found %d satellites", len(results))
		if searchLimit > 0 && len(results) > searchLimit {
			fmt.Printf(" (showing first %d)", searchLimit)
		}
		fmt.Println("\n")

		// Display verbose output
		displaySatellites(satellites)

		if searchLimit > 0 && len(results) > searchLimit {
			fmt.Printf("\n... %d more results. Use --limit to show more.\n", len(results)-searchLimit)
		}
	} else {
		// Display simple list
		fmt.Printf("Found %d satellites", len(results))
		if searchLimit > 0 && len(results) > searchLimit {
			fmt.Printf(" (showing first %d)", searchLimit)
		}
		fmt.Println("\n")

		for i := 0; i < displayCount; i++ {
			sat := results[i]
			fmt.Printf("%-8d  %s\n", sat.NoradID, sat.Name)
		}

		if searchLimit > 0 && len(results) > searchLimit {
			fmt.Printf("\n... %d more results. Use --limit to show more.\n", len(results)-searchLimit)
		}
	}
}

func searchSATCAT(satcats []types.SATCAT, name, owner, objType string) []types.SATCAT {
	results := make([]types.SATCAT, 0)

	nameLower := strings.ToLower(name)
	ownerUpper := strings.ToUpper(owner)
	typeLower := strings.ToLower(objType)

	for _, sat := range satcats {
		// Filter by name (partial match)
		if name != "" && !strings.Contains(strings.ToLower(sat.Name), nameLower) {
			continue
		}

		// Filter by owner (partial match)
		if owner != "" && !strings.Contains(strings.ToUpper(sat.Owner), ownerUpper) {
			continue
		}

		// Filter by type (partial match)
		if objType != "" && !strings.Contains(strings.ToLower(sat.ObjectType), typeLower) {
			continue
		}

		results = append(results, sat)
	}

	return results
}
