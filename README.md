# ADSB Position Demo

This folder contains several items that together make a demo allowing you to track civilian aircraft and send that data
in to RabbitMQ for processing, collation, and later central display.

There are several software products within this code base:-

- mutability-dump1090 - not included - download from https://github.com/mutability/dump1090 - provides the low level C code for running a Software Defined Radio. This is designed to be ran on a Raspberry Pi, but this version works well on a Mac
- adsb-console - A standalone Java app, designed to be run on the receiver (Raspberry Pi or Laptop) that sends data to a remote RabbitMQ instance
- adsb-live-processor - .NET and Spring Boot apps that Listens to a RabbitMQ exchange for new aircraft positions, and updates a Live (60 second TTL) record in Redis per aircraft
- adsb-history-processor - Spring Boot app that listens to a RabbitMQ exchange for new aircraft positions, and updates a postgres database table
- aircraft-monitor - .NET and Spring Boot web apps that provides a rest endpoint (/data/aircraft.json) which pulls data from Redis on request, and a web front end (/gmap.html) to show collated aircraft position information

Note two of the apps above are provided with C#.NET and Java Spring Boot varieties to showcase the same patterns being ran with different application runtimes.

## How can I add my own data to the global view?

Every Pivot can run their own receiver and contribute to the global map! If you're interested in this, follow the below steps

First, contact Adam Fowler in the UK Modern Apps SE team <adamf@vmware.com> for a key so you can send data to RabbitMQ.

### Buy a receiver

The code you will use requires a specific chipset, so not all DVB-T receivers will work. A good one is the NeoElec R820T2 SDR NESDR Mini 2 which is what I use. This is an inexpensive item that can be found easily online, including from Amazon.

### Install receiver software

Now, install and compile dump1090:-

```sh
git clone https://github.com/mutability/dump1090
mv dump1090 mutability-dump1090
cd mutability-dump1090
# Following line is for Raspbian on the Raspberry Pi. Your system may vary
sudo apt-get install libusb-1.0-0-dev librtlsdr-dev maven
make
cd ..
```

If it fails, read the README in the project's folder for ideas, or get Adam Fowler on a zoom meeting.

Now prepare the data web server:-

```sh
sudo mkdir /run/dump1090
sudo chown pi:pi /run/dump1090
sudo apt-get install lighttpd
sudo cp lighttpd.conf /etc/lighttpd/
sudo /etc/init.d/lighttpd restart
```

Edit your receiver app to put the approximate (not too accurate!) location of your base station. This can be found in the --lon and --lat settings of ./runreceiver.sh

Now plugin your Software Defined Radio stick and run the receiver:-

```sh
./runreceiver.sh
```

To test view the internal (your base station only) json view: http://localhost:8081/data/aircraft.json

You should see aircraft information in the console, and in the map view

### Compile and run the Java RabbitMQ app

You now need to compile and run the adsb-console-java project.

```sh
cd adsb-console-java
mvn package
```

Now place the below in the file ./mutability-dump1090/runitmq-configured.sh

```sh
#!/bin/sh
export AMQP_URL=amqp://SOMEURL-PROVIDED-BY-ADAMFOWLER
./runitmq.sh
```

Edit runitmq.sh to reflect the name of your Base Station. E.g. "Fred's Base Station". Do not put your full name or address.

Now run it :-

```sh
./mutability-dump1090/runitmq-configured.sh
```

You should see messages about aircraft appear on the console. Now Ctrl+C to stop this.

### Make this all run on startup

Execute the following commands

```sh
sudo cp startup.sh /etc/init.d/dump1090
sudo chown root:root /etc/init.d/dump1090
sudo chmod gu+x /etc/init.d/dump1090
sudo ln -s /etc/init.d/dump1090 /etc/rc5.d/
sudo reboot
```

Upon restart you should be able to hit the JSON http endpoint again, as well as see all your data on the central map, below.

### View your aircraft online

Your aircraft, along with aircraft from all global receivers, will now be visible on the central map here: http://aircraft-monitor-central.cfapps.io/gmap.html

### Running live-processor and aircraft-monitor yourself

DON'T DO THIS!!!

If you run a local copy then you will 'steal' messages from the central exchange unless you change the name of your queue in the adsb-live-processor RabbitMQ template configuration.

Running this app without modification will mean the global view will not list all aircraft, breaking the demo!!! So please DO NOT DO THIS!

Still wanna do it? Well...

1. Create your own Small Redis cloud 'smallredis' and Cloud AMQP 'surveil-amqp' services
1. Reconfigure your local adsb-console app to use this RabbitMQ URL, not the one provided by Adam Fowler
1. Edit adsb-live-processor/src/resources/application.properties and change the 'group' setting to adsbposition.MYOWNQUEUENAME and save the file
1. Run mvn package
1. Run cf push for the app. This stands up three copies of the processor (just to prove it can - doesn't require three copies, one will suffice)
1. Edit aircraft-monitor-java/manifest.yml To specify the app name of your choice (If pushing to EMEA, don't use the default!!!)
1. Perform cf push on this app
1. Bind both services to live-processor, and the smallredis service to aircraft-monitor, and re-configure the services
1. You will now see your own data, you anti-social person you, not wanting to share. Tsk Tsk ;o)

## What is happening under the hood?

The basic outline is:-

1. Each aircraft transmits ADSB messages at varying rates and signal strengths on 1090 MHz. This uses the same transmission tech as Digital TV, allowing it to be receiver by cheap software defined radio TV sticks (ONLY IF their frequency range stretches to 1090 MHz)
2. A receiver hears the signal, and decodes the raw data, creating a full (including partial messages) output on the console, and exposing a JSON endpoint with the latest data on http://localhost:8081/data/aircraft.json
3. The adsb-console app reads this local URL once a second. It drops data that has already been sent, and drops partial information data (data without a flight code or longitude and latitude), and sends the remaining data to a RabbitMQ message exchange
4. The adsb-live-processor app instance each listen to the same queue attached to this exchange, ensuring each message is processed only once whilst maintaining HA, and updates the aircraft's live view (a Redis object with the key as the flight code, and value as a JSON object). Each object has a time to live of 60 seconds, so aircraft that land or drop out of range of all receivers will vanish out of Redis after 60 seconds. Hence the name 'live view'.
5. The aircraft monitor app has a Spring based rest endpoint (at /data/aircraft.json for ease of memory) that scans Redis for all aircraft keys, and produces a combined JSON output
6. The web page of the aircraft monitoring app (at /gmap.html again for each of memory), queries this JSON endpoint once a second, and displays the global view on a map for you

## Sub project status

|Processor|Java status|C# status|Notes|
|----|----|----|----|
adsb-console-java|Working, instance per base station (local)|N/A|Reads ADSB JSON and sends to RabbitMQ
adsb-live-processor|Working, multi instance|Working, multi instance|Takes live feed and sends to Redis
aircraft-monitor|Working|Working|Aircraft REST endpoint (and older map display web app)
aircraftui|N/A|N/A|ReactJS (staticfile buildpack) new demo web UI

## License and Copyright

All code and material is copyright VMware, Inc 2020-2021 all rights reserved, and licensed under the Apache 2 license unless
otherwise stated.

aircraft-monitor-java contains the web front end from mutability-dump1090 within its src/main/resources/static folder which is from the mutability version of the dump1090 application available at https://github.com/mutability/dump1090 and are licensed under the GPL V2 License. These files are copied verbatim and have not been modified, except the config.js file which is designed for user editable configuration.

## Support

This app is provided as-is, without support. If you're using the internal pivotal demo app, then please email adamf@vmware.com with all relevant information and I'll endeavour to give you pointers.
