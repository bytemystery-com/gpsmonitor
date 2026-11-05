
# GPS Monitor

![alt text](/webserver/web/favicon.png "Logo")

GPS Monitor is a private project for displaying information received from a GPS receiver.

I run it on a Raspberry Pi 5 using a serial interface.

The project includes an NMEA parser, which has been tested with a u-blox NEO-M9N receiver.

The main page displays the most important GPS information. There is also a satellite view (skyplot), which is generated dynamically as an SVG using JavaScript. The SVG can also be exported.

Additionally, a small map displays the current position.

In `gpsmonitor.js`, you can configure which map provider to use:

- MapLibre
- OpenStreetMap
- Google Maps (embedded via iframe)

Since Google Maps does not offer a free API for this purpose, I use the simple iframe approach.

To prevent unnecessary flickering, the map is only updated when the position changes by more than 10 metres.  

Author: Reiner Pröls  
License: MIT  

## Photos

![alt text](/doc/fotos/m9n.jpg "M9N receiver")

## Screenshots

![alt text](/screenshots/main.jpg "Main page")
![alt text](/screenshots/sat.jpg "Sky view")

## Links

- [Readme](https://bytemystery-com.github.io/gpsmonitor/)
- [Repository](https://github.com/bytemystery-com/gpsmonitor/)

© Copyright Reiner Pröls, 2026-2027

