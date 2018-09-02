package main

var IndexHtml = `<!DOCTYPE html>
<html>

<head>
	<title>GoWeck - Alarm Clock for Raumserver</title>
	<meta charset="utf-8"  />
	<link rel="stylesheet" href="https://netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/timepicker/1.3.5/jquery.timepicker.css" />
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
	<style type="text/css">
  table.currentAlarms {
		border-collapse: collapse;
	}
	td.tblth {
		font-weight: bold;
		text-align: center;
    padding: 3px;
		border: 1px solid black;
	}
	td.tbltd {
		text-align: center;
		padding: 2px;
		border: 1px solid black;
	}
	input.center {
		display: block;
		margin : 0 auto;
	}
	</style>
	<script type="text/javascript">
	function getAlarms() {
		var t = document.getElementById('currentAlarms');
		t.innerHTML = "";
		var tr_head = t.insertRow();
		var captions = ['Enabled', 'Time', 'WeekDays', 'Weekend', 'Stream', 'Zone', 'Delete'];
		captions.forEach(function(h) {
			td = tr_head.insertCell();
			td.className = 'tblth';
			td.appendChild(document.createTextNode(h));
		});
		var docVal = ['enable', 'hourMinute', 'weekDays', 'weekEnds', 'radioChannel', 'zoneName'];
		$.getJSON("/alarms", function(data) {
			if (data !== "null") {
				JSON.parse(data).forEach(function(alarm) {
					var tr_alarm = t.insertRow();
					docVal.forEach(function(k){
						if (k === 'enable') {
							toggle = document.createElement('a');
							toggle.setAttribute('onclick', 'toggleAlarm("'+alarm['_id']+'",'+alarm[k]+')');
							toggle.setAttribute('href', '#');
							td = tr_alarm.insertCell();
							td.className = 'tbltd';
							td.appendChild(toggle);
							toggle.appendChild(document.createTextNode(alarm[k]));
						} else {
							td = tr_alarm.insertCell();
							td.className = 'tbltd';
							td.appendChild(document.createTextNode(alarm[k]));
						}
					});
					var td_delete = document.createElement("td");
					td_delete.className = "tbltd";
					td_delete.innerHTML = "<a href=\"#\"><i onclick=\"deleteAlarm('"+alarm['_id']+"')\" class=\"fa fa-trash\"></i></a>";
					tr_alarm.appendChild(td_delete);
				});
		  }
		});
	}
	function populateZoneSelect() {
		var t = document.getElementById('zoneSelect');
		t.innerHTML = "";
		$.getJSON("/zones", function(data) {
			if (data !== "null") {
				JSON.parse(data).forEach(function(zone) {
					var s = document.createElement("option");
					s.text = zone.name;
					s.value = zone.udn;
					t.add(s);
				});
		  }
		});
	}
	function toggleAlarm(id, curr) {
		var alarmUpdate = {
			enable: (curr ? false : true).toString(),
		};
		$.ajax({
			url: '/alarm/' + id,
			type: 'POST',
			data: JSON.stringify(alarmUpdate),
			contentType: 'application/json',
			statusCode: {
				200: function() {
				  getAlarms();
			  }
			}
		});
	}
	function createAlarm() {
		var newAlarm = {
			enable: document.getElementById('alarmEnabled').checked.toString(),
			hourMinute: document.getElementById('timepicker').value,
			weekDays: document.getElementById('weekDaysEnabled').checked.toString(),
			weekEnds: document.getElementById('weekEndsEnabled').checked.toString(),
			zoneUuid: document.getElementById('zoneSelect').value,
      radioChannel: document.getElementById('channel').value
		};
		$.ajax({
			url: '/alarm',
			type: 'POST',
			data: JSON.stringify(newAlarm),
			contentType: 'application/json',
			statusCode: {
				201: function() {
				  getAlarms();
			  }
			}
		});
	}
	function deleteAlarm(id) {
		$.ajax({
	    url: '/alarm/' + id,
	    type: 'DELETE',
	    success: function(data) {
				getAlarms();
			}
	  });
	}
	</script>
</head>

<body onload="getAlarms();populateZoneSelect()">
	<p class="lead text-center">
		Current alarms
		<table align="center" class="currentAlarms" id="currentAlarms">
    </table>
	</p>

	<p class="lead text-center">
		Create new alarm
		<table class="currentAlarms" align="center">
			<tr><td class="tblth">Enable</td><td class="tblth">Time</td><td class="tblth">WeekDays</td><td class="tblth">Weekend</td><td class="tblth">Stream</td><td class="tblth">Zone</td><td class="tblth">Add</td></tr>
			<tr>
	<td class="tbltd"><input type="checkbox" class="checkbox center" id="alarmEnabled" /></td>
	<td class="tbltd"><input id="timepicker" class="timepicker text-center" jt-timepicker time="model.time" time-string="model.timeString" default-time="model.options.defaultTime" time-format="model.options.timeFormat" start-time="model.options.startTime" min-time="model.options.minTime" max-time="model.options.maxTime" interval="model.options.interval" dynamic="model.options.dynamic" scrollbar="model.options.scrollbar" dropdown="model.options.dropdown" /></td>
  <td class="tbltd"><input type="checkbox" class="checkbox center" id="weekDaysEnabled" /></td>
	<td class="tbltd"><input type="checkbox" class="checkbox center" id="weekEndsEnabled" /></td>
	<td class="tbltd"><select name="channel" id="channel">
  <option value="StarFM">Playlist: Star FM</option>
  <option value="http://mp3channels.webradio.rockantenne.de/alternative">Stream: Rock Antenne</option>
  <option value="http://stream.berliner-rundfunk.de/brf/mp3-128/internetradio">Stream: BRF 91.4</option>
  <option value="radio888">Playlist: RadioBerlin 88.8</option>
	<option value="radioeins">Playlist: radioeins</option>
</select></td>
<td class="tbltd"><select name="zoneSelect" id="zoneSelect">
</select></td>
<td class="tbltd"><a onclick="createAlarm()" href="#"><i class="fa fa-plus-square"></i></a></td>
</tr>
</table>
	</p>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/1/jquery.min.js"></script>
	<script src="https://netdna.bootstrapcdn.com/bootstrap/3.1.1/js/bootstrap.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/timepicker/1.3.5/jquery.timepicker.min.js"></script>
  <script>
	$(document).ready(function(){
	$('input.timepicker').timepicker({
    timeFormat: 'HH:mm',
    interval: 15,
    minTime: '00',
    maxTime: '23',
    defaultTime: '07',
    startTime: '04',
    dynamic: false,
    dropdown: true,
    scrollbar: true
});
});
	</script>

</body>
</html>
`
