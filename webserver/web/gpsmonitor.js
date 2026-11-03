const version = "v1.0";

var updateDataTimer = null;
var debug = 0;
var updateInterval = 5000;
var data_request = null;
var gpsViews = new Array();

gpsView = new GpsView("gps_");
gpsViews.push(gpsView);

function doLoad()
{
	doLoadGpsViews();
	doStart();
	doPeriodic();
	
	if ((tmp = document.getElementById("version")))
		tmp.textContent = version;
}

function doLoadGpsViews()
{
	gpsViews.forEach(function(view)
	{
		view.doInit();
	});
}

function doUpdateGpsViews(sats)
{
	gpsViews.forEach(function(view)
	{
		view.doUpdate(sats);
	});
}

function doRestart()
{
	var restart_request = new XMLHttpRequest();
	restart_request.open("GET","restart", true);
	noCacheInRequest(restart_request);
	restart_request.responseType = "text";
	restart_request.onload = function()
	{
		document.body.style.cursor = "default";
		window.location.href = "restart.html"
	}
	document.body.style.cursor = "wait";
	restart_request.send();
}

function doCfg()
{
	window.location.href = "config.html";
}

function doStart()
{
	if (!updateDataTimer)
		updateDataTimer = setInterval(function(){doPeriodic()}, updateInterval);
}

function kB2Str(kbSize)
{
	if (kbSize < 1024)
		return kbSize + " kB";
	
	else if (kbSize < 1024*1024)
		return (kbSize/1024).toFixed(2) + " MB";
	
	else if (kbSize < 1024*1024*1024)
		return (kbSize/(1024*1024)).toFixed(2) + " GB";

	else if (kbSize < 1024*1024*1024*1024)
		return (kbSize/(1024*1024*1024)).toFixed(2) + " TB";
	else if (kbSize < 1024*1024*1024*1024*1024)
		return (kbSize/(1024*1024*1024*1024)).toFixed(2) + " PB";
	else if (kbSize < 1024*1024*1024*1024*1024*1024)
		return (kbSize/(1024*1024*1024*1024*1023)).toFixed(2) + " EB";
	else
		return "???";
}

