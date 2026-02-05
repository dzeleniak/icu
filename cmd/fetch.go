package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/dzeleniak/icu/internal/client"
	"github.com/dzeleniak/icu/internal/propagate"
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

// mergeSatelliteData combines TLE and SATCAT data into Satellite objects
// Only includes satellites that have TLEs (SATCAT entries without TLEs are ignored)
func mergeSatelliteData(tles []types.TLE, satcats []types.SATCAT) []*types.Satellite {
	// Build TLE map by NORAD ID
	tleMap := make(map[int]*types.TLE)
	for i := range tles {
		noradID := tles[i].GetNoradID()
		if noradID > 0 {
			tleMap[noradID] = &tles[i]
		}
	}

	// Build SATCAT map by NORAD ID
	satcatMap := make(map[int]*types.SATCAT)
	for i := range satcats {
		satcatMap[satcats[i].NoradID] = &satcats[i]
	}

	// Merge data - start with TLEs (only include satellites with TLEs)
	satellites := make([]*types.Satellite, 0, len(tleMap))

	for noradID, tle := range tleMap {
		// Get corresponding SATCAT entry if it exists
		satcat, hasSatcat := satcatMap[noradID]

		var sat *types.Satellite

		if hasSatcat {
			// Calculate orbital regime
			regime := propagate.DetermineOrbitRegime(
				satcat.Apogee,
				satcat.Perigee,
				satcat.Period,
				satcat.Inclination,
			)

			sat = &types.Satellite{
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
				OrbitRegime: string(regime),
				TLE:         tle,
				SATCAT:      satcat,
			}
		} else {
			// TLE exists but no SATCAT entry - create minimal satellite record
			sat = &types.Satellite{
				NoradID:     noradID,
				Name:        "",
				OrbitRegime: string(propagate.RegimeUnknown),
				TLE:         tle,
				SATCAT:      nil,
			}
		}

		satellites = append(satellites, sat)
	}

	return satellites
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

	fmt.Println("Merging satellite data...")
	satellites := mergeSatelliteData(tles, satcats)

	catalog := &types.Catalog{
		Satellites: satellites,
		FetchedAt:  time.Now(),
	}

	if err := store.Save(catalog); err != nil {
		log.Fatalf("Error saving catalog: %v", err)
	}

	fmt.Println("\n✓ Data fetched successfully")
	fmt.Printf("  TLE entities: %d\n", len(tles))
	fmt.Printf("  SATCAT entities: %d\n", len(satcats))
	fmt.Printf("  Merged satellites (with TLEs): %d\n", len(satellites))
	fmt.Printf("\nCatalog saved to %s/catalog.json\n", config.DataDir)
}
