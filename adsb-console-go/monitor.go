package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// StartMonitor starts a new Go routine monitoring the provided file for
// changes changes in aircraft position. Any updates are reflected in the
// provided data Store. Aircraft in the data Store older than maxAge are
// removed from the store. An error is returned if the file is inaccessible
// at the point the monitor is started. Cancelling the provided context
// will terminate the Go routine.
func startMonitor(ctx context.Context, path string, dur, maxAge time.Duration, store *Store, station string) error {
	if store == nil {
		return errors.New("no data store provided")
	}

	ticker := time.NewTicker(dur).C

	go func() {
		lastModified := time.Now()

		for {
			select {
			case <-ticker:
				info, err := os.Stat(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to stat file: %v\n", err)
					continue
				}

				if info.ModTime().After(lastModified) {
					lastModified = info.ModTime()

					f, err := os.Open(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
					}

					dec := json.NewDecoder(f)
					scan := Scan{}
					err = dec.Decode(&scan)
					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to parse file: %v\n", err)
					}

					f.Close()

					updateAircraft(scan, store, station)
					purgeAircraft(scan, store, maxAge)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
