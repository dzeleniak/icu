package types

import "time"

// TLE represents a Two-Line Element set entry (two lines of text)
type TLE struct {
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
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
	TLEs      []TLE     `json:"tles"`
	SATCATs   []SATCAT  `json:"satcats"`
	FetchedAt time.Time `json:"fetched_at"`
}
