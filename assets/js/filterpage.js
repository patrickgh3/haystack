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
    $(detailsElm).height(0).animate({height:'4em'}, 200);

    // Style stream element.
    streamElm.classList.add('selectedstream');

    // If dirty, then mark clean and send request.
    if (streamElm.dataset.dirty) {
        streamElm.dataset.dirty = ""; // falsy
        streamElm.dataset.detailsContent = 'Loading...';
        streamRequest(streamElm, 3, 200);
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
