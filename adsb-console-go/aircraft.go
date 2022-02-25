package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// Aircraft holds all known details about a flight.
// Property definitions: https://github.com/SDRplay/dump1090/blob/master/README-json.md
// Property definitions: http://www.nathanpralle.com/downloads/DUMP1090-FA_ADS-B_Aircraft.JSON_Field_Descriptions.pdf
type Aircraft struct {
	Hex            string  `json:"hex,omitempty"`          // Hexidecimal 24-bit ICAO code for aircraft
	Flight         string  `json:"flight,omitempty"`       // Flight Number as filed
	AltBaro        int     `json:"alt_baro,omitempty"`     // Barometric altitude of the aircraft
	AltGeom        int     `json:"alt_geom,omitempty"`     // Geometric altitude of the aircraft
	Gs             float64 `json:"gs,omitempty"`           // Ground Speed in knots
	Ias            int     `json:"ias,omitempty"`          // Indicated Air Speed in knots
	Tas            int     `json:"tas,omitempty"`          // True Air Speed in knots
	Mach           float64 `json:"mach,omitempty"`         // Mach number
	Track          float64 `json:"track,omitempty"`        // True track angle in degrees
	TrackRate      float64 `json:"track_rate,omitempty"`   // Track angle rate in degrees/second
	Roll           float64 `json:"roll,omitempty"`         // Roll Angle in degrees
	MagHeading     float64 `json:"mag_heading,omitempty"`  // Magnetic Heading
	TrueHeading    float64 `json:"true_heading,omitempty"` // True Heading
	BaroRate       int     `json:"baro_rate,omitempty"`    // Barometric rate of change of altitude in feet/minute
	GeomRate       int     `json:"geom_rate,omitempty"`    // Geometric rate of change of altitude in feet/minute
	Squawk         string  `json:"squawk,omitempty"`       // Aircraft's assigned squawk code
	Emergency      string  `json:"emergency,omitempty"`    // Whether or not the captain or crew has indicated plane is in a state of emergency
	Category       string  `json:"category,omitempty"`     // Indicates what type of transmission equipment is on board: Class A1, A1S, A2, A3, B1S, or B1 equipment
	NavQnh         float64 `json:"nav_qnh,omitempty"`      // Related to QNH, which is a barometer corrected for ground altitude
	NavAltitudeMcp int     `json:"nav_altitude_mcp,omitempty"`
	NavHeading     float64 `json:"nav_heading,omitempty"`
	Lat            float64 `json:"lat,omitempty"` // Latitude of current position
	Lon            float64 `json:"lon,omitempty"` // Longitude of current position
	Nic            int     `json:"nic,omitempty"`
	Rc             int     `json:"rc,omitempty"`
	SeenPos        float64 `json:"seen_pos,omitempty"`
	Version        int     `json:"version,omitempty"`  // DO-260, DO-260(A), or DO-260(B), version 0, 1, 2 repectively
	NicBaro        int     `json:"nic_baro,omitempty"` // Navigation Integrity Category (NIC) specifies an integrity containment radius around an aircraft's reported position. Similar to NAC_P but for different versions. Ranges from 0 to 11 where 0 = Unknown and 11 = <7.5m.
	NacP           int     `json:"nac_p,omitempty"`    // Navigation Accuracy Category for Position (NACp) specifies the 95% accuracy range of a reported aircraft's reported position within a circle of particular radius around the actual horizontal position. Values 0 to 11 indicating (in order): Unknown, <10 NM, <4 NM, <2 NM, <1 NM, <0.5 NM, <0.3 NM, <0.1 NM, <0.05 NM, <30 meters, <10 m, <3 m
	NacV           int     `json:"nac_v,omitempty"`    // Navigation Accuracy Category for Velocity (NACv) specifies the accuracy of a reported aircraft's velocity. Values 0 to 4. 0 = Unknown or greater than 10 m/s, 1 = less than 10 m/s; 2 = less than 3 m/s; 3 = less than 1 m/s; 4 = less than 0.3 m/s
	Sil            int     `json:"sil,omitempty"`      // Source Integrity Level (SIL) indicates the probability of the reported horizontal position exceeding the containment radius defined by the NIC on a per sample or per hour basis, as defined in TSO–C166b and TSO–C154c.
	SilType        string  `json:"sil_type,omitempty"` // SIL measurement type
	Gva            int     `json:"gva,omitempty"`      // Geometric Vertical Accuracy (GVA); Accuracy of vertical geometric position; 0 = unknown or greater than 150 meters; 1 = less than or equal to 150 meters; 2 = less than or equal to 45 meters
	Sda            int     `json:"sda,omitempty"`      // System Design Assurance (SDA) indicates the probability of an aircraft malfunction causing false or misleading information to be transmitted, as defined in TSO–C166b and TSO–C154c.
	// Mlat           []interface{} `json:"mlat,omitempty"`              // An object (array) that defines what values in the message have been derived from MLAT vs. the antenna
	// Tisb           []interface{} `json:"tisb,omitempty"`              // Traffic Information Service-Broadcast (TIS-B); near as I can tell, this would be an array that would define which of these values were obtained through TIS-B (ADS-B IN), but I'm not positive.
	Messages    int     `json:"messages,omitempty"`          // total number of Mode S messages received from this aircraft
	Seen        float64 `json:"seen,omitempty"`              // how long ago (in seconds before "now") a message was last received from this aircraft
	Rssi        float64 `json:"rssi,omitempty"`              // recent average RSSI (signal power), in dbFS; this will always be negative.
	Timestamp   int64   `json:"timestamp,omitempty"`         // the timestamp ("now") when this record was created
	Type        string  `json:"type,omitempty"`              // set to 'AIRCRAFT'
	StationName string  `json:"groundStationName,omitempty"` // ground station name used to identify the receiver
}

