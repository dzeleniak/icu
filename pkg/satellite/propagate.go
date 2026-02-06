package satellite

import (
	"fmt"
	"math"
	"time"

	"github.com/joshuaferrara/go-satellite"
)

// OrbitRegime represents the orbital regime classification
type OrbitRegime string

const (
	RegimeLEO     OrbitRegime = "LEO"     // Low Earth Orbit (< 2,000 km)
	RegimeMEO     OrbitRegime = "MEO"     // Medium Earth Orbit (2,000 - 35,786 km)
	RegimeGEO     OrbitRegime = "GEO"     // Geostationary Earth Orbit (~35,786 km, low inclination)
	RegimeHEO     OrbitRegime = "HEO"     // Highly Elliptical Orbit (high eccentricity)
	RegimeUnknown OrbitRegime = "UNKNOWN" // Unknown or insufficient data
)

// ObserverPosition represents the observer's location on Earth
type ObserverPosition struct {
	Latitude  float64 // degrees
	Longitude float64 // degrees
	Altitude  float64 // meters above sea level
}

// SatellitePosition represents a satellite's position at a specific time
type SatellitePosition struct {
	Time       time.Time
	X, Y, Z    float64 // ECEF coordinates in km
	Vx, Vy, Vz float64 // ECEF velocity in km/s
}

// ObservationAngles represents the satellite's position relative to the observer
type ObservationAngles struct {
	Time      time.Time
	Azimuth   float64 // degrees (0-360, 0=North, 90=East)
	Elevation float64 // degrees (-90 to 90)
	Range     float64 // kilometers
	RangeRate float64 // km/s
}

// PropagateSatellite propagates a satellite's position using SGP4.
// Returns the satellite's ECEF position at the given time.
func PropagateSatellite(tle *TLE, t time.Time) (*SatellitePosition, error) {
	if tle == nil {
		return nil, fmt.Errorf("TLE is nil")
	}

	// Parse the TLE using go-satellite library
	satrec := satellite.TLEToSat(tle.Line1, tle.Line2, "wgs72")

	// Get time components
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	// Propagate the satellite position
	position, velocity := satellite.Propagate(satrec, year, int(month), day, hour, min, sec)

	// Check for propagation errors
	if satrec.Error != 0 {
		return nil, fmt.Errorf("SGP4 propagation error: %d", satrec.Error)
	}

	return &SatellitePosition{
		Time: t,
		X:    position.X,
		Y:    position.Y,
		Z:    position.Z,
		Vx:   velocity.X,
		Vy:   velocity.Y,
		Vz:   velocity.Z,
	}, nil
}

