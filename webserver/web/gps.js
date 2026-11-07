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

// GPS satellite view (skyplot)

class GpsView {
	static get NUMBER_OF_NETWORKS() { return 7; }
	static get SVG_NAMESPACE() { return "http://www.w3.org/2000/svg"; }
	static get SVG_SCALEFACTOR() { return 1000.0; }
  	constructor(prefix) {
	this.prefix = prefix;
	this.satColors = new Map();
	this.showMap = new Map();
	this.checkBoxMap = new Map();
  }
 
	doInit() {
		let svg = document.getElementById(this.prefix + "svg");
		if (svg.style.vectorEffect === undefined) {
			alert("no vector effect");
		}
		svg.setAttribute("viewBox", `${-1.15*GpsView.SVG_SCALEFACTOR} ${-1.15*GpsView.SVG_SCALEFACTOR} ${2.3*GpsView.SVG_SCALEFACTOR} ${2.3*GpsView.SVG_SCALEFACTOR}`);
		let back = document.getElementById(this.prefix + "back");
		removeAllChilds(back);

		const savedText = localStorage.getItem('gps_sat_show');
		if (savedText !== null) {
			this.showMap = new Map(JSON.parse(savedText));
		}
		if (!this.showMap.has(-1)){
			this.showMap.set(-1, 1);
		}

		for (let i = 0; i < GpsView.NUMBER_OF_NETWORKS; i++) {
			this.satColors.set(i, getCustomCssProperty(`--gps-sat_network_${i}_color`));
			if (!this.showMap.has(i)) {
				this.showMap.set(i, true);
			}
		}

		this.defaultBlue = getDefaultBlue();

		this.tooltip = document.getElementById("sat-tooltip");
		this.satMinStrengthRadius = Number(getCustomCssProperty(`--gps-sat_strength_min_radius`))*GpsView.SVG_SCALEFACTOR;
		this.satMaxStrengthRadius = Number(getCustomCssProperty(`--gps-sat_strength_max_radius`))*GpsView.SVG_SCALEFACTOR;
		this.radiusDelta = this.satMaxStrengthRadius;
		this.satUsedColor = getCustomCssProperty(`--gps-sat_used_color`);
		this.satUsedStrokeWidth = getCustomCssProperty(`--gps-sat_used_stroke_width`);
		this.fontSize = Number(getCustomCssProperty('--gps-axis_text_font_size'))*GpsView.SVG_SCALEFACTOR;

		this.paintBackground();
		this.coordinateSystem();

		document.querySelectorAll('.gps_network_toggle').forEach(checkbox => {
			const netWorkId = Number(checkbox.getAttribute('data-network'));
			checkbox.checked = this.showMap.get(netWorkId) ? true : false;
			this.checkBoxMap.set(netWorkId, checkbox);
			checkbox.addEventListener('change', (event) => {
				const netWorkId = Number(event.target.getAttribute('data-network'));
				const g = document.getElementById(`svg_${this.prefix}network_${netWorkId}`);
				this.showMap.set(netWorkId, Boolean(event.target.checked));
				if (g) {
					 g.style.display = event.target.checked ? "" : "none";
				}
				localStorage.setItem('gps_sat_show', JSON.stringify(Array.from(this.showMap.entries())));
			});
			const itemContainer = checkbox.closest('.sat_checkbox-item');
			if (itemContainer) {
				itemContainer.style.backgroundColor = this.satColors.get(netWorkId);
			}
		});
		
		document.querySelectorAll('.gps_sat_used').forEach(selectbox => {
			selectbox.value = this.showMap.get(-1);
			selectbox.addEventListener('change', (event) => {
				const showUsed = Number(event.target.value);
				this.showMap.set(-1, showUsed);
				localStorage.setItem('gps_sat_show', JSON.stringify(Array.from(this.showMap.entries())));
				document.querySelectorAll('.svg_sat_used, .svg_sat_not_used').forEach(circle => {
					const isUsed = circle.classList.contains('svg_sat_used')
					if (isUsed) {
						circle.style.visibility = (showUsed == 1 || showUsed == 2) ? "visible" : "hidden";
					} else {
						circle.style.visibility = (showUsed == 1 || showUsed == 0) ? "visible" : "hidden";
					}
				});
			});
		});
		const save = document.getElementById('svg_save');
			save.addEventListener('click', (event) => {
				// 2. Wandle das SVG-Element in einen XML-Textstring um (Original bleibt unverändert)
				const serializer = new XMLSerializer();
				let svgString = serializer.serializeToString(svg);

				const festeBreite = "1000";
				const festeHoehe = "1000";
				const licTxt = '<!-- Copyright (c) 2026 Reiner Pröls | SPDX-License-Identifier: MIT -->\n';
				if (!svgString.match(/^<svg[^>]+xmlns="http:\/\/www\.w3\.org\/2000\/svg"/)) {
					// Namespace, Breite und Höhe ergänzen
					svgString = svgString.replace(/^<svg/, `<svg xmlns="http://www.w3.org/2000/svg" width="${festeBreite}" height="${festeHoehe}"`);
				} else {
					// Breite und Höhe ergänzen
					svgString = svgString.replace(/^<svg/, `<svg width="${festeBreite}" height="${festeHoehe}"`);
				}
				svgString = licTxt + svgString;

				// 4. Erzeuge ein Blob-Objekt (eine virtuelle Textdatei im Arbeitsspeicher des Browsers)
				const blob = new Blob([svgString], { type: "image/svg+xml;charset=utf-8" });
				const url = URL.createObjectURL(blob);

				// 5. Erstelle einen unsichtbaren Download-Link und löse ihn automatisch aus
				const downloadLink = document.createElement("a");
				downloadLink.href = url;

				const now = new Date();
				const year = now.getFullYear();
				const month = String(now.getMonth() + 1).padStart(2, '0'); // Monate sind 0-basiert
				const day = String(now.getDate()).padStart(2, '0');
				const hours = String(now.getHours()).padStart(2, '0');
				const minutes = String(now.getMinutes()).padStart(2, '0');
				const seconds = String(now.getSeconds()).padStart(2, '0');
				const formattedTimestamp = `${year}${month}${day}_${hours}${minutes}${seconds}`;

				downloadLink.download = `gpsmonitor_${formattedTimestamp}.svg`;
				document.body.appendChild(downloadLink);
				downloadLink.click();
				
				// 6. Aufräumen: Link löschen und Speicher freigeben
				document.body.removeChild(downloadLink);
				URL.revokeObjectURL(url);
			});
	}

