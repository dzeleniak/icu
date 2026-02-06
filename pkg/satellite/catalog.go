package satellite

import (
	"sort"
	"strings"
	"time"
)

// SearchCriteria represents multi-criteria search parameters for satellites.
type SearchCriteria struct {
	Name   string // partial match, case-insensitive
	Owner  string // partial match, case-insensitive
	Type   string // partial match, case-insensitive
	Regime string // exact match, case-insensitive
}

// VisibilityCriteria represents visibility search parameters.
type VisibilityCriteria struct {
	SearchCriteria              // Embed standard search criteria
	MinElevation   float64      // degrees
	MaxElevation   float64      // degrees
}

// VisibleSatellite represents a satellite with its current observation angles.
type VisibleSatellite struct {
	Satellite *Satellite
	Angles    *ObservationAngles
}

// MergeSatelliteData combines TLE and SATCAT data into Satellite objects.
// TLEs are used as the primary key, with SATCAT data merged when available.
// Satellites with missing orbital parameters have their orbit regime classified.
func MergeSatelliteData(tles []TLE, satcats []SATCAT) []*Satellite {
	// Create maps for efficient lookup
	tleMap := make(map[int]*TLE)
	for i := range tles {
		noradID := tles[i].GetNoradID()
		if noradID > 0 {
			tleMap[noradID] = &tles[i]
		}
	}

	satcatMap := make(map[int]*SATCAT)
	for i := range satcats {
		satcatMap[satcats[i].NoradID] = &satcats[i]
	}

	// Merge satellites using TLE as primary key
	satellites := make([]*Satellite, 0, len(tleMap))

	for noradID, tle := range tleMap {
		sat := &Satellite{
			NoradID: noradID,
			TLE:     tle,
		}

		// Merge SATCAT data if available
		if satcat, exists := satcatMap[noradID]; exists {
			sat.SATCAT = satcat
			sat.Name = satcat.Name
			sat.IntlID = satcat.IntlID
			sat.ObjectType = satcat.ObjectType
			sat.Owner = satcat.Owner
			sat.LaunchDate = satcat.LaunchDate
			sat.DecayDate = satcat.DecayDate
			sat.LaunchSite = satcat.LaunchSite
			sat.Period = satcat.Period
			sat.Inclination = satcat.Inclination
			sat.Apogee = satcat.Apogee
			sat.Perigee = satcat.Perigee
			sat.RCSSize = satcat.RCSSize

			// Determine orbit regime from orbital parameters
			sat.OrbitRegime = string(DetermineOrbitRegime(
				sat.Apogee,
				sat.Perigee,
				sat.Period,
				sat.Inclination,
			))
		} else {
			// TLE without SATCAT entry - use NORAD ID as name
			sat.Name = ""
			sat.OrbitRegime = "UNKNOWN"
		}

		satellites = append(satellites, sat)
	}

	// Sort by NORAD ID for consistent ordering
	sort.Slice(satellites, func(i, j int) bool {
		return satellites[i].NoradID < satellites[j].NoradID
	})

	return satellites
}

// FetchAndMergeCatalog fetches TLE and SATCAT data from the client and merges them into a Catalog.
// This is a convenience function that combines fetching and merging in a single operation.
func FetchAndMergeCatalog(client *Client) (*Catalog, error) {
	tles, err := client.FetchTLEs()
	if err != nil {
		return nil, err
	}

	satcats, err := client.FetchSATCATs()
	if err != nil {
		return nil, err
	}

	satellites := MergeSatelliteData(tles, satcats)

	return &Catalog{
		Satellites: satellites,
		FetchedAt:  time.Now(),
	}, nil
}

// FilterSatellites filters satellites by NORAD ID and/or name.
// If both noradID and name are zero/empty, returns all satellites.
// Name filtering is case-insensitive exact match.
func FilterSatellites(satellites []*Satellite, noradID int, name string) []*Satellite {
	if noradID == 0 && name == "" {
		return satellites
	}

	filtered := make([]*Satellite, 0)
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

// SearchSatellites performs multi-criteria search on satellites.
// All criteria are optional - empty strings are ignored.
// Name, owner, and type use partial matching (case-insensitive).
// Regime uses exact matching (case-insensitive).
// Results are sorted by NORAD ID.
func SearchSatellites(satellites []*Satellite, criteria SearchCriteria) []*Satellite {
	results := make([]*Satellite, 0)

	nameLower := strings.ToLower(criteria.Name)
	ownerUpper := strings.ToUpper(criteria.Owner)
	typeLower := strings.ToLower(criteria.Type)
	regimeUpper := strings.ToUpper(criteria.Regime)

	for _, sat := range satellites {
		// Filter by name (partial match)
		if criteria.Name != "" && !strings.Contains(strings.ToLower(sat.Name), nameLower) {
			continue
		}

		// Filter by owner (partial match)
		if criteria.Owner != "" && !strings.Contains(strings.ToUpper(sat.Owner), ownerUpper) {
			continue
		}

		// Filter by type (partial match)
		if criteria.Type != "" && !strings.Contains(strings.ToLower(sat.ObjectType), typeLower) {
			continue
		}

		// Filter by orbital regime (exact match)
		if criteria.Regime != "" && strings.ToUpper(sat.OrbitRegime) != regimeUpper {
			continue
		}

		results = append(results, sat)
	}

	// Sort results by NORAD ID
	sort.Slice(results, func(i, j int) bool {
		return results[i].NoradID < results[j].NoradID
	})

	return results
}

// FindVisibleSatellites finds satellites currently visible from the observer's location.
// Applies search criteria first, then filters by elevation bounds.
// Returns satellites with their observation angles, sorted by elevation (highest first).
func FindVisibleSatellites(
	satellites []*Satellite,
	observer *ObserverPosition,
	t time.Time,
	criteria VisibilityCriteria,
) ([]*VisibleSatellite, error) {
	// Apply search filters first
	candidates := SearchSatellites(satellites, criteria.SearchCriteria)

	visible := make([]*VisibleSatellite, 0)

	for _, sat := range candidates {
		if sat.TLE == nil {
			continue
		}

		pos, err := PropagateSatellite(sat.TLE, t)
		if err != nil {
			continue
		}

		angles := CalculateObservationAngles(pos, observer)

		if angles.Elevation >= criteria.MinElevation &&
			angles.Elevation <= criteria.MaxElevation {
			visible = append(visible, &VisibleSatellite{
				Satellite: sat,
				Angles:    angles,
			})
		}
	}

	// Sort by elevation (highest first)
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].Angles.Elevation > visible[j].Angles.Elevation
	})

	return visible, nil
}
