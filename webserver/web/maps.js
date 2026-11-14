/*!
 * Copyright (c) 2026 Reiner Pröls
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 * SPDX-License-Identifier: MIT
 *
 * Author: Reiner Pröls
 */

// Classes for (free) maps

class MapLibreMap {
    constructor(id) {
        this.map = null;
        this.marker = null;
		this.id = id;
        this.followPosition = true;
    }

    Load() {
        const css = document.createElement('link');
        css.rel = 'stylesheet';
        css.href = 'https://unpkg.com/maplibre-gl@5.6.1/dist/maplibre-gl.css';
        document.head.appendChild(css);

        const script = document.createElement('script');
        script.src = 'https://unpkg.com/maplibre-gl@5.6.1/dist/maplibre-gl.js';

        script.onload = () => {
            this.init();
        };

        document.head.appendChild(script);
    }

    init() {
        const div = document.getElementById(this.id);
        if (div) {
            div.style.display = "block";
        }
        this.map = new maplibregl.Map({
            container: this.id,
            // Liberty – farbiges Standard-Design
            // style: 'https://tiles.openfreemap.org/styles/liberty',

            // Bright – klar und kontrastreich
            style: 'https://tiles.openfreemap.org/styles/bright',

            // Positron – hellgrau und dezent
            // style: 'https://tiles.openfreemap.org/styles/positron',

            // Dark Matter – dunkles Design (CARTO)
            // style: 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json',            

            center: [0.0, 0.0],
            zoom: 16
        });

        this.map.addControl(new maplibregl.NavigationControl());

        this.map.on('load', () => {
            this.marker = new maplibregl.Marker()
                .setLngLat([0.0, 0.0])
                .addTo(this.map);
        });

        this.map.on('click', (e) => {
            openGoogleMaps(e.lngLat.lat, e.lngLat.lng);
        });

        document.addEventListener('mouseup', () => {
            this.followPosition = true;
        });

        this.map.on('mousedown', () => {
            // console.log('Mouse down');
            this.followPosition = false;
        });

        this.map.on('mouseup', () => {
            // console.log('Mouse up');
            this.followPosition = true;
        });

        this.map.on('dragstart', () => {
            this.followPosition = false;
            // console.log("dragstart");
        });

        this.map.on('dragend', () => {
            this.followPosition = true;
            // console.log("dragend");
        });        
    }

    SetPosition(lat, lng) {
        if (this.marker) 
            this.marker.setLngLat([lng, lat]);
        if (this.map && this.followPosition)
          this.map.panTo([lng, lat]);
        // this.map.jumpTo({ center: [lng, lat] });
    }
}

class OpenStreetMap {
    constructor(id){
        this.id = id;
        this.map = null;
        this.marker = null;
        this.followPosition = true;
        this.lat = 0
        this.lng = 0
    }
    Load(){
        const css = document.createElement('link');
        css.rel = 'stylesheet';
        css.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';

        css.onload = () => {
            const script = document.createElement('script');
            script.src = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js';

            script.onload = () => {
                this.init();
            };
            document.head.appendChild(script);
        };
        document.head.appendChild(css);
    }

