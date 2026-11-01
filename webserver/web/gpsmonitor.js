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
	// doStart();
	// doPeriodic();
	
	if ((tmp = document.getElementById("version")))
		tmp.textContent = version;
}

function doLoadGpsViews()
{
	window.addEventListener("resize", function(event)
		{
			// doPeriodicGpsView();
		}
	);

	gpsViews.forEach(function(view)
	{
		view.doInit();
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

				if (json.status.time !== undefined)
				{
					var wday = ["So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"];
					var date_str = wday[json.status.time.wday] + ", " + (json.status.time.day < 10 ? "0" : "") + json.status.time.day + "." + (json.status.time.month < 10 ? "0" : "") + json.status.time.month + "." + json.status.time.year;
					var time_str = (json.status.time.hour < 10 ? "0" : "") + json.status.time.hour + ":" + (json.status.time.min < 10 ? "0" : "") + json.status.time.min + ":" + (json.status.time.sec < 10 ? "0" : "") + json.status.time.sec;
					if ((tmp = document.getElementById("status_date")))
						tmp.textContent = date_str;
					if ((tmp = document.getElementById("sysstat_date")))
						tmp.textContent = date_str;
					if ((tmp = document.getElementById("status_time")))
						tmp.textContent = time_str;
					if ((tmp = document.getElementById("sysstat_time")))
						tmp.textContent = time_str;
				}
				else
				{
					if ((tmp = document.getElementById("status_tr_date")))
						tmp.style.display = 'none';
					if ((tmp = document.getElementById("status_tr_time")))
						tmp.style.display = 'none';
				}
				
				if (json.status.version !== undefined)
				{
					if ((tmp = document.getElementById("summary_version")))
						tmp.textContent = json.status.version;
				}
				else
				{
					if ((tmp = document.getElementById("summary_tr_version")))
						tmp.style.display = 'none';
				}
				
				if (json.status.uptime !== undefined)
				{
					if ((tmp = document.getElementById("sysstat_uptime")))
						tmp.textContent = json.status.uptime;
				}
				else
				{
					if ((tmp = document.getElementById("sysstat_tr_uptime")))
						tmp.style.display = 'none';
				}
				
				if (tmp = document.getElementById("summary_temperature"))
					tmp.textContent = json.sensors[1].temperature.toFixed(1) + " (" + json.sensors[0].temperature.toFixed(1) + ") °C";
				if (tmp = document.getElementById("summary_temperature2"))
					tmp.textContent = json.sensors[1].temperature2.toFixed(1) + " (" + json.sensors[0].temperature2.toFixed(1) + ") °C";
				if (tmp = document.getElementById("summary_humidity"))
					tmp.textContent = json.sensors[1].humidity.toFixed(1) + " % rel";
				if (tmp = document.getElementById("summary_humidity2"))
					tmp.textContent = json.sensors[1].humidity2.toFixed(1) + " % rel";
				if (tmp = document.getElementById("summary_dewpoint"))
					tmp.textContent = json.sensors[1].dewpoint.toFixed(1) + " °C";
				if (tmp = document.getElementById("summary_dewpoint2"))
					tmp.textContent = json.sensors[1].dewpoint2.toFixed(1) + " °C";
				if (tmp = document.getElementById("summary_pressure"))
					tmp.textContent = json.sensors[0].pressure.toFixed(1) + " / " + json.sensors[0].pressure_normal.toFixed(1) + " hPa";
				if (tmp = document.getElementById("summary_pressure2"))
					tmp.textContent = json.sensors[0].pressure2.toFixed(1) + " / " + json.sensors[0].pressure2_normal.toFixed(1) + " hPa";
				if (tmp = document.getElementById("summary_pressure_trend"))
					tmp.textContent = json.status.pressure_trend_3.toFixed(2) + " / " + json.status.pressure_trend_1.toFixed(2) + " hPa/h";
				if (tmp = document.getElementById("summary_wetbulb"))
					tmp.textContent = json.status.wet_bulb.toFixed(1) + " °C";
				if (tmp = document.getElementById("summary_wetbulb2"))
					tmp.textContent = json.status.wet_bulb2.toFixed(1) + " °C";
				if (tmp = document.getElementById("summary_lux"))
					tmp.textContent = json.sensors[2].lux.toFixed(1) + " lx";
				if (tmp = document.getElementById("summary_lux2"))
					tmp.textContent = json.sensors[2].lux2.toFixed(1) + " lx";
				if (tmp = document.getElementById("summary_errors"))
					tmp.textContent = json.status.errors + " / " + json.status.resets;
				if (tmp = document.getElementById("summary_sensors"))
					tmp.textContent = json.sensors.map(sensor => sensor.name).join(", ").toLowerCase();
				
				if (json.status.raspi !== undefined)
				{
					if (tmp = document.getElementById("sysstat_fan"))
						tmp.textContent = json.status.raspi.fan + " rpm";
					if (tmp = document.getElementById("sysstat_cpu"))
						tmp.textContent = (json.status.raspi.cpu/10.0).toFixed(1) + " %";
					if (tmp = document.getElementById("sysstat_cpu_freq"))
						tmp.textContent = (json.status.raspi.cpu_freq/1000.0).toFixed(2) + " GHz";
					if (tmp = document.getElementById("sysstat_temperature"))
						tmp.textContent = (json.status.raspi.temp/10.0).toFixed(1) + " °C";
					if (tmp = document.getElementById("sysstat_5volt"))
						tmp.textContent = (json.status.raspi.voltage_5/1000.0).toFixed(2) + " V";
					if (tmp = document.getElementById("sysstat_3volt"))
						tmp.textContent = (json.status.raspi.voltage_3/1000.0).toFixed(2) + " V";
					if (tmp = document.getElementById("sysstat_max_current"))
						tmp.textContent = (json.status.raspi.max_current/1000.0).toFixed(1) + " A";
					if (tmp = document.getElementById("sysstat_serial"))
						tmp.textContent = json.status.raspi.serial;
					if (tmp = document.getElementById("sysstat_throttle"))
						tmp.textContent = "0x" + (json.status.raspi.throttle.toString(16).padStart(8, "0"));
					if (tmp = document.getElementById("sysstat_version"))
					{
						const major = (json.status.raspi.version >> 8) & 0xFF;
						const minor = json.status.raspi.version & 0xFF;
						tmp.textContent = major + "." + minor;
					}
					if (tmp = document.getElementById("sysstat_ram"))
					{
						tmp.textContent = kB2Str(json.status.raspi.ram_u) + " / " + kB2Str(json.status.raspi.ram_t);
					}
				}

				if (json.status.sd !== undefined)
				{
					if ((tmp = document.getElementById("sysstat_sd")))
						tmp.textContent = kB2Str(json.status.sd.free) + " / " + kB2Str(json.status.sd.total);
				}
				else
				{
					if ((tmp = document.getElementById("sysstat_tr_sd")))
						tmp.style.display = 'none';
				}
				
				if (json.status.raspi.wifi_strength !== undefined && json.status.raspi.wifi_rx !== undefined && json.status.raspi.wifi_tx !== undefined)
				{
					if ((tmp = document.getElementById("sysstat_wifi")))
						tmp.textContent = json.status.raspi.wifi_strength + " dBm / " + kB2Str(json.status.raspi.wifi_rx + json.status.raspi.wifi_tx);
				}
				else
				{
					if ((tmp = document.getElementById("sysstat_tr_wifi")))
						tmp.style.display = 'none';
				}

				/*
				if ((tmp = document.getElementById("summary_inttemp")))
					tmp.textContent = json.sensors[json.sensors.length - 1].temperature + " °C";
				*/
				
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
