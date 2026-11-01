class GpsView {
  constructor(prefix) {
	this.prefix = prefix
	this.svgNameSpace = "http://www.w3.org/2000/svg";
  }
 
	doInit() {
		let svg = document.getElementById(this.prefix + "svg");
		if (svg.style.vectorEffect === undefined) {
			alert("no vector effect")
		}
		// 2. Die neue viewBox als Text-String setzen
		// Format: "X-Start Y-Start Gesamtbreite Gesamthöhe"
		svg.setAttribute("viewBox", "-1.12 -1.12 2.24 2.24");
		let back = document.getElementById(this.prefix + "back");
		removeAllChilds(back);
		this.paintBackground()
		this.coordinateSystem()
	}

	paintBackground() {
		const color = getCustomCssProperty('--gps-back_color');
		let back = document.getElementById(this.prefix + "back");
		let circle = document.createElementNS(this.svgNameSpace, "circle");
		circle.setAttribute("cx", 0);
		circle.setAttribute("cy", 0);
		circle.setAttribute("r", 1.0);
		circle.setAttribute("fill", color);
		back.appendChild(circle);
	}

	coordinateSystem(){
		this.paintCoordLines();	
	}

	paintCoordLines() {
		const color = getCustomCssProperty('--gps-main_axis_color');
		const stroke_width = getCustomCssProperty('--gps-main_axis_stroke_width');
		let lines = document.getElementById(this.prefix + "coord_lines");
		const fragment = document.createDocumentFragment();

		let lineX = document.createElementNS(this.svgNameSpace, "line");
 		lineX.setAttribute("x1", "-1.0");
  		lineX.setAttribute("y1", "0.0");
  		lineX.setAttribute("x2", "1.0");
  		lineX.setAttribute("y2", "0.0");
  		lineX.setAttribute("stroke", color);
  		lineX.setAttribute("stroke-width", stroke_width);
		lineX.setAttribute("vector-effect", "non-scaling-stroke");
		fragment.appendChild(lineX);

		let lineY = document.createElementNS(this.svgNameSpace, "line");
 		lineY.setAttribute("x1", "0.0");
  		lineY.setAttribute("y1", "1.0");
  		lineY.setAttribute("x2", "0.0");
  		lineY.setAttribute("y2", "-1.0");
  		lineY.setAttribute("stroke", color);
  		lineY.setAttribute("stroke-width", stroke_width);
		lineY.setAttribute("vector-effect", "non-scaling-stroke");
		fragment.appendChild(lineY);

		let x1 = Math.cos((Math.PI/180)*45)
		let y1 = Math.sin((Math.PI/180)*45)
		const color2 = getCustomCssProperty('--gps-help_axis_color');
		const stroke_width2 = getCustomCssProperty('--gps-help_axis_stroke_width');
		const stroke_dash = removeQuotes(getCustomCssProperty('--gps-help_axis_dash'));

		const line1 = document.createElementNS(this.svgNameSpace, "line");
		line1.setAttribute("x1", x1);
  		line1.setAttribute("y1", -y1);
  		line1.setAttribute("x2", -x1);
  		line1.setAttribute("y2", y1);
  		line1.setAttribute("stroke", color2);
  		line1.setAttribute("stroke-width", stroke_width2);
		line1.setAttribute("vector-effect", "non-scaling-stroke");
		line1.setAttribute("stroke-dasharray", stroke_dash);
		fragment.appendChild(line1);

		const line2 = document.createElementNS(this.svgNameSpace, "line");
		line2.setAttribute("x1", -x1);
  		line2.setAttribute("y1", -y1);
  		line2.setAttribute("x2", x1);
  		line2.setAttribute("y2", y1);
  		line2.setAttribute("stroke", color2);
  		line2.setAttribute("stroke-width", stroke_width2);
		line2.setAttribute("vector-effect", "non-scaling-stroke");
		line2.setAttribute("stroke-dasharray", stroke_dash);
		fragment.appendChild(line2);

		// 30°
		const circle1 = document.createElementNS(this.svgNameSpace, "circle");
		const color_circle = getCustomCssProperty('--gps-help_circle_color');
		const stroke_width_circle = getCustomCssProperty('--gps-help_circle_stroke_width');
		const stroke_dash_circle = removeQuotes(getCustomCssProperty('--gps-help_circle_dash'))
		circle1.setAttribute("cx", 0);
		circle1.setAttribute("cy", 0);
		circle1.setAttribute("r", 2.0/3.0);
  		circle1.setAttribute("fill", "none");
  		circle1.setAttribute("stroke", color_circle);
  		circle1.setAttribute("stroke-width", stroke_width_circle);
		circle1.setAttribute("vector-effect", "non-scaling-stroke");
		circle1.setAttribute("stroke-dasharray", stroke_dash_circle);
		fragment.appendChild(circle1);
	    
		// 60°
		const circle2 = document.createElementNS(this.svgNameSpace, "circle");
		circle2.setAttribute("cx", 0);
		circle2.setAttribute("cy", 0);
		circle2.setAttribute("r", 1.0/3.0);
  		circle2.setAttribute("fill", "none");
  		circle2.setAttribute("stroke", color_circle);
  		circle2.setAttribute("stroke-width", stroke_width_circle);
		circle2.setAttribute("vector-effect", "non-scaling-stroke");
		circle2.setAttribute("stroke-dasharray", stroke_dash_circle);
		fragment.appendChild(circle2);

		const labelE = document.createElementNS(this.svgNameSpace, "text");
    	labelE.setAttribute("y", "0.0");   // Genau auf der horizontalen Achsen-Linie
    	// 2. Ausrichtung für die Außenseite
    	labelE.setAttribute("text-anchor", "start");       // Linksbündig (Text fließt nach rechts außen weg)
    	labelE.setAttribute("dominant-baseline", "central"); // Vertikal perfekt mittig auf der Achse zentriert

		// 3. Styling & Inhalt
		const colorText = getCustomCssProperty('--gps-axis_text_color');
		const fontSize = getCustomCssProperty('--gps-axis_text_font_size');
		const fontFamily = getCustomCssProperty('--gps-axis_text_font_family');
		const fontWeight = getCustomCssProperty('--gps-axis_text_font_weight');
		labelE.setAttribute("font-size", fontSize); // Etwas größer, da es außen steht
		labelE.setAttribute("fill", colorText);
		labelE.setAttribute("font-family", fontFamily);
		labelE.setAttribute("font-weight", fontWeight);
		labelE.textContent = "E";
		const textSize = this.getTextSize(lines, "W", fontSize, fontFamily, fontWeight)
	    labelE.setAttribute("x", 1.0 + textSize.width / 3.0);
	    fragment.appendChild(labelE);

	    const labelW = document.createElementNS(this.svgNameSpace, "text");
    	labelW.setAttribute("y", "0.0");   // Genau auf der horizontalen Achsen-Linie
    	labelW.setAttribute("text-anchor", "end");
    	labelW.setAttribute("dominant-baseline", "central");
		labelW.setAttribute("font-size", fontSize);
		labelW.setAttribute("fill", colorText);
		labelW.setAttribute("font-family", fontFamily);
		labelW.setAttribute("font-weight", fontWeight);
		labelW.textContent = "W";
	    labelW.setAttribute("x", -1.0 - textSize.width / 3.0);
	    fragment.appendChild(labelW);

	    const labelN = document.createElementNS(this.svgNameSpace, "text");
    	labelN.setAttribute("x", "0.0");
    	labelN.setAttribute("text-anchor", "middle");
    	labelN.setAttribute("dominant-baseline", "text-after-edge");
		labelN.setAttribute("font-size", fontSize);
		labelN.setAttribute("fill", colorText);
		labelN.setAttribute("font-family", fontFamily);
		labelN.setAttribute("font-weight", fontWeight);
		labelN.textContent = "N";
	    labelN.setAttribute("y", -1.0);
	    fragment.appendChild(labelN);

	    const labelS = document.createElementNS(this.svgNameSpace, "text");
    	labelS.setAttribute("x", "0.0");
    	labelS.setAttribute("text-anchor", "middle");
    	labelS.setAttribute("dominant-baseline", "text-before-edge");
		labelS.setAttribute("font-size", fontSize);
		labelS.setAttribute("fill", colorText);
		labelS.setAttribute("font-family", fontFamily);
		labelS.setAttribute("font-weight", fontWeight);
		labelS.textContent = "S";
	    labelS.setAttribute("y", 1.0);
	    fragment.appendChild(labelS);

		lines.append(fragment)
	}

	getTextSize(root, text, fontSize, fontFamily, fontWeight) {
		const label = document.createElementNS(this.svgNameSpace, "text");
		label.setAttribute("font-size", fontSize);
		label.setAttribute("font-family", fontFamily);
		label.setAttribute("font-weight", fontWeight);
		label.textContent = text;
		root.appendChild(label);
		const rect = label.getBBox()
		label.remove()
		return rect
	}

	hideLayer(layerElement) {
    	layerElement.setAttribute("visibility", "hidden");
	}

	showLayer(layerElement) {
    	layerElement.setAttribute("visibility", "visible");
	}

	toggleLayer(layerElement) {
    	const isHidden = layerElement.getAttribute("visibility") === "hidden";
    	layerElement.setAttribute("visibility", isHidden ? "visible" : "hidden");
	}
}

