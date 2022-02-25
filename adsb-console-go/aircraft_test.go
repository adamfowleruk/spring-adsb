package main

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHasMoved(t *testing.T) {

	testCases := []struct {
		name    string
		a1      Aircraft
		a2      Aircraft
		wantVal bool
		wantErr error
	}{
		{
			name:    "identical",
			a1:      Aircraft{Flight: "a1", Lat: 1.1, Lon: 2.2, AltGeom: 3, Track: 4.1},
			a2:      Aircraft{Flight: "a1", Lat: 1.1, Lon: 2.2, AltGeom: 3, Track: 4.1},
			wantVal: false,
			wantErr: nil,
		},
		{
			name:    "moved_lat",
			a1:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			a2:      Aircraft{Flight: "a1", Lat: 2, Lon: 2, AltGeom: 3, Track: 4},
			wantVal: true,
			wantErr: nil,
		},
		{
			name:    "moved_lon",
			a1:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			a2:      Aircraft{Flight: "a1", Lat: 1, Lon: 3, AltGeom: 3, Track: 4},
			wantVal: true,
			wantErr: nil,
		},
		{
			name:    "moved_alt",
			a1:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			a2:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 4, Track: 4},
			wantVal: true,
			wantErr: nil,
		},
		{
			name:    "moved_track",
			a1:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			a2:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 3, Track: 5},
			wantVal: true,
			wantErr: nil,
		},
		{
			name:    "different",
			a1:      Aircraft{Flight: "a1", Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			a2:      Aircraft{Flight: "a2", Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			wantVal: false,
			wantErr: fmt.Errorf("a1 and a2 represent different aircraft"),
		},
		{
			name:    "unknown",
			a1:      Aircraft{Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			a2:      Aircraft{Lat: 1, Lon: 2, AltGeom: 3, Track: 4},
			wantVal: false,
			wantErr: fmt.Errorf("a1 and/or a2 represents unknown aircraft"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotVal, gotErr := HasMoved(tc.a1, tc.a2)

			if tc.wantErr != nil {
				if errors.Is(gotErr, tc.wantErr) {
					t.Errorf("%v != %v", gotErr, tc.wantErr)
				}
			} else {
				if gotErr != tc.wantErr {
					t.Errorf("%v != %v", gotErr, tc.wantErr)
				}
			}

			if gotVal != tc.wantVal {
				t.Errorf("%v != %v", gotVal, tc.wantVal)
			}
		})
	}
}

func TestUpdateAircraft(t *testing.T) {
	store := Store{aircraft: make(map[string]AircraftPos), lock: new(sync.Mutex)}

	var station = "dummy station"
	a1 := Aircraft{Flight: "A", Lat: 1, Lon: 2, AltGeom: 3, Track: 4, Seen: 90, Type: "AIRCRAFT", StationName: station, Timestamp: 1}
	a2 := Aircraft{Flight: "B", Lat: 1, Lon: 2, AltGeom: 3, Track: 4, Seen: 90, Type: "AIRCRAFT", StationName: station, Timestamp: 1}
	a3 := Aircraft{Flight: "C", Lat: 1, Lon: 2, AltGeom: 3, Track: 4, Seen: 90, Type: "AIRCRAFT", StationName: station, Timestamp: 1}
	a4 := Aircraft{Lat: 1, Lon: 2, AltGeom: 3, Track: 4, Seen: 60, Type: "AIRCRAFT", StationName: station, Timestamp: 1}

	// Data Store starts off with two known aircraft.
	store.aircraft[a1.Flight] = AircraftPos{aircraft: a1}
	store.aircraft[a2.Flight] = AircraftPos{aircraft: a2}

	// One aircraft moves position
	a1.Lat = -1

	// Scan contains four aircraft (one without a flight identifier)l
	scan := Scan{Now: 100.0, Aircraft: []Aircraft{a1, a2, a3, a4}}

	updateAircraft(scan, &store, station)

	// We expect the position of the known aircraft that moved to be updated.
	if store.aircraft[a1.Flight].aircraft != a1 {
		t.Errorf("%v != %v", store.aircraft[a1.Flight], a1)
	}

	// We expect the position of the aircraft that didn't move to remain unchanged
	if store.aircraft[a1.Flight].aircraft != a1 {
		t.Errorf("%v != %v", store.aircraft[a2.Flight], a2)
	}

	// We expect the data store to contain three aircraft the two it knew about and
	// the new aircraft that contained a flight identifier. We don't expect it to
	// contain the aircraft that didn't have a flight identifier.
	if got, want := len(store.aircraft), 3; got != want {
		t.Errorf("%d != %d", got, want)
	}
}

func TestPurgeAircraft(t *testing.T) {
	maxAge := time.Second * 60
	store := Store{aircraft: make(map[string]AircraftPos), lock: new(sync.Mutex)}

	// Data store contains two aircraft, one old, one new.
	a1 := Aircraft{Flight: "A", Seen: 10}
	a2 := Aircraft{Flight: "B", Seen: 90}
	store.aircraft[a1.Flight] = AircraftPos{aircraft: a1}
	store.aircraft[a2.Flight] = AircraftPos{aircraft: a2}

	// Scan contains no aircraft.
	scan := Scan{Aircraft: []Aircraft{a1, a2}}

	purgeAircraft(scan, &store, maxAge)

	// We expect the old aircraft to be removed from the store, but the new to remain.
	if got, want := len(store.aircraft), 1; got != want {
		t.Errorf("%d != %d", got, want)
	}
}
