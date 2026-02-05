package types

import (
	"strconv"
	"strings"
	"time"
)

// TLE represents a Two-Line Element set entry (two lines of text)
type TLE struct {
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
}

// GetNoradID extracts the NORAD catalog number from the TLE
func (t *TLE) GetNoradID() int {
	// NORAD catalog number is in columns 3-7 of line 1 (after "1 ")
	if len(t.Line1) < 7 {
		return 0
	}

	// Extract the catalog number (typically "1 00005U" -> "00005")
	parts := strings.Fields(t.Line1)
	if len(parts) < 2 {
		return 0
	}

	// Remove trailing 'U' or 'C' classification
	numStr := strings.TrimRight(parts[1], "UC")
	noradID, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	return noradID
}

// SATCAT represents a Satellite Catalog entry
type SATCAT struct {
	ID          string  `json:"id"`
	IntlID      string  `json:"intlId"`
	Name        string  `json:"name"`
	NoradID     int     `json:"noradId"`
	LaunchDate  string  `json:"launchDate"`
	DecayDate   string  `json:"decayDate"`
	ObjectType  string  `json:"objectType"`
	Owner       string  `json:"owner"`
	LaunchSite  string  `json:"launchSite"`
	Period      float64 `json:"period"`
	Inclination float64 `json:"inclination"`
	Apogee      float64 `json:"apogee"`
	Perigee     float64 `json:"perigee"`
	RCSSize     string  `json:"rcsSize"`
}

// Catalog represents the stored satellite catalog data
type Catalog struct {
	Satellites []*Satellite `json:"satellites"`
	FetchedAt  time.Time    `json:"fetched_at"`
}

// Satellite represents a merged view of TLE and SATCAT data
type Satellite struct {
	NoradID     int     `json:"noradId"`
	Name        string  `json:"name"`
	IntlID      string  `json:"intlId"`
	ObjectType  string  `json:"objectType"`
	Owner       string  `json:"owner"`
	LaunchDate  string  `json:"launchDate"`
	DecayDate   string  `json:"decayDate"`
	LaunchSite  string  `json:"launchSite"`
	Period      float64 `json:"period"`
	Inclination float64 `json:"inclination"`
	Apogee      float64 `json:"apogee"`
	Perigee     float64 `json:"perigee"`
	RCSSize     string  `json:"rcsSize"`
	OrbitRegime string  `json:"orbitRegime"` // LEO, MEO, GEO, HEO, or UNKNOWN
	TLE         *TLE    `json:"tle"`
	SATCAT      *SATCAT `json:"satcat"`
}
