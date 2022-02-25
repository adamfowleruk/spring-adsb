package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	// GitRevision is set by the build process
	GitRevision string
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/go-adsb-console/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	if viper.IsSet("aircraftJSON") == false {
		log.Fatalln("Configuration file doesn't include a value for aircraftJSON.")
	}
	aircraftJSON := viper.GetString("aircraftJSON")

	if viper.IsSet("monitorDuration") == false {
		log.Fatalln("Configuration file doesn't include a value for monitorDuration.")
	}
	monitorDuration := viper.GetDuration("monitorDuration")

	if viper.IsSet("updateDuration") == false {
		log.Fatalln("Configuration file doesn't include a value for updateDuration.")
	}
	updateDuration := viper.GetDuration("updateDuration")

	if viper.IsSet("maxAircraftAge") == false {
		log.Fatalln("Configuration file doesn't include a value for maxAircraftAge.")
	}
	maxAircraftAge := viper.GetDuration("maxAircraftAge")

	if viper.IsSet("amqpURL") == false {
		log.Fatalln("Configuration file doesn't include a value for amqpURL.")
	}
	amqpURL := viper.GetString("amqpURL")

	if viper.IsSet("amqpExchange") == false {
		log.Fatalln("Configuration file doesn't include a value for amqpExchange.")
	}
	amqpExchange := viper.GetString("amqpExchange")

	if viper.IsSet("stationName") == false {
		log.Fatalln("Configuration file doesn't include a value for stationName.")
	}
	stationName := viper.GetString("stationName")

	// Handle OS signals gracefully
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	// Create an in-memory store to hold the latest aircraft positions
	var store = Store{
		aircraft: make(map[string]AircraftPos),
		lock:     new(sync.Mutex),
	}

	// Start monitoring for aircraft positions
	err = startMonitor(ctx, aircraftJSON, monitorDuration, maxAircraftAge, &store, stationName)
	if err != nil {
		log.Fatalln("failed to start monitor:", err)
	}

	// Start sending updates to RabbitMQ
	for n := 1; n <= 10; n++ {
		err = startUpdater(ctx, amqpURL, amqpExchange, updateDuration, stationName, &store)
		if err != nil {
			log.Printf("failed to start updater: attempt %d/%d: %s\n", n, 10, err)
			time.Sleep(time.Second * time.Duration(n))
			continue
		}
		break
	}
	if err != nil {
		log.Fatalln("failed to start updater:", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
