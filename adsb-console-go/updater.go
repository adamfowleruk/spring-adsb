package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func startUpdater(ctx context.Context, conStr, exchange string, dur time.Duration, station string, store *Store) error {

	conn, err := amqp.Dial(conStr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	rmqCh, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	closures := conn.NotifyClose(make(chan *amqp.Error))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-closures:
				var err error
				rmqCh, err = conn.Channel()
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to open a channel: %s", err)
				}
			}
		}
	}()

	rmqCh.ExchangeDeclare(
		exchange, // name
		"fanout", // kind
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)

	ticker := time.NewTicker(dur)

	go func() {
		defer conn.Close()
		defer rmqCh.Close()
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:

				for _, v := range store.aircraft {
					if v.modified == false {
						continue
					}

					// use the old aircraft definition here
					a := aircraft{
						Flight:      v.aircraft.Flight,
						Lon:         v.aircraft.Lon,
						Lat:         v.aircraft.Lat,
						Track:       v.aircraft.Track,
						Speed:       v.aircraft.Tas, // use True Air Speed
						Hex:         v.aircraft.Hex,
						Squawk:      v.aircraft.Squawk,
						Seen:        v.aircraft.Seen,
						SeenPos:     v.aircraft.SeenPos,
						Messages:    v.aircraft.Messages,
						Category:    v.aircraft.Category,
						Timestamp:   v.aircraft.Timestamp,
						Altitude:    v.aircraft.AltGeom, // use the Geometric Altitude
						VertRate:    v.aircraft.GeomRate,
						Rssi:        v.aircraft.Rssi,
						Type:        v.aircraft.Type,
						StationName: v.aircraft.StationName,
					}

					body, err := json.Marshal(a)
					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to marshal Aircraft: %v\n", err)
					}

					msg := amqp.Publishing{
						DeliveryMode: amqp.Transient,
						Timestamp:    time.Now(),
						ContentType:  "application/json",
						Body:         body,
					}

					store.lock.Lock()
					err = rmqCh.Publish(exchange, "", false, false, msg)
					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to publish to exchange: %v\n", err)
					}
					v.modified = false
					store.lock.Unlock()
				}
			}
		}
	}()

	return nil
}

// Aircraft is an internal representation of the aircraft schema. It is used to preserve the
// structure of aircraft messages while clients switch to the FlightAware version of the JSON.
// A long term goal should look at creating an internal structure specificly for the information
// we use.
type aircraft struct {
	Flight      string  `json:"flight"`
	Lon         float64 `json:"lon"`
	Lat         float64 `json:"lat"`
	Track       float64 `json:"track"`
	Speed       int     `json:"speed,omitempty"`
	Hex         string  `json:"hex"`
	Squawk      string  `json:"squawk,omitempty"`
	Seen        float64 `json:"seen,omitempty"`
	SeenPos     float64 `json:"seen_pos,omitempty"`
	Messages    int     `json:"messages,omitempty"`
	Category    string  `json:"category,omitempty"`
	NUCP        int     `json:"nucp,omitempty"`
	Timestamp   int64   `json:"timestamp,omitempty"`
	Altitude    int     `json:"altitude"`
	VertRate    int     `json:"vert_rate,omitempty"`
	Rssi        float64 `json:"rssi,omitempty"`
	Type        string  `json:"type"`
	StationName string  `json:"groundStationName"`
}
