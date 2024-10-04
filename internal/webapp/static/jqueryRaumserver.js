function getAlarms() {
    var h = document.getElementById('configuredAlarmsHeading');
    var t = document.getElementById('currentAlarms');
    t.innerHTML = "";
    $.getJSON("/alarm/all", function(data) {
        if (data !== "null") {
            h.style.display = "block";
            t.style.display = "flex";
            JSON.parse(data).forEach(function(alarm) {
                // active / inactive toggle
                var statusCol = document.createElement('div');
                statusCol.className = "Rtable-cell Rtable-cell--head";
                var statusButton = document.createElement('input');
                statusButton.setAttribute('type', 'button');
                statusButton.className = "widthfix bgfix";
                statusButton.setAttribute('value', alarm.status);
                var toggleStatus = (alarm.status === "active") ? "inactive" : "active";
                statusButton.setAttribute('onclick', 'toggleAlarm("' + alarm['_id'] + '","' + toggleStatus + '")')
                statusCol.appendChild(statusButton);

                // time and days
                var timeCol = document.createElement('div');
                timeCol.className = "Rtable-cell";
                var heading = document.createElement('h3');
                var timeframe = (alarm.weekDays === "true") ? "Mon-Fri" : "Sat-Sun";
                heading.innerHTML = alarm.hourMinute + ' ' + timeframe;
                timeCol.appendChild(heading);

                // stream and volume
                var streamCol = document.createElement('div');
                streamCol.className = "Rtable-cell";
                streamCol.appendChild(document.createTextNode(alarm.streamName + ' - MaxVol. ' + alarm.volumeEnd));

                // delete
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
                //t.appendChild(timeframeCol);
                t.appendChild(streamCol);
                //t.appendChild(volumeEndCol);
                t.appendChild(deleteCol);
            });
        } else {
            h.style.display = "none";
            t.style.display = "none";
        }
    });
}
function populateZoneSelect() {
    var t = document.getElementById('raumfeldZoneSelect');
    if (t !== null) {
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
}
function populateStreamSelect() {
    var t = document.getElementById('streamSelect');
    if (t !== null) {
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
            console.log('alarm stopped successfully')
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