// Scan holds flight details for all currently visible aircraft.
type Scan struct {
	Now      float64    `json:"now"`      // the time this file was generated, in seconds since Jan 1 1970 00:00:00 GMT (the Unix epoch)
	Messages int        `json:"messages"` // the total number of Mode S messages processed since scanning started
	Aircraft []Aircraft `json:"aircraft"` // a slice of Aircraft, one entry for each known aircraft
}

// AircraftPos is a record that maintains the last known position of an aircraft
type AircraftPos struct {
	modified bool
	aircraft Aircraft
}

// Store is an in memory map of aircraft
type Store struct {
	lock     *sync.Mutex
	aircraft map[string]AircraftPos
}

// HasMoved takes two Aircraft positions and returns a boolean to indicate
// whether the aircraft has moved. An error is returned if the positions
// provided relate to different aircraft.
func HasMoved(a1, a2 Aircraft) (bool, error) {
	if a1.Flight == "" || a2.Flight == "" {
		return false, errors.New("a1 and/or a2 represents unknown aircraft")
	}

	if a1.Flight != a2.Flight {
		return false, errors.New("a1 and a2 represent different aircraft")
	}

	return !(a1.Lon == a2.Lon &&
		a1.Lat == a2.Lat &&
		a1.AltGeom == a2.AltGeom &&
		a1.AltBaro == a2.AltBaro &&
		a1.Track == a2.Track), nil
}

// UpdateAircraft takes a Scan and updates the data Store with the latest
// aircraft positions. Aircraft positions older than maxAge are removed
// from the data Store. The data Store is marked as modified if changes
// are made.
func updateAircraft(s Scan, store *Store, station string) {

	// update aircraft positions in the data Store
	for i := range s.Aircraft {

		if s.Aircraft[i].Flight == "" || s.Aircraft[i].Lon == 0 || s.Aircraft[i].Lat == 0 {
			continue
		}

		// Update and clean the aircraft data
		s.Aircraft[i].Flight = strings.TrimSpace(s.Aircraft[i].Flight)
		s.Aircraft[i].Type = "AIRCRAFT"
		s.Aircraft[i].StationName = station
		if s.Aircraft[i].Timestamp == 0 {
			s.Aircraft[i].Timestamp = time.Now().UnixNano() / 1000
		}

		a2, ok := store.aircraft[s.Aircraft[i].Flight]
		moved, _ := HasMoved(s.Aircraft[i], a2.aircraft)
		if ok && !moved {
			continue
		}

		store.lock.Lock()
		store.aircraft[s.Aircraft[i].Flight] = AircraftPos{aircraft: s.Aircraft[i], modified: true}
		store.lock.Unlock()
	}
}

// PurgeAircraft removes any aircraft not present in the scan from the
// data Store. Any aircraft that are included in the scan but are older
// than maxAge are also removed.
func purgeAircraft(s Scan, store *Store, maxAge time.Duration) {
	seen := map[string]bool{}
	for _, a := range s.Aircraft {
		seen[a.Flight] = true
	}

	for k, v := range store.aircraft {

		if _, ok := seen[k]; ok != true {
			delete(store.aircraft, k)
			continue
		}

		lastSeen := time.Second * time.Duration(v.aircraft.Seen)
		if lastSeen > maxAge {
			delete(store.aircraft, k)
		}
	}
}