	paintBackground() {
		const color = getCustomCssProperty('--gps-back_color');
		let back = document.getElementById(this.prefix + "back");
		let circle = document.createElementNS(GpsView.SVG_NAMESPACE, "circle");
		const baseRadius = 1.0 * GpsView.SVG_SCALEFACTOR + this.radiusDelta;
		circle.setAttribute("cx", 0);
		circle.setAttribute("cy", 0);
		circle.setAttribute("r", baseRadius);
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
		const baseRadius = 1.0 * GpsView.SVG_SCALEFACTOR;

		let lineX = document.createElementNS(GpsView.SVG_NAMESPACE, "line");
 		lineX.setAttribute("x1", `${-baseRadius - this.radiusDelta}`);
  		lineX.setAttribute("y1", "0.0");
  		lineX.setAttribute("x2", `${baseRadius + this.radiusDelta}`);
  		lineX.setAttribute("y2", "0.0");
  		lineX.setAttribute("stroke", color);
  		lineX.setAttribute("stroke-width", stroke_width);
		lineX.setAttribute("vector-effect", "non-scaling-stroke");
		fragment.appendChild(lineX);

		let lineY = document.createElementNS(GpsView.SVG_NAMESPACE, "line");
 		lineY.setAttribute("x1", "0.0");
  		lineY.setAttribute("y1", `${baseRadius + this.radiusDelta}`);
  		lineY.setAttribute("x2", "0.0");
  		lineY.setAttribute("y2", `${-baseRadius - this.radiusDelta}`);
  		lineY.setAttribute("stroke", color);
  		lineY.setAttribute("stroke-width", stroke_width);
		lineY.setAttribute("vector-effect", "non-scaling-stroke");
		fragment.appendChild(lineY);

		let x1 = baseRadius * Math.cos((Math.PI/180)*45);
		let y1 = baseRadius * Math.sin((Math.PI/180)*45);
		const color2 = getCustomCssProperty('--gps-help_axis_color');
		const stroke_width2 = getCustomCssProperty('--gps-help_axis_stroke_width');
		const stroke_dash = removeQuotes(getCustomCssProperty('--gps-help_axis_dash'));

		const line1 = document.createElementNS(GpsView.SVG_NAMESPACE, "line");
		line1.setAttribute("x1", x1);
  		line1.setAttribute("y1", -y1);
  		line1.setAttribute("x2", -x1);
  		line1.setAttribute("y2", y1);
  		line1.setAttribute("stroke", color2);
  		line1.setAttribute("stroke-width", stroke_width2);
		line1.setAttribute("vector-effect", "non-scaling-stroke");
		line1.setAttribute("stroke-dasharray", stroke_dash);
		fragment.appendChild(line1);

		const line2 = document.createElementNS(GpsView.SVG_NAMESPACE, "line");
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
		const circle1 = document.createElementNS(GpsView.SVG_NAMESPACE, "circle");
		const color_circle = getCustomCssProperty('--gps-help_circle_color');
		const stroke_width_circle = getCustomCssProperty('--gps-help_circle_stroke_width');
		const stroke_dash_circle = removeQuotes(getCustomCssProperty('--gps-help_circle_dash'))
		circle1.setAttribute("cx", 0);
		circle1.setAttribute("cy", 0);
		circle1.setAttribute("r", baseRadius*(2.0/3.0));
  		circle1.setAttribute("fill", "none");
  		circle1.setAttribute("stroke", color_circle);
  		circle1.setAttribute("stroke-width", stroke_width_circle);
		circle1.setAttribute("vector-effect", "non-scaling-stroke");
		circle1.setAttribute("stroke-dasharray", stroke_dash_circle);
		fragment.appendChild(circle1);
	    
		// 60°
		const circle2 = document.createElementNS(GpsView.SVG_NAMESPACE, "circle");
		circle2.setAttribute("cx", 0);
		circle2.setAttribute("cy", 0);
		circle2.setAttribute("r", baseRadius*(1.0/3.0));
  		circle2.setAttribute("fill", "none");
  		circle2.setAttribute("stroke", color_circle);
  		circle2.setAttribute("stroke-width", stroke_width_circle);
		circle2.setAttribute("vector-effect", "non-scaling-stroke");
		circle2.setAttribute("stroke-dasharray", stroke_dash_circle);
		fragment.appendChild(circle2);

		const labelE = document.createElementNS(GpsView.SVG_NAMESPACE, "text");
    	// 2. Ausrichtung für die Außenseite
    	labelE.setAttribute("text-anchor", "start");       // Linksbündig (Text fließt nach rechts außen weg)
    	labelE.setAttribute("dominant-baseline", "alphabetic"); // Vertikal perfekt mittig auf der Achse zentriert

		// 3. Styling & Inhalt
		let colorText = getCustomCssProperty('--gps-axis_text_color');
		if (colorText == "auto") {
			colorText = this.defaultBlue;
		}
		const fontFamily = getCustomCssProperty('--gps-axis_text_font_family');
		const fontWeight = getCustomCssProperty('--gps-axis_text_font_weight');
		labelE.setAttribute("font-size", this.fontSize); // Etwas größer, da es außen steht
		labelE.setAttribute("fill", colorText);
		labelE.setAttribute("font-family", fontFamily);
		labelE.setAttribute("font-weight", fontWeight);
		labelE.textContent = "E";
		const textSize = this.getTextSize(lines, "W", this.fontSize, fontFamily, fontWeight)
	    labelE.setAttribute("x", baseRadius + textSize.width / 3.0 + this.radiusDelta);
    	labelE.setAttribute("y", textSize.height * 0.3);
	    fragment.appendChild(labelE);

	    const labelW = document.createElementNS(GpsView.SVG_NAMESPACE, "text");
    	labelW.setAttribute("y", textSize.height/3.0);   // Genau auf der horizontalen Achsen-Linie
    	labelW.setAttribute("text-anchor", "end");
    	labelW.setAttribute("dominant-baseline", "alphabetic");
		labelW.setAttribute("font-size", this.fontSize);
		labelW.setAttribute("fill", colorText);
		labelW.setAttribute("font-family", fontFamily);
		labelW.setAttribute("font-weight", fontWeight);
		labelW.textContent = "W";
	    labelW.setAttribute("x", -baseRadius - textSize.width / 3.0 - this.radiusDelta);
	    fragment.appendChild(labelW);

	    const labelN = document.createElementNS(GpsView.SVG_NAMESPACE, "text");
    	labelN.setAttribute("x", "0.0");
    	labelN.setAttribute("text-anchor", "middle");
    	labelN.setAttribute("dominant-baseline", "alphabetic");
		labelN.setAttribute("font-size", this.fontSize);
		labelN.setAttribute("fill", colorText);
		labelN.setAttribute("font-family", fontFamily);
		labelN.setAttribute("font-weight", fontWeight);
		labelN.textContent = "N";
	    labelN.setAttribute("y", -baseRadius - textSize.height *0.3 - this.radiusDelta);
	    fragment.appendChild(labelN);

	    const labelS = document.createElementNS(GpsView.SVG_NAMESPACE, "text");
    	labelS.setAttribute("x", "0.0");
    	labelS.setAttribute("text-anchor", "middle");
    	labelS.setAttribute("dominant-baseline", "alphabetic");
		labelS.setAttribute("font-size", this.fontSize);
		labelS.setAttribute("fill", colorText);
		labelS.setAttribute("font-family", fontFamily);
		labelS.setAttribute("font-weight", fontWeight);
		labelS.textContent = "S";
	    labelS.setAttribute("y", baseRadius + textSize.height + this.radiusDelta);
	    fragment.appendChild(labelS);

		lines.append(fragment);
	}

