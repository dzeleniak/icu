package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/dzeleniak/icu/pkg/satellite"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch TLE and SATCAT data from spacebook.com",
	Long: `Fetch retrieves the latest TLE (Two-Line Element) and SATCAT
(Satellite Catalog) data from spacebook.com and stores it locally
in ~/.icu/catalog.json for later use.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFetch()
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}

func runFetch() {
	// Create client with config values
	timeout := time.Duration(config.APITimeout) * time.Second
	apiClient := satellite.NewClient(config.TLEEndpoint, config.SATCATEndpoint, timeout)

	// Create storage
	store, err := satellite.NewStorage(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	fmt.Println("Fetching TLE data...")
	fmt.Println("Fetching SATCAT data...")
	fmt.Println("Merging satellite data...")

	// Use library function to fetch and merge catalog
	catalog, err := satellite.FetchAndMergeCatalog(apiClient)
	if err != nil {
		log.Fatalf("Error fetching catalog: %v", err)
	}

	if err := store.Save(catalog); err != nil {
		log.Fatalf("Error saving catalog: %v", err)
	}

	fmt.Println("\n✓ Data fetched successfully")
	fmt.Printf("  Merged satellites: %d\n", len(catalog.Satellites))
	fmt.Printf("\nCatalog saved to %s/catalog.json\n", config.DataDir)
}
