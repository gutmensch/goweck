// Get the stopModal
var stopModal = document.getElementById('myModal');

// Get the button that opens the stopModal
var btn = document.getElementById("myBtn");

// Get the <span> element that closes the stopModal
var span = document.getElementsByClassName("close")[0];

function checkRunningAlarm() {
    $.ajax({
        url: '/alarm/running',
        type: 'GET',
        success: function(data) {
            stopModal.style.display = "block";
        }
    });
}

// When the user clicks on <span> (x), close the stopModal
span.onclick = function() {
    stopAlarm();
    stopModal.style.display = "none";
}

// When the user clicks anywhere outside of the stopModal, close it
window.onclick = function(event) {
    if (event.target == stopModal) {
        stopModal.style.display = "none";
    }
}