// http://javascript.info/coordinates#getCoords
// Get absolute document coordinates from element coordinates
function getDocumentCoords(elem) {
  let box = elem.getBoundingClientRect();
  return {
    top: box.top + pageYOffset,
    bottom: box.bottom + pageYOffset,
    left: box.left + pageXOffset,
    right: box.right + pageXOffset
  };
}

// Keep reference to currently selected (last selected) stream panel
lastSelectedPanel = null;

// Select a given panel element when it's clicked
function selectPanel(panelElt) {
    // Grab scrubber and calculate its new coordinates
    var s = document.getElementById("scrubber");
    var containerCoords = getDocumentCoords(panelElt.parentElement);
    var x = (containerCoords.left + containerCoords.right)/2 - s.offsetWidth/2;
    var y = getDocumentCoords(panelElt).bottom;

    // Position scrubber
    s.style.visibility = "visible";
    s.style.left = x+"px";
    s.style.top = y+"px";

    // Set selected and deselected CSS classes of current panel and last panel
    if (lastSelectedPanel != null) {
        lastSelectedPanel.classList.remove('selected');
        lastSelectedPanel.classList.add('deselected');
    }
    panelElt.classList.remove('deselected');
    panelElt.classList.add('selected');
    lastSelectedPanel = panelElt;

    // Request scrubber HTML if not gotten yet
    if (!panelElt.dataset.clicked) {
        panelElt.dataset.clicked = "1";
        var imgurl = haystackBaseUrl+'/images/loading.gif';
        panelElt.dataset.scrubberhtml = '<img src="'+imgurl+'">';
        tryGetStream(panelElt, 5, 100);
    }

    updateScrubberContents();
}

// Set scrubber contents to html of the currently selected panel
function updateScrubberContents() {
    if (lastSelectedPanel != null) {
        var s = document.getElementById("scrubber");
        s.innerHTML = lastSelectedPanel.dataset.scrubberhtml;
    }
}

// Send a request for a panel's stream, and save/update the result if successful
function tryGetStream(panelElt, max, delay) {
    var xhr = new XMLHttpRequest();
    var url = haystackBaseUrl+'/stream?id='+panelElt.dataset.streamid;
    xhr.open('GET', url, true);
    xhr.onreadystatechange = function() {
        if (this.readyState == 4) {
            if (this.status == 200) {
                panelElt.dataset.scrubberhtml = xhr.responseText;
                updateScrubberContents();
            }
            else {
                if (max > 0) {
                    setTimeout(function() {
                        tryGetStream(panelElt, max - 1, delay * 2);
                    }, delay);
                } else {
                    panelElt.dataset.clicked = "";
                    panelElt.dataset.scrubberhtml =
                            "Unable to load stream, please try again later.";
                    updateScrubberContents();
                }
            }
        }
    }
    xhr.send();
}

// Deselect panel and hide scrubber
function deselectPanels() {
    if (lastSelectedPanel != null) {
        lastSelectedPanel.classList.remove('selected');
        lastSelectedPanel.classList.add('deselected');
    }
    lastSelectedPanel = null;
    var s = document.getElementById("scrubber");
    s.style.visibility = "hidden";
}

//document.addEventListener('click', deselectPanels);
