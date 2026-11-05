
class MapLibreMap {
    constructor(id) {
        this.map = null;
        this.marker = null;
		this.id = id
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
    }

    SetPosition(lat, lng) {
        if (this.marker) 
            this.marker.setLngLat([lng, lat]);
        if (this.map)
          this.map.panTo([lng, lat]);
        // this.map.jumpTo({ center: [lng, lat] });
    }
}

class OpenStreetMap {
    constructor(id){
        this.id = id;
        this.map = null;
        this.marker = null;
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

        // Standardkarte anzeigen
        osm.addTo(this.map);
        // Auswahlmenü
        L.control.layers({
            "OpenStreetMap": osm,
            "Topografisch": topo,
            "Satellit": satellite
        }).addTo(this.map);

        this.marker = L.marker([0.0, 0.0]).addTo(this.map);

        this.map.on('click', function(e) {
            openGoogleMaps(e.latlng.lat, e.latlng.lng);
        });
    }

    SetPosition(lat, lng) {
        if (this.marker)
            this.marker.setLatLng([lat, lng]);
        if (this.map)
          this.map.panTo([lat, lng]);
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
