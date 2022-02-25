# go-adsb-console

## Installation

These instructions assume you are running Raspbian and have `dump1090-fa` installed.

1. Dowload the Debian package

```plain
wget https://github.com/billglover/go-adsb-console/releases/download/v0.10/go-adsb-console_0.10_linux_arm.deb
```

2. Install the Debian package

```plain
sudo dpkg -i go-adsb-console_0.10_linux_arm.deb
```

3. If this is your first installation, you will need to modify your configuration file: `/etc/go-adsb-console/config.yaml`

```yaml
---
aircraftJSON: /run/dump1090-fa/aircraft.json
monitorDuration: 1s
updateDuration: 5s
maxAircraftAge: 60s
amqpURL: "request-this-from-adam"
amqpExchange: "adsb-fan-exchange"
stationName: "unnamed-station"
```

You will need to update the value of `amqpURL` with a device key from Adam. Give you ground station a name by modifying the value of `stationName`.

4. If you have modified the configuration file, you will need to restart the application.

```plain
sudo systemctl restart go-adsb-console
```

## References

* dump1090 JSON field descriptions [pdf](http://www.nathanpralle.com/downloads/DUMP1090-FA_ADS-B_Aircraft.JSON_Field_Descriptions.pdf)