// PropagateRange propagates a satellite over a time range with a given step size.
// Returns a slice of satellite positions.
func PropagateRange(tle *TLE, startTime, endTime time.Time, stepSize time.Duration) ([]*SatellitePosition, error) {
	if tle == nil {
		return nil, fmt.Errorf("TLE is nil")
	}

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	positions := make([]*SatellitePosition, 0)

	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(stepSize) {
		pos, err := PropagateSatellite(tle, t)
		if err != nil {
			return nil, fmt.Errorf("propagation failed at %v: %w", t, err)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

// ECEFToTopocentric converts ECEF coordinates to topocentric (ENU) coordinates
// relative to an observer's position
func ECEFToTopocentric(satPos *SatellitePosition, observer *ObserverPosition) (east, north, up float64) {
	// Convert observer geodetic coordinates to radians
	obsLatRad := observer.Latitude * math.Pi / 180.0
	obsLonRad := observer.Longitude * math.Pi / 180.0
	obsAltKm := observer.Altitude / 1000.0 // convert meters to km

	// For observer position in ECEF, use geodetic to ECEF conversion
	// Using WGS84 constants
	const (
		a  = 6378.137            // Earth semi-major axis in km
		f  = 1.0 / 298.257223563 // Earth flattening
		e2 = 2*f - f*f           // First eccentricity squared
	)

	sinLat := math.Sin(obsLatRad)
	cosLat := math.Cos(obsLatRad)
	sinLon := math.Sin(obsLonRad)
	cosLon := math.Cos(obsLonRad)

	N := a / math.Sqrt(1-e2*sinLat*sinLat)

	obsX := (N + obsAltKm) * cosLat * cosLon
	obsY := (N + obsAltKm) * cosLat * sinLon
	obsZ := (N*(1-e2) + obsAltKm) * sinLat

	// Calculate difference vector (satellite - observer) in ECEF
	dx := satPos.X - obsX
	dy := satPos.Y - obsY
	dz := satPos.Z - obsZ

	// Rotation matrix from ECEF to topocentric (ENU)
	east = -sinLon*dx + cosLon*dy
	north = -sinLat*cosLon*dx - sinLat*sinLon*dy + cosLat*dz
	up = cosLat*cosLon*dx + cosLat*sinLon*dy + sinLat*dz

	return east, north, up
}

// CalculateObservationAngles calculates azimuth, elevation, range, and range rate
// for a satellite position relative to an observer.
func CalculateObservationAngles(satPos *SatellitePosition, observer *ObserverPosition) *ObservationAngles {
	// Convert to topocentric coordinates
	east, north, up := ECEFToTopocentric(satPos, observer)

	// Calculate range (distance)
	rangeKm := math.Sqrt(east*east + north*north + up*up)

	// Calculate azimuth (0-360 degrees, 0=North, 90=East)
	azimuthRad := math.Atan2(east, north)
	azimuthDeg := azimuthRad * 180.0 / math.Pi
	if azimuthDeg < 0 {
		azimuthDeg += 360.0
	}

	// Calculate elevation (-90 to 90 degrees)
	elevationRad := math.Asin(up / rangeKm)
	elevationDeg := elevationRad * 180.0 / math.Pi

	// Calculate range rate (requires velocity)
	// Transform velocity to topocentric frame
	obsLatRad := observer.Latitude * math.Pi / 180.0
	obsLonRad := observer.Longitude * math.Pi / 180.0

	sinLat := math.Sin(obsLatRad)
	cosLat := math.Cos(obsLatRad)
	sinLon := math.Sin(obsLonRad)
	cosLon := math.Cos(obsLonRad)

	vEast := -sinLon*satPos.Vx + cosLon*satPos.Vy
	vNorth := -sinLat*cosLon*satPos.Vx - sinLat*sinLon*satPos.Vy + cosLat*satPos.Vz
	vUp := cosLat*cosLon*satPos.Vx + cosLat*sinLon*satPos.Vy + sinLat*satPos.Vz

	// Range rate is the dot product of velocity and range unit vector
	rangeRate := (east*vEast + north*vNorth + up*vUp) / rangeKm

	return &ObservationAngles{
		Time:      satPos.Time,
		Azimuth:   azimuthDeg,
		Elevation: elevationDeg,
		Range:     rangeKm,
		RangeRate: rangeRate,
	}
}

// CalculateObservationAnglesRange calculates observation angles over a time range.
func CalculateObservationAnglesRange(tle *TLE, observer *ObserverPosition, startTime, endTime time.Time, stepSize time.Duration) ([]*ObservationAngles, error) {
	positions, err := PropagateRange(tle, startTime, endTime, stepSize)
	if err != nil {
		return nil, err
	}

	observations := make([]*ObservationAngles, len(positions))
	for i, pos := range positions {
		observations[i] = CalculateObservationAngles(pos, observer)
	}

	return observations, nil
}

// IsVisible checks if a satellite is visible (above horizon) from the observer's position.
func IsVisible(obs *ObservationAngles, minElevation float64) bool {
	return obs.Elevation >= minElevation
}

// FindPasses finds visible passes of a satellite over a time range.
// A pass is defined as a continuous period where the satellite is above the minimum elevation.
func FindPasses(tle *TLE, observer *ObserverPosition, startTime, endTime time.Time, stepSize time.Duration, minElevation float64) ([][]*ObservationAngles, error) {
	observations, err := CalculateObservationAnglesRange(tle, observer, startTime, endTime, stepSize)
	if err != nil {
		return nil, err
	}

	passes := make([][]*ObservationAngles, 0)
	var currentPass []*ObservationAngles

	for _, obs := range observations {
		if IsVisible(obs, minElevation) {
			currentPass = append(currentPass, obs)
		} else {
			if len(currentPass) > 0 {
				passes = append(passes, currentPass)
				currentPass = nil
			}
		}
	}

	// Don't forget the last pass if it extends to the end
	if len(currentPass) > 0 {
		passes = append(passes, currentPass)
	}

	return passes, nil
}

// DetermineOrbitRegime classifies a satellite's orbital regime based on orbital parameters.
// Uses apogee, perigee (km), period (minutes), and inclination (degrees).
func DetermineOrbitRegime(apogee, perigee, period, inclination float64) OrbitRegime {
	// Check for invalid/missing data
	if apogee <= 0 || perigee <= 0 || period <= 0 {
		return RegimeUnknown
	}

	// Calculate semi-major axis (average altitude)
	const earthRadius = 6371.0 // km
	semiMajorAxis := ((apogee + earthRadius) + (perigee + earthRadius)) / 2.0
	avgAltitude := semiMajorAxis - earthRadius

	// Calculate eccentricity
	eccentricity := (apogee - perigee) / (apogee + perigee + 2*earthRadius)

	// HEO: Highly Elliptical Orbit (eccentricity > 0.25)
	if eccentricity > 0.25 {
		return RegimeHEO
	}

	// GEO: Geostationary orbit
	// Period ~1436 minutes (23.93 hours), altitude ~35,786 km, low inclination
	// Allow some tolerance for period and altitude
	periodTolerance := 30.0        // minutes
	altitudeTolerance := 500.0     // km
	inclinationTolerance := 5.0    // degrees

	geoAltitude := 35786.0
	geoPeriod := 1436.0

	if math.Abs(avgAltitude-geoAltitude) < altitudeTolerance &&
		math.Abs(period-geoPeriod) < periodTolerance &&
		math.Abs(inclination) < inclinationTolerance {
		return RegimeGEO
	}

	// LEO: Low Earth Orbit (< 2,000 km)
	if avgAltitude < 2000.0 {
		return RegimeLEO
	}

	// MEO: Medium Earth Orbit (2,000 - 35,786 km)
	if avgAltitude >= 2000.0 && avgAltitude < 35786.0 {
		return RegimeMEO
	}

	// GEO altitude range (for satellites that might be drifting or in GEO transfer)
	if avgAltitude >= 35786.0 {
		return RegimeGEO
	}

	return RegimeUnknown
}
