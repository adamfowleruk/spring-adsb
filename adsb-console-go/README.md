# go-adsb-console

## Installation

These instructions assume you are running Raspbian and have `dump1090-fa` installed.

1. Install `dump1090-fa` by issuing the following command. If you've installed PiAware, you probably already have this running.

```sh
sudo apt-get install dump1090-fa
```

2. Dowload the package for your Operating System from the latest releases area:-

https://github.com/adamfowleruk/spring-adsb/releases

3. (Optional) Compare the SHA256 checksum in the checksum.txt file against the files you downloaded

4. Install the package

- Raspbian / Debian on a Raspberry Pi 64-bit (E.g. RPi 3, 4, Zero 2):-
  ```sh
  sudo dpkg -i go-adsb-console_linux_arm64.deb
  ```
- Raspbian / Debian on a Raspberry Pi 32-bit (E.g. RPi 1, 2, Zero):-
  ```sh
  sudo dpkg -i go-adsb-console_linux_arm.deb
  ```
- Debian on Intel 64-bit linux:-
  ```sh
  sudo dpkg -i go-adsb-console_linux_amd64.deb
  ```
- RPM based distribution on Intel 64-bit linux:-
  ```sh
  sudo dpkg -i go-adsb-console_linux_amd64.rpm
  ```
- Newer M1 MacOS machines:-
  ```sh
  tar xzf go-adsb-console_VERSION_darwin_arm64.tar.gz
  ```
- Older Intel MacOS machines:-
  ```sh
  tar xzf go-adsb-console_VERSION_darwin_amd64.tar.gz
  ```
- RPMs are also available for arm and arm64

  Note: If you've had your Raspberry Pi for a while, you may still be on a 32bit Operating System. In this case, use the 32 bit deb file above instead. For details on Raspberry Pi OS versions and architectures, see this blog: https://www.raspberrypi.com/news/raspberry-pi-os-64-bit/

5. If this is your first installation, you will need to modify your configuration file: `/etc/go-adsb-console/config.yaml`

```yaml
---
aircraftJSON: /run/dump1090-fa/aircraft.json
monitorDuration: 1s
updateDuration: 5s
maxAircraftAge: 60s
amqpURL: "request-this-from-adam"
amqpExchange: "adsb-fan-exchange"
stationName: "Fred's station"
```

You will need to update the value of `amqpURL` with a device key from Adam (adamf at vmware dot com). Give you ground station a name by modifying the value of `stationName`. Please DO NOT alter the above duration and age flags.

6. If you have modified the configuration file, you will need to restart the application.

```sh
sudo systemctl restart go-adsb-console
```

7. Now validate that your aircraft are visible at the below URL:-

https://aircraft.shared.12factor.xyz/

## References

* dump1090 JSON field descriptions [pdf](http://www.nathanpralle.com/downloads/DUMP1090-FA_ADS-B_Aircraft.JSON_Field_Descriptions.pdf)