function doPeriodic()
{
	if (data_request == null)
	{
		data_request = new XMLHttpRequest();
		data_request.open("GET","update", true);
		noCacheInRequest(data_request);
		data_request.responseType = "text";
		data_request.onerror = function()
		{
			log("Error in data request");
			data_request = null;
		}
		data_request.onabort = function()
		{
			log("Abort data request");
			data_request = null;
		}

		data_request.onload = function()
		{
			if (data_request.response)
			{
				var json = JSON.parse(data_request.response);
				// log(data_request.response);
				if (json.config.update_interval !== undefined)
				{
					var interval = parseInt(json.config.update_interval);
					if (interval != updateInterval)
					{
						updateInterval = interval;
						clearInterval(updateDataTimer);
						updateDataTimer = setInterval(function(){doPeriodic()}, updateInterval);
					}
				}

				setValue("sysstat_uptime", "sysstat_tr_uptime", json.status.uptime, null);
				if (json.status.raspi !== undefined)
				{
					const raspi = json.status.raspi
					setValue("sysstat_fan", "sysstat_tr_fan", raspi.fan, (val) =>  val + " rpm");
					setValue("sysstat_cpu", "sysstat_tr_cpu", raspi.cpu, (val) =>  (val/10.0).toFixed(1) + " %");
					setValue("sysstat_cpu_freq", "sysstat_tr_cpu_freq", raspi.cpu_freq, (val) =>  (val/1000.0).toFixed(2) + " GHz");
					setValue("sysstat_temperature", "sysstat_tr_temperature", raspi.temp, (val) =>  (val/10.0).toFixed(1) + " °C");
					setValue("sysstat_5volt", "sysstat_tr_5volt", raspi.voltage_5, (val) =>  (val/1000.0).toFixed(2) + " V");
					setValue("sysstat_3volt", "sysstat_tr_3volt", raspi.voltage_3, (val) =>  (val/1000.0).toFixed(2) + " V");
					setValue("sysstat_max_current", "sysstat_tr_max_current", raspi.max_current, (val) =>  (val/1000.0).toFixed(1) + " A");
					setValue("sysstat_serial", "sysstat_tr_serial", raspi.serial, null);
					setValue("sysstat_throttle", "sysstat_tr_throttle", raspi.throttle, (val) => val.toString(16).padStart(8, "0"));
					setValue("sysstat_version", "sysstat_tr_version", raspi.version, (val) => {
						const major = (raspi.version >> 8) & 0xFF;
						const minor = raspi.version & 0xFF;
						return major + "." + minor;
					});
					if (raspi.ram_u !== undefined && raspi.ram_t !== undefined)
						setValue("sysstat_ram", "sysstat_tr_ram", raspi.ram_u, (val) => kB2Str(val) + " / " + kB2Str(raspi.ram_t));
					else 
						hideElement("sysstat_tr_ram")
					if (raspi.wifi_strength !== undefined && raspi.wifi_rx !== undefined && raspi.wifi_tx !== undefined)
						setValue("sysstat_wifi", "sysstat_tr_wifi", raspi.wifi_strength, (val) => val + " dBm / " + kB2Str(raspi.wifi_rx + raspi.wifi_tx));
					else
						hideElement("sysstat_tr_wifi")
				}
				// GPS data
				setValue("summary_version", "summary_tr_version", json.status.version, null);
				const gps = json.data.gps
				if (gps !== undefined) {
					setValue("summary_latitude", "summary_tr_latitude", gps.la, (val) => val.toFixed(6));
					setValue("summary_longitude", "summary_tr_longitude", gps.lo, (val) => val.toFixed(6));
					setValue("summary_height", "summary_tr_height", gps.al, (val) => val.toFixed(1) + " m");
					setValue("summary_speed", "summary_tr_speed", gps.sp, (val) => val.toFixed(2) + " m/s");
					setValue("summary_course", "summary_tr_course", gps.co, (val) => {
						let v= ""
						 v = val.toFixed(2) + " °";
						 if (gps.mv !== undefined) {
							v += " / " + gps.mv.toFixed(2) + " °"
						 }
					})
					setValue("summary_mode", "summary_tr_mode", gps.mo, (val) => {
						let mode = ""
						switch (val) {
							case 0:
								mode = "No fix"
								break
							case 1:
								mode = "2D"
								break
							case 3:
								mode = "3D"
								break
							default :
								mode = "!!!"
								break
						}
						if (gps.qu !== undefined){
							mode += " / "
							switch (gps.qu) {
								case 0:
									mode += "No fix"
									break
								case 1:
									mode += "Normal fix"
									break
								case 2:
									mode += "Diff Fix"
									break
								case 3:
									mode += "PPS Fix"
									break
								case 4:
									mode += "RTK Fix"
									break
								case 5:
									mode += "RTK Float"
									break
								case 6:
									mode += "Estimated"
									break
								case 7:
									mode += "Manual"
									break
								case 8:
									mode += "Simulation"
									break
								default :
									mode += "!!!"
									break
							}
						}
						return mode
					});
					setValue("summary_nsats", "summary_tr_nsats", gps.su, (val) => val + " / " + (gps.st != null ? gps.st : "---"));
					setValue("summary_stdlatitude", "summary_tr_stdlatitude", gps.stla, (val) => val.toFixed(2) + " m");
					setValue("summary_stdlongitude", "summary_tr_stdlongitude", gps.stlo, (val) => val.toFixed(2) + " m");
					setValue("summary_stdaltitude", "summary_tr_altitude", gps.stal, (val) => val.toFixed(1) + " m");
					setValue("summary_date", "summary_tr_date", gps.dt, (val) => {
						const date = new Date(val);

						const formattedString = new Intl.DateTimeFormat('de-De', {
							weekday: 'short', // "Mo"
							day: 'numeric',   // "2" (ohne führende Null)
							month: 'numeric', // "11"
							year: 'numeric'   // "2026"
						}).format(date);
						return formattedString
					});
					setValue("summary_time", "summary_tr_time", gps.dt, (val) => {
						const date = new Date(val);

						const formattedString = new Intl.DateTimeFormat('de-De', {
							hour: '2-digit',   // "20"
							minute: '2-digit', // "03"
							second: '2-digit', // "12"
							hour12: false      // Erzwingt 24-Stunden-Format
						}).format(date);
						return formattedString
					});
				}
				doUpdateGpsViews(json.data.sats)
			} // response != NULL

			data_request = null;
		}
		data_request.send();
	}
}

function pad(num, size)
{
    var s = num + "";
    while (s.length < size)
		s = "0" + s;
    return s;
}

function doUpdate()
{
	var tmp = document.getElementById("update_file_upload");
	if (tmp)
	{
		var request = new XMLHttpRequest();
		request.open("POST", "firmwareupdate", true);
		noCacheInRequest(request);
		request.responseType = "text";
		request.onload = function()
		{
			document.body.style.cursor = "default";
			window.location.href = "update.html"
		}
			
		var data = new FormData();
		data.append("firmwarefile", tmp.files[0])
		document.body.style.cursor = "wait";
		request.send(data);
	}
	return false;
}

function noCacheInRequest(request)
{
	request.setRequestHeader("Cache-Control", "no-cache")
	request.setRequestHeader("Pragma", "no-cache")
}

function log(msg)
{
	if (debug && window.console)
		console.log(msg);
}

function hideElement(id) {
	const tmp = document.getElementById(id) 
	if (tmp)
		tmp.style.display = 'none';
}

function setValue(id, container, val, convert) {
	const tmp = document.getElementById(id) 
	if (!tmp) return
	if (val === undefined) {
		hideElement(container)
	} else {
		if (val == null) {
			tmp.textContent = "---";
		} else {
			if (convert && (typeof convert === 'function'))
				val = convert(val)
			tmp.textContent = val;
		}
	}
}

