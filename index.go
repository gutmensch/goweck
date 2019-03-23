package main

var indexHTML = `<!DOCTYPE html>
<html>

<head>
	<title>GoWeck - Alarm Clock for Raumserver</title>
	<meta charset="utf-8"  />
	<meta name="viewport" content="width=device-width,initial-scale=1" />
	<link rel="stylesheet" href="https://netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/timepicker/1.3.5/jquery.timepicker.css" />
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/modernizr/2.8.3/modernizr.min.js"></script>

	<style type="text/css">
	  @import url('https://fonts.googleapis.com/css?family=Roboto+Condensed');
		/* Variables
		================================== */
		/* Tables
		================================== */
		.Rtable {
		  display: flex;
		  flex-wrap: wrap;
		  margin: 0 0 3em 0;
		  padding: 0;
		}
		.Rtable-cell {
		  box-sizing: border-box;
		  flex-grow: 1;
		  width: 100%;
		  padding: 0.8em;
		  overflow: hidden;
		  list-style: none;
		  border: solid 3px white;
		  background: rgba(112, 128, 144, 0.2);
		}
		.Rtable-cell > h1,
		.Rtable-cell > h2,
		.Rtable-cell > h3,
		.Rtable-cell > h4,
		.Rtable-cell > h5,
		.Rtable-cell > h6 {
		  margin: 0;
		}
		/* Table column sizing
		================================== */
		.Rtable--2cols > .Rtable-cell {
		  width: 50%;
		}
		.Rtable--3cols > .Rtable-cell {
		  width: 33.33%;
		}
		.Rtable--4cols > .Rtable-cell {
		  width: 25%;
		}
		.Rtable--5cols > .Rtable-cell {
		  width: 20%;
		}
		.Rtable--6cols > .Rtable-cell {
		  width: 16.6%;
		}
		/* Page styling
		================================== */
		html {
		  height: 100%;
		  background-color: #EEE;
		}
		body {
		  box-sizing: border-box;
		  min-height: 100%;
		  margin: 0 auto;
		  padding: 0.8em;
		  max-width: 1000px;
		  font-family: 'Roboto Condensed', sans-serif;
		  font-size: 2em;
		  background-color: white;
		  border: double 3px #DDD;
		  border-top: none;
		  border-bottom: none;
		}
		h1,
		h2,
		h3,
		h4,
		h5,
		h6 {
		  margin-top: 0;
		}
		h3 {
		  font-size: 1.2em;
		}
		h4 {
		  font-size: 1em;
		}
		strong {
		  color: #434d57;
		}
		/* Apply styles
		================================== */
		.Rtable {
		  position: relative;
		  top: 3px;
		  left: 3px;
		}
		.Rtable-cell {
		  margin: -3px 0 0 -3px;
		  background-color: white;
		  border-color: #e2e6e9;
		}
		/* Cell styles
		================================== */
		.Rtable-cell--dark {
		  background-color: slategrey;
		  border-color: #5a6673;
		  color: white;
		}
		.Rtable-cell--dark > h1,
		.Rtable-cell--dark > h2,
		.Rtable-cell--dark > h3,
		.Rtable-cell--dark > h4,
		.Rtable-cell--dark > h5,
		.Rtable-cell--dark > h6 {
		  color: white;
		}
		.Rtable-cell--medium {
		  background-color: #b8c0c8;
		  border-color: #a9b3bc;
		}
		.Rtable-cell--light {
		  background-color: white;
		  border-color: #e2e6e9;
		}
		.Rtable-cell--highlight {
		  background-color: lightgreen;
		  border-color: #64e764;
		}
		.Rtable-cell--alert {
		  background-color: darkorange;
		  border-color: #cc7000;
		  color: white;
		}
		.Rtable-cell--alert > h1,
		.Rtable-cell--alert > h2,
		.Rtable-cell--alert > h3,
		.Rtable-cell--alert > h4,
		.Rtable-cell--alert > h5,
		.Rtable-cell--alert > h6 {
		  color: white;
		}
		.Rtable-cell--head {
		  background-color: slategrey;
		  border-color: #5a6673;
		  color: white;
		}
		.Rtable-cell--head > h1,
		.Rtable-cell--head > h2,
		.Rtable-cell--head > h3,
		.Rtable-cell--head > h4,
		.Rtable-cell--head > h5,
		.Rtable-cell--head > h6 {
		  color: white;
		}
		.Rtable-cell--foot {
		  background-color: #b8c0c8;
		  border-color: #a9b3bc;
		}
		/* Responsive
		==================================== */
		@media all and (max-width: 500px) {
		  .Rtable--collapse {
		    display: block;
		  }
		  .Rtable--collapse > .Rtable-cell {
		    width: 100% !important;
		  }
		  .Rtable--collapse > .Rtable-cell--foot {
		    margin-bottom: 1em;
		  }
		}
		.no-flexbox .Rtable {
		  display: block;
		}
		.no-flexbox .Rtable > .Rtable-cell {
		  width: 100%;
		}
		.no-flexbox .Rtable > .Rtable-cell--foot {
		  margin-bottom: 1em;
		}
		.bgfix {
			background-color: slategrey;
		}
		.widthfix {
			width: 100%;
		}
	</style>
	<script type="text/javascript">
	function getAlarms() {
		var h = document.getElementById('configuredAlarmsHeading');
		var t = document.getElementById('currentAlarms');
		t.innerHTML = "";
		$.getJSON("/alarm/all", function(data) {
			if (data !== "null") {
				JSON.parse(data).forEach(function(alarm) {
					var statusCol = document.createElement('div');
          statusCol.className = "Rtable-cell Rtable-cell--head";
					var statusButton = document.createElement('input');
					statusButton.setAttribute('type', 'button');
					statusButton.className = "widthfix bgfix";
					statusButton.setAttribute('value', alarm.status);
					var toggleStatus = (alarm.status === "active") ? "inactive" : "active";
					statusButton.setAttribute('onclick', 'toggleAlarm("' + alarm['_id'] + '","' + toggleStatus + '")')
					statusCol.appendChild(statusButton);

					var timeCol = document.createElement('div');
          timeCol.className = "Rtable-cell";
					var heading = document.createElement('h3');
					heading.innerHTML = alarm.hourMinute;
					timeCol.appendChild(heading);

					var timeframeCol = document.createElement('div');
          timeframeCol.className = "Rtable-cell";
					var timeframe = (alarm.weekDays === "true") ? "Mon-Fri" : "Sat-Sun";
					timeframeCol.appendChild(document.createTextNode(timeframe));

					var streamCol = document.createElement('div');
          streamCol.className = "Rtable-cell";
					streamCol.appendChild(document.createTextNode(alarm.streamName));

					var volumeEndCol = document.createElement('div');
          volumeEndCol.className = "Rtable-cell";
					volumeEndCol.appendChild(document.createTextNode('EndVol: ' + alarm.volumeEnd));

					var deleteCol = document.createElement('div');
          deleteCol.className = "Rtable-cell Rtable-cell--foot";
					var deleteButton = document.createElement('input');
					deleteButton.setAttribute('type', 'button');
					deleteButton.className = "widthfix";
					deleteButton.setAttribute('value', 'Delete');
					deleteButton.setAttribute('onclick', 'deleteAlarm("' + alarm['_id'] + '")')
					deleteCol.appendChild(deleteButton);

					t.appendChild(statusCol);
					t.appendChild(timeCol);
					t.appendChild(timeframeCol);
					t.appendChild(streamCol);
					t.appendChild(volumeEndCol);
					t.appendChild(deleteCol);
				});
		  }
		});
	}
	function populateZoneSelect() {
		var t = document.getElementById('raumfeldZoneSelect');
		t.innerHTML = "";
		$.getJSON("/zone/all", function(data) {
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
	function populateStreamSelect() {
		var t = document.getElementById('streamSelect');
		t.innerHTML = "";
		$.getJSON("/stream/all", function(data) {
			if (data !== "null") {
				JSON.parse(data).forEach(function(stream) {
					var s = document.createElement("option");
					s.text = stream.name;
					s.value = stream.name;
					t.add(s);
				});
			}
		});
	}
	function toggleAlarm(id, curr) {
		var alarmUpdate = {
			status: curr,
		};
		$.ajax({
			url: '/alarm/update/' + id,
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
		var weekDaysVal = (document.getElementById('timeframeSelect').value === "Mon-Fri") ? "true" : "false";
		var weekEndsVal = (document.getElementById('timeframeSelect').value === "Sat-Sun") ? "true" : "false";
		var newAlarm = {
			status: 'active',
			hourMinute: document.getElementById('timepicker').value,
			weekDays: weekDaysVal,
			weekEnds: weekEndsVal,
			zoneUuid: document.getElementById('raumfeldZoneSelect').value,
      streamName: document.getElementById('streamSelect').value,
		  volumeEnd: parseInt(document.getElementById('endVolumeSelect').value)
		};
		$.ajax({
			url: '/alarm/create',
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
	function stopAlarm() {
		$.ajax({
			url: '/alarm/stop',
			type: 'POST',
			success: function(data) {
				getAlarms();
			}
		});
	}
	function deleteAlarm(id) {
		$.ajax({
	    url: '/alarm/delete/' + id,
	    type: 'DELETE',
	    success: function(data) {
				getAlarms();
			}
	  });
	}
	</script>
	<style type="text/css">
		/* The Modal (background) */
		.modal {
		  display: none; /* Hidden by default */
		  position: fixed; /* Stay in place */
		  z-index: 1; /* Sit on top */
		  padding-top: 100px; /* Location of the box */
		  left: 0;
		  top: 0;
		  width: 100%; /* Full width */
		  height: 100%; /* Full height */
		  overflow: auto; /* Enable scroll if needed */
		  background-color: rgb(0,0,0); /* Fallback color */
		  background-color: rgba(0,0,0,0.4); /* Black w/ opacity */
		}

		/* Modal Content */
		.modal-content {
		  background-color: #ff0000;
		  margin: auto;
		  padding: 20px;
		  border: 1px solid #888;
		  width: 80%;
		}

		/* The Close Button */
		.close {
		  color: #dddddd;
		  float: center;
		  font-size: 20px;
		  font-weight: bold;
		}

		.close:hover,
		.close:focus {
		  color: #000;
		  text-decoration: none;
		  cursor: pointer;
		}
	</style>


</head>

<body onload="getAlarms();populateZoneSelect();populateStreamSelect();checkRunningAlarm()">

<!-- The Modal -->
<div id="myModal" class="modal">

  <!-- Modal content -->
  <div class="modal-content">
    <span class="close">Stop Alarm</span>
  </div>

</div>
	<script>
	// Get the modal
	var modal = document.getElementById('myModal');

	// Get the button that opens the modal
	var btn = document.getElementById("myBtn");

	// Get the <span> element that closes the modal
	var span = document.getElementsByClassName("close")[0];

	function checkRunningAlarm() {
		$.ajax({
			url: '/alarm/running',
			type: 'GET',
			success: function(data) {
				modal.style.display = "block";
			}
		});
	}

	// When the user clicks on <span> (x), close the modal
	span.onclick = function() {
		stopAlarm();
		getAlarms();
		modal.style.display = "none";
	}

	// When the user clicks anywhere outside of the modal, close it
	window.onclick = function(event) {
		if (event.target == modal) {
			modal.style.display = "none";
		}
	}
	</script>
	<h2>Configured Alarms</h2>

	<div class="Rtable Rtable--6cols Rtable--collapse" name="currentAlarms" id="currentAlarms">
	</div>

	<h2>Create New Alarm</h2>
	<div class="Rtable Rtable--6cols Rtable--collapse">
	  <div class="Rtable-cell Rtable-cell--head">
			<input id="timepicker" class="timepicker text-center bgfix widthfix" jt-timepicker time="model.time" time-string="model.timeString" default-time="model.options.defaultTime" time-format="model.options.timeFormat" start-time="model.options.startTime" min-time="model.options.minTime" max-time="model.options.maxTime" interval="model.options.interval" dynamic="model.options.dynamic" scrollbar="model.options.scrollbar" dropdown="model.options.dropdown" />
		</div>
	  <div class="Rtable-cell">
			<select name="timeframeSelect" id="timeframeSelect" class="widthfix">
				<option value="Mon-Fri">Mon-Fri</option>
				<option value="Sat-Sun">Sat-Sun</option>
			</select>
		</div>
		<div class="Rtable-cell">
			<select name="streamSelect" id="streamSelect" class="widthfix">
			</select>
		</div>
		<div class="Rtable-cell">
			<select name="raumfeldZoneSelect" id="raumfeldZoneSelect" class="widthfix">
			</select>
		</div>
		<div class="Rtable-cell">
			<select name="endVolumeSelect" id="endVolumeSelect" class="widthfix">
				<option value="35">Silent (35)</option>
				<option value="40">Normal (40)</option>
				<option value="45">Loud (45)</option>
			</select>
		</div>
		<div class="Rtable-cell Rtable-cell--foot">
			<input class="widthfix" type="button" onclick="createAlarm()" id="alarmEnabled" value="Add" />
		</div>
	</div>

	<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.1/jquery.min.js"></script>
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
    	startTime: '00',
    	dynamic: false,
    	dropdown: true,
    	scrollbar: true
		});
		$('input, select, textarea').on('focus blur', function(event) {
    	$('meta[name=viewport]').attr('content', 'width=device-width,initial-scale=1,maximum-scale=' + (event.type == 'blur' ? 10 : 1));
  	});
	});
	</script>

</body>
</html>
`
