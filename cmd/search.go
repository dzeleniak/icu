package cmd

import (
	"fmt"
	"log"

	"github.com/dzeleniak/icu/pkg/satellite"
	"github.com/spf13/cobra"
)

var (
	searchName    string
	searchOwner   string
	searchType    string
	searchRegime  string
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
	searchCmd.Flags().StringVarP(&searchRegime, "regime", "r", "", "Filter by orbital regime (LEO, MEO, GEO, HEO)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 0, "Maximum number of results to display (0 = no limit)")
	searchCmd.Flags().BoolVarP(&searchVerbose, "verbose", "v", false, "Display verbose satellite information")
}

func runSearch() {
	// Load catalog
	store, err := satellite.NewStorage(config.DataDir)
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

	// Search satellites using library function
	results := satellite.SearchSatellites(catalog.Satellites, satellite.SearchCriteria{
		Name:   searchName,
		Owner:  searchOwner,
		Type:   searchType,
		Regime: searchRegime,
	})

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
		// Display verbose output
		fmt.Printf("Found %d satellites", len(results))
		if searchLimit > 0 && len(results) > searchLimit {
			fmt.Printf(" (showing first %d)", searchLimit)
		}
		fmt.Println("\n")

		displaySatellitesVerbose(results[:displayCount])

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

