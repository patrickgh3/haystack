selectedStream = null;

// Expand a stream when clicked, and collapse others.
function streamClicked(streamElm) {
    // Just collapse if clicked on currently open stream.
    if (streamElm == selectedStream) {
        collapseStream(streamElm);
        selectedStream = null;
    // Otherwise, expand new stream and collapse previous one.
    } else {
        expandStream(streamElm);
        if (selectedStream != null) {
            collapseStream(selectedStream);
        }
        selectedStream = streamElm;
        // Update the details of the new selected stream.
        updateSelectedDetails();
    }
}

// Create the details element for this stream.
function expandStream(streamElm) {
    // Create element.
    var detailsElm = document.createElement('div');
    detailsElm.className = 'streamdetails';
    streamElm.insertAdjacentElement('afterend', detailsElm);
    $(detailsElm).height(0).animate({height:'5em'}, 200);

    // Style stream element.
    streamElm.classList.add('selectedstream');

    // If dirty, then mark clean and send request.
    if (streamElm.dataset.dirty) {
        streamElm.dataset.dirty = ""; // falsy
        streamElm.dataset.detailsContent = 'Loading...';
        streamRequest(streamElm, 3, 300);
    }
}

// Send request on behalf of a stream element, with exponential backoff.
function streamRequest(streamElm, max, delay) {
    var xhr = new XMLHttpRequest();
    var url = haystackBaseUrl+'/stream?id='+streamElm.dataset.streamid;
    xhr.open('GET', url, true);
    xhr.onreadystatechange = function() {
        if (xhr.readyState == XMLHttpRequest.DONE) {
            // Success
            if (xhr.status == 200) {
                streamElm.dataset.detailsContent = xhr.responseText;
                updateSelectedDetails();
            } else {
                // Retry
                if (max > 1) {
                    setTimeout(function() {
                        streamRequest(streamElm, max-1, delay*2);
                    }, delay);
                // Back off
                } else {
                    streamElm.dataset.dirty = '1'; // truthy
                    streamElm.dataset.detailsContent =
                        'Failed to get stream data, please try again later.';
                    updateSelectedDetails();
                }
            }
        }
    }
    xhr.send();
}

// Delete this stream's trailing details element.
function collapseStream(streamElm) {
    var elm = streamElm.nextElementSibling;
    $(elm).animate({height:0}, 200, function(){elm.remove();});

    // Style stream element.
    streamElm.classList.remove('selectedstream');
}

// Update the contents of the deatils element of the currently open stream.
function updateSelectedDetails() {
    if (selectedStream != null) {
        selectedStream.nextElementSibling.innerHTML =
            selectedStream.dataset.detailsContent;
    }
}

function updateFilter() {
    // Collapse selected stream.
    if (selectedStream != null) {
        selectedStream.nextElementSibling.remove();
        selectedStream.classList.remove('selectedstream');
        selectedStream = null;
    }

    var children = document.getElementById('streamList').children;
    var filter = document.getElementById('filterSelect').value;
    var searchText = document.getElementById('searchtext').value.toLowerCase();
    for (var i=0; i<children.length; i++) {
        var s = children[i];
        // Skip day spacers.
        if (s.classList.contains('day')) {
            continue;
        }
        var rightLink = s.children[1];

        var visible = true;
        // Apply filter (top10, etc)
        if (filter == 'top10') {
            visible = rightLink.dataset['filtertop10'];
        }

        // Check children for search text (this is janky!)
        if (searchText && visible) {
            var inChannel = false;
            var inTitle = false;
            for (var j=0; j<rightLink.children.length; j++) {
                var child = rightLink.children[j];
                if (child.classList.contains('channel')) {
                    inChannel = child.innerHTML.toLowerCase().
                        indexOf(searchText) != -1;
                }
                else if (child.classList.contains('title')) {
                    inTitle = child.innerHTML.toLowerCase().
                        indexOf(searchText) != -1;
                }
            }
            if (!inChannel && !inTitle) {
                visible = false;
            }
        }

        if (visible) {
            s.style.display = 'block';
        } else {
            s.style.display = 'none';
        }
    }

    // Save filter to cookie.
    createCookie('filter', filter);
}

// Not sure where I got this from...
function createCookie(name,value) {
      var expires = "; expires=" + 'Fri, 31 Dec 9999 23:59:59 GMT';
      document.cookie = name + "=" + value + expires + "; path=/";
}

function searchKeyPressed(e) {
    if (e.keyCode === 13) {
        e.preventDefault();
        search();
    }
}

// Search pressed
function search() {
    if (!document.getElementById('searchtext').value) {
        clearSearch();
    } else {
        document.getElementById('searchclearbutton').style.display = 'inline';
        updateFilter();
    }
}

// Clear pressed
function clearSearch() {
    console.log('clear');
    document.getElementById('searchclearbutton').style.display = 'none';
    document.getElementById('searchtext').value = '';
    updateFilter();
}

// Toggle dark mode
function toggleDarkMode() {
    darkmode = !darkmode;
    applyDarkMode();
}

function applyDarkMode() {
    if (darkmode) {
        document.getElementById('stylesheet').href = haystackBaseUrl + '/style_dark.css';
        document.getElementById('darkmodelink').innerHTML = 'Light Mode';
    } else {
        document.getElementById('stylesheet').href = haystackBaseUrl + '/style.css';
        document.getElementById('darkmodelink').innerHTML = 'Dark Mode';
    }
    createCookie('darkmode', darkmode);
}

// Set filter from cookie.
var filter = document.cookie.replace(/(?:(?:^|.*;\s*)filter\s*\=\s*([^;]*).*$)|^.*$/, "$1");
if (filter) {
    document.getElementById('filterSelect').value = filter;
    updateFilter();
}

// Set dark mode from cookie.
darkmode = document.cookie.replace(/(?:(?:^|.*;\s*)darkmode\s*\=\s*([^;]*).*$)|^.*$/, "$1") == 'true';
if (darkmode) {
    applyDarkMode();
}
