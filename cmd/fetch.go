package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/dzeleniak/icu/internal/client"
	"github.com/dzeleniak/icu/internal/storage"
	"github.com/dzeleniak/icu/internal/types"
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
	apiClient := client.NewClient(config.TLEEndpoint, config.SATCATEndpoint, timeout)

	// Create storage
	store, err := storage.NewStorage(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	fmt.Println("Fetching TLE data...")
	tles, err := apiClient.FetchTLEs()
	if err != nil {
		log.Fatalf("Error fetching TLEs: %v", err)
	}

	fmt.Println("Fetching SATCAT data...")
	satcats, err := apiClient.FetchSATCATs()
	if err != nil {
		log.Fatalf("Error fetching SATCATs: %v", err)
	}

	catalog := &types.Catalog{
		TLEs:      tles,
		SATCATs:   satcats,
		FetchedAt: time.Now(),
	}

	if err := store.Save(catalog); err != nil {
		log.Fatalf("Error saving catalog: %v", err)
	}

	fmt.Println("\n✓ Data fetched successfully")
	fmt.Printf("  TLE entities: %d\n", len(tles))
	fmt.Printf("  SATCAT entities: %d\n", len(satcats))
	fmt.Printf("  Total entities: %d\n", len(tles)+len(satcats))
	fmt.Printf("\nCatalog saved to %s/catalog.json\n", config.DataDir)
}
