package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStartMonitor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	path := "data/aircraft.json"
	dur := time.Second * 1
	maxAge := time.Second * 60

	store := Store{aircraft: make(map[string]AircraftPos), lock: new(sync.Mutex)}

	t.Run("success", func(t *testing.T) {
		err := startMonitor(ctx, path, dur, maxAge, &store, "dummy station")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("invalid store", func(t *testing.T) {
		err := startMonitor(ctx, path, dur, maxAge, nil, "dummy station")
		if err == nil {
			t.Error("expected an error, got none")
		}
	})

	t.Run("invalid file", func(t *testing.T) {
		err := startMonitor(ctx, "data/invalid.no.file", dur, maxAge, &store, "dummy station")
		if err != nil {
			t.Error(err)
		}
	})

}