function removeAllChilds(node)
{
	let last;
	while (last = node.lastChild)
		node.removeChild(last);
}

function getCustomCssProperty(name) {
	return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

function removeQuotes(val) {
	return val ? val.replace(/"/g, "") : "";
}

/*
const satCircle = document.createElementNS(this.svgNameSpace, "circle");
satCircle.setAttribute("cx", x);
satCircle.setAttribute("cy", y);
satCircle.setAttribute("r", 5);

// Nativen Tooltip erstellen
const title = document.createElementNS(this.svgNameSpace, "title");
title.textContent = "GPS Satellit PRN 24\nSignal: Stark\nElevation: 45°";

// Wichtig: Den Title INS INNERE des Kreises hängen!
satCircle.appendChild(title); 
layer.appendChild(satCircle);


-----------------------------------------------
<div id="sat-tooltip" style="
    position: fixed; 
    display: none; 
    background: rgba(0,0,0,0.85); 
    color: white; 
    padding: 8px 12px; 
    border-radius: 4px; 
    font-family: sans-serif; 
    font-size: 12px; 
    pointer-events: none; // Verhindert, dass der Mauszeiger auf dem Tooltip flackert
    z-index: 9999;">
</div>

// Einmalig im Constructor oder deiner Init-Funktion holen:
this.tooltipEl = document.getElementById("sat-tooltip");

// In deiner Funktion, die den Satelliten erstellt:
addSatellite(satData) {
    const satCircle = document.createElementNS(this.svgNameSpace, "circle");
    satCircle.setAttribute("cx", satData.x);
    satCircle.setAttribute("cy", satData.y);
    satCircle.setAttribute("r", 6);
    satCircle.setAttribute("fill", "#00FF00");

    // 🚀 MAUS FÄHRT DRÜBER
    satCircle.addEventListener("pointerenter", (event) => {
        // Inhalt befüllen (Hier geht sogar HTML!)
        this.tooltipEl.innerHTML = `
            <strong>${satData.network.toUpperCase()} - PRN ${satData.prn}</strong><br>
            Elevation: ${satData.elevation}°<br>
            Azimut: ${satData.azimuth}°
        `;
        
        // Tooltip sichtbar machen
        this.tooltipEl.style.display = "block";
    });

    // 🚀 MAUS BEWEGT SICH (Tooltip folgt dem Zeiger)
    satCircle.addEventListener("pointermove", (event) => {
        // Positionierung anhand der echten Maus-Bildschirmpixel (+15px Offset, damit es nicht unter dem Zeiger klebt)
        this.tooltipEl.style.left = (event.clientX + 15) + "px";
        this.tooltipEl.style.top = (event.clientY + 15) + "px";
    });

    // 🚀 MAUS VERLÄSST DEN SATELLITEN
    satCircle.addEventListener("pointerleave", () => {
        this.tooltipEl.style.display = "none";
    });

    this.layers[satData.network].appendChild(satCircle);
}

*/