	doUpdate(sats) {
		if (sats === undefined || sats == null) {
			return
		}
		let data = document.getElementById(this.prefix + "data");
		removeAllChilds(data);
		const networks = new Map();
		const fragment = document.createDocumentFragment();
		const minStrengthRange = 10
		const maxStrengthRange = 50
		const baseRadius = 1.0 * GpsView.SVG_SCALEFACTOR;
		let g
		const satNumberMap = new Map();
		for (let i = 0; i < GpsView.NUMBER_OF_NETWORKS; i++) {
		    g = document.createElementNS(GpsView.SVG_NAMESPACE, "g");
			g.id = `svg_${this.prefix}network_${i}`;
			g.style.display = this.showMap.get(i) ? "" : "none";
			networks.set(i, g);
			fragment.prepend(g);
			satNumberMap.set(i, 0);
		}
		const showUsed = this.showMap.get(-1)

		for (const sat of sats) {
			let networkId = Number(sat.nw);
			let networkName = this.getNetworkName(networkId);
			if (networkId >= GpsView.NUMBER_OF_NETWORKS) {
				networkId = 0;
			}
			if (networks.has(networkId)) {
				g = networks.get(networkId);
			} else {
			    g = document.createElementNS(GpsView.SVG_NAMESPACE, "g");
				g.id = `svg_${this.prefix}network_${i}`;
				networks.set(networkId, g);
				fragment.append(g);
				alert("XXX");
			}
			if (satNumberMap.has(networkId)) {
				satNumberMap.set(networkId, satNumberMap.get(networkId) + 1);
			}
			
			// add circle data to g
			let r = baseRadius*(1.0 - 1.0 * sat.el/90.0);
			let x = r*Math.sin((Math.PI/180.0)*sat.az);
			let y = -r*Math.cos((Math.PI/180.0)*sat.az);
			
			const circle = document.createElementNS(GpsView.SVG_NAMESPACE, "circle");
			circle.setAttribute("cx", x);
			circle.setAttribute("cy", y);
			circle.setAttribute("fill", this.satColors.get(networkId));
			let maxStrength = Number(0);
			for (const strength of Object.values(sat.st)) {
				if (strength != null) {
					const n = Number(strength);
					if (maxStrength == 0 && n > maxStrength) { 
						maxStrength = n;
					}
				}
			}
			if (maxStrength > maxStrengthRange) {
				maxStrength = maxStrengthRange;
			}
			if (maxStrength < minStrengthRange) {
				maxStrength = minStrengthRange;
			}
			const f = (maxStrength - minStrengthRange)/(maxStrengthRange - minStrengthRange);
			r = this.satMinStrengthRadius + f *(this.satMaxStrengthRadius - this.satMinStrengthRadius)
			circle.setAttribute("r", r);
			if (sat.us) {
				circle.setAttribute("stroke", this.satUsedColor);
				circle.setAttribute("stroke-width", this.satUsedStrokeWidth);
				circle.setAttribute("vector-effect", "non-scaling-stroke");
				circle.classList.add('svg_sat_used');
				circle.style.visibility = (showUsed == 1 || showUsed == 2) ? "visible" : "hidden";
			} else {
				circle.setAttribute("stroke", "none");
				circle.classList.add('svg_sat_not_used');
				circle.style.visibility = (showUsed == 1 || showUsed == 0) ? "visible" : "hidden";
			}

			// Tooltip - Logik
			const buildTooltipContent = () => {
				let tt = `<div class="sat_tooltip_grid">` +
				`<span class="sat_tooltip_grid_key">Network:</span><span>${networkName}</span>` + 
				`<span class="sat_tooltip_grid_key">Id:</span><span>${sat.id}</span>` +
				`<span class="sat_tooltip_grid_key">Used:</span><span>${sat.us ? "yes" : "no"}</span>`;

				let first = true;
				for (const [signal, strength] of Object.entries(sat.st)) {
					if (strength != null) {
						if (first) {
							first = false;
							// 'grid-column: 1 / -1' bricht in die nächste Zeile um.
							// 'padding-left: 16px' sorgt für die gewünschte Einrückung nach rechts.
							tt +=`<span class="sat_tooltip_grid_key">Strength:</span><span></span>` + 
							`<span class="sat_tooltip_grid_inner">`; 
						}
						tt += `<span class="sat_tooltip_grid_key_inner">${this.getSignalName(signal)}:</span>` +
							`<span>${strength} dB</span>`;
					}
				}
				if (!first) {
					tt += `</span>`;
				} else {
					tt +=`<span class="sat_tooltip_grid_key">Strength:</span><span>---</span>`;
				}
				tt +=`<span class="sat_tooltip_grid_key">Azimuth:</span><span>${sat.az} °</span>` +
				`<span class="sat_tooltip_grid_key">Elevation:</span><span>${sat.el} °</span>` +
				`</div>`;
				
				this.tooltip.innerHTML = tt;
			};

			// Wir speichern den Zustand, ob der Tooltip für diesen Kreis offen ist, direkt im Element
			circle.tooltipOpen = false;

			circle.addEventListener("pointerenter", (event) => {
				if (event.pointerType === "touch") return; 
				buildTooltipContent();
				this.tooltip.style.display = "block";
			});

			circle.addEventListener("pointermove", (event) => {
				if (event.pointerType === "touch") return; 
				this.tooltip.style.left = (event.clientX + 15) + "px";
				this.tooltip.style.top = (event.clientY + 15) + "px";
			});

			circle.addEventListener("pointerleave", (event) => {
				if (event.pointerType === "touch") return; 
				this.tooltip.style.display = "none";
			});

			circle.addEventListener("click", (event) => {
				event.stopPropagation(); // Wichtig: Verhindert das Weiterleiten an das Dokument
				
				if (event.pointerType === "touch") {
					if (circle.tooltipOpen) {
						this.tooltip.style.display = "none";
						circle.tooltipOpen = false;
					} else {
						buildTooltipContent();
						this.tooltip.style.left = (event.clientX - 40) + "px";
						this.tooltip.style.top = (event.clientY - 70) + "px";
						this.tooltip.style.display = "block";
						circle.tooltipOpen = true;
					}
				}
			});

			document.addEventListener("click", () => {
				this.tooltip.style.display = "none";
				circle.tooltipOpen = false;
			});

			g.append(circle);
		}
		data.append(fragment);
		for (const [netWorkId, value] of satNumberMap) {
			if (this.checkBoxMap.has(netWorkId)) {
				const ctrl = this.checkBoxMap.get(netWorkId)
				if (ctrl.nextSibling) {
            		ctrl.nextSibling.textContent = `${this.getNetworkName(netWorkId)} (${satNumberMap.get(netWorkId)})`;
				}
			}
		}
	}

	getSignalName(signalId) {
		switch (signalId) {
		case "UNKNOWN":
			return "---"
		case "GPS_L1CA":
			return "GPS L1 C/A"
		case "GPS_L2CL":
			return "GPS L2 CL"
		case "GPS_L2CM":
			return "GPS L2 CM"
		case "GPS_L5I":
			return "GPS L5 I"
		case "GPS_L5Q":
			return "GPS L5 Q"
		case "SBAS_L1CA":
			return "SBAS L1 C/A"
		case "GALILEO_E1":
			return "Galileo E1"
		case "GALILEO_E5A":
			return "Galileo E5 a"
		case "GALILEO_E5B":
			return "Galileo E5 b"
		case "BEIDOU_B1I":
			return "BeiDou B1 I"
		case "BEIDOU_B2I":
			return "BeiDou B2 I"
		case "BEIDOU_B2A": 
			return "BeiDou B2 A"
		case "QZSS_L1CA":
			return "QZSS L1 C/A"
		case "QZSS_L1S":
			return "QZSS L1 S"
		case "QZSS_L2CM":
			return "QZSS L2 CM"
		case "QZSS_L2CL":
			return "QZSS L2 CL"
		case "QZSS_L5I":
			return "QZSS L5 I"
		case "QZSS_L5Q":
			return "QZSS L5 Q"
		case "GLONASS_L1OF":
			return "GLONASS L1 OF"
		case "GLONASS_L2OF":
			return "GLONASS L2 OF"
		case "NAVIC_L5A":
			return "NaviC L5A"
		default:
			return "???"
		}
	}

	getNetworkName(networkId) {
		switch (networkId) {
		case 0:
			return "Other"
		case 1:
			return "GPS"
		case 2:
			return "GLONASS"
		case 3:
			return "BeiDou"
		case 4:
			return "Galileo"
		case 5:
			return "QZSS"
		case 6:
			return "NaviC"
		case 7:
			return "SBAS"
		case 8:
			return "IMES"
		default:
			return "???"
		}
	}

	getTextSize(root, text, fontSize, fontFamily, fontWeight) {
		const label = document.createElementNS(GpsView.SVG_NAMESPACE, "text");
		label.setAttribute("font-size", fontSize);
		label.setAttribute("font-family", fontFamily);
		label.setAttribute("font-weight", fontWeight);
		label.textContent = text;
		root.appendChild(label);
		const rect = label.getBBox()
		label.remove()
		return rect
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

function getDefaultBlue() {
	// 1. Ein temporäres, unsichtbares Element erstellen
	const tempDiv = document.createElement('div');

	// 2. Ihm die offizielle System-Akzentfarbe als Hintergrund geben
	tempDiv.style.backgroundColor = 'AccentColor'; 
	tempDiv.style.display = 'none';
	document.body.appendChild(tempDiv);

	// 3. Jetzt MUSS der Browser das "System-Blau" in echtes RGB umrechnen!
	const echerRgbWert = window.getComputedStyle(tempDiv).backgroundColor;

	// 4. Das Hilfselement wieder sauber löschen
	tempDiv.remove();

    // Filtert alle Zahlen heraus (gibt ein Array zurück, z.B. ["166", "208", "255"])
    const matches = echerRgbWert.match(/\d+/g);
    
    if (!matches || matches.length < 3) {
        return "#ffffff"; // Sicherer Fallback (Weiß), falls der String leer oder ungültig ist
    }

    // WICHTIG: Die Indizes, [1], [2] einzeln ansprechen!
    const r = Number(matches[0]).toString(16).padStart(2, '0');
    const g = Number(matches[1]).toString(16).padStart(2, '0');
    const b = Number(matches[2]).toString(16).padStart(2, '0');

    return `#${r}${g}${b}`;
}