    init(){
        const div = document.getElementById(this.id);
        if (div) {
            div.style.display = "block";
        }
        this.map = L.map(this.id).setView([0.0, 0.0], 16);
        this.map.attributionControl.setPrefix(false);

        // Kartenebenen
        const osm = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            // attribution: '&copy; OpenStreetMap contributors'
        });
        const topo = L.tileLayer('https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png', {
            // attribution: '&copy; OpenStreetMap contributors | OpenTopoMap'
        });
        const satellite = L.tileLayer(
            'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
            // attribution: 'Tiles &copy; Esri'
        });

        // CARTO Positron – hell und minimalistisch
        const positron = L.tileLayer(
            'https://{s}.basemaps.cartocdn.com/rastertiles/light_all/{z}/{x}/{y}{r}.png?key=cb1_52qc_1_5e889000f77b390288d794b3', {
            // attribution: '&copy; OpenStreetMap contributors &copy; CARTO'
        });

        // CARTO Dark Matter – dunkle Karte
        const dark = L.tileLayer(
            'https://{s}.basemaps.cartocdn.com/rastertiles/dark_all/{z}/{x}/{y}{r}.png?key=cb1_52qc_1_5e889000f77b390288d794b3', {
            // attribution: '&copy; OpenStreetMap contributors &copy; CARTO'
        });

        // CARTO Voyager – farbige Straßenkarte
        const voyager = L.tileLayer(
            'https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png?key=cb1_52qc_1_5e889000f77b390288d794b3', {
            // attribution: '&copy; OpenStreetMap contributors &copy; CARTO'
        });

        // OpenStreetMap HOT – Humanitarian
        const hot = L.tileLayer(
            'https://{s}.tile.openstreetmap.fr/hot/{z}/{x}/{y}.png', {
            // attribution: '&copy; OpenStreetMap contributors, HOT'
        });

        // CyclOSM – Radwege und Wanderwege
        const cyclosm = L.tileLayer(
            'https://{s}.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png', {
            // attribution: '&copy; OpenStreetMap contributors | CyclOSM'
        });

        const ttaceTrack = L.tileLayer('https://tile.tracestrack.com/_/{z}/{x}/{y}.webp?key=d9e7bba145b76f2ece9f8372b66f037e82e65139dfbb8e84', {
            maxNativeZoom: 19,
            tileSize: 512,
            zoomOffset: -1,
            // attribution: 'Map data &copy; <a href="https://openstreetmap.org">OpenStreetMap</a> contributors, Map render &copy; <a href="https://tracestrack.com">Tracestrack</a>'
        });
        
        // Standardkarte anzeigen
        osm.addTo(this.map);
        // Auswahlmenü
        L.control.layers({
            "OpenStreetMap": osm,
            "Topografisch": topo,
            "Satellit": satellite,
            "Positron (hell)": positron,
            "Dark Matter (dunkel)": dark,
            "Voyager": voyager,
            "HOT": hot,
            "CyclOSM (Radwege)": cyclosm,
            "Trace Track": ttaceTrack,
        }).addTo(this.map);

        this.marker = L.marker([0.0, 0.0]).addTo(this.map);

        this.map.on('click', (e) => {
            // openGoogleMaps(this.lat, this.lng);
            openOpenStreetMaps(this.lat, this.lng);
            // openOpenStreetMaps(e.latlng.lat, e.latlng.lng);
        });
        document.addEventListener('mouseup', () => {
            this.followPosition = true;
        });

        this.map.on('mousedown', () => {
            // console.log('Mouse down');
            this.followPosition = false;
        });

        this.map.on('mouseup', () => {
            // console.log('Mouse up');
            this.followPosition = true;
        });

        this.map.on('dragstart', () => {
            this.followPosition = false;
            // console.log("dragstart");
        });

        this.map.on('dragend', () => {
            this.followPosition = true;
            // console.log("dragend");
        });        
    }

    SetPosition(lat, lng) {
        if (this.followPosition)
        {
            if (this.marker)
                this.marker.setLatLng([lat, lng]);
            if (this.map && this.followPosition)
            this.map.panTo([lat, lng]);
        }
        this.lat = lat;
        this.lng = lng;
    }
}

class GoogleMap {
    constructor(id, minDist) {
		this.id = id;
        this.oldLong = 0.0;
        this.oldLat = 0.0;
        this.minDist = minDist;
        this.frame = null;
    }

    Load() {
        this.init()
    }

    init(){
        this.frame = document.getElementById(this.id);
        if (this.frame) {
            this.frame.style.display = "block";
        }
    }

    getDistanceInMeters(lat1, lon1, lat2, lon2) {
        const R = 6371000; // Erdradius in Metern
        const dLat = (lat2 - lat1) * Math.PI / 180;
        const dLon = (lon2 - lon1) * Math.PI / 180;
        
        const a = 
            Math.sin(dLat / 2) * Math.sin(dLat / 2) +
            Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * 
            Math.sin(dLon / 2) * Math.sin(dLon / 2);
            
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        return R * c; // Distanz in Metern
    }

    SetPosition(lat, lng) {
        if (this.frame) {
            if ((this.oldLat == 0.0 && this.oldLong == 0.0) ||
                this.getDistanceInMeters(this.oldLat, this.oldLong, lat, lng) >= this.minDist) {
                    this.frame.src = `https://www.google.com/maps?q=${lat},${lng}&z=16&output=embed`;
                    this.oldLat =lat;
                    this.oldLong = lng;
            }
        }
    }
}

function openGoogleMaps(lat, lng){
    window.open(
        `https://www.google.com/maps?q=${lat},${lng}`,
        '_blank'
    );
}

function openOpenStreetMaps(lat, lng){
    window.open(
        `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lng}&zoom=18&layers=M`,
        '_blank'
    );
}
