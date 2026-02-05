package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/dzeleniak/icu/internal/storage"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display catalog statistics",
	Long: `Display statistics about the locally stored satellite catalog,
including the number of TLE and SATCAT entries and when the data
was last fetched. If no catalog exists, it will automatically fetch
the data (unless auto_fetch is disabled in config).`,
	Run: func(cmd *cobra.Command, args []string) {
		runStats()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats() {
	// Create storage
	store, err := storage.NewStorage(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Load catalog
	catalog, err := store.Load()
	if err != nil {
		log.Fatalf("Error loading catalog: %v", err)
	}

	// If no catalog exists and auto_fetch is enabled, fetch it
	if catalog == nil {
		if config.AutoFetch {
			fmt.Println("No catalog found. Fetching data...\n")
			runFetch()
			return
		} else {
			fmt.Println("No catalog found. Run 'icu fetch' to download data.")
			return
		}
	}

	// Check if catalog is stale and refresh if needed
	if config.IsCatalogStale(catalog) {
		age := time.Since(catalog.FetchedAt)
		maxAge := time.Duration(config.MaxCatalogAge) * time.Hour
		fmt.Printf("Catalog is stale (age: %v, max: %v). Refreshing...\n\n",
			age.Round(time.Minute), maxAge)
		runFetch()
		return
	}

	// Display statistics
	fmt.Println("Catalog Statistics")
	fmt.Println("==================")
	fmt.Printf("Satellites:      %d\n", len(catalog.Satellites))
	fmt.Printf("Last fetched:    %s\n", catalog.FetchedAt.Format("2006-01-02 15:04:05 MST"))

	// Show catalog age and staleness info
	age := time.Since(catalog.FetchedAt)
	fmt.Printf("Catalog age:     %v\n", age.Round(time.Minute))

	if config.MaxCatalogAge > 0 {
		maxAge := time.Duration(config.MaxCatalogAge) * time.Hour
		remaining := maxAge - age
		if remaining > 0 {
			fmt.Printf("Refresh in:      %v\n", remaining.Round(time.Minute))
		}
	}
}
