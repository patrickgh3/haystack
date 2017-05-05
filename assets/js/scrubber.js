// http://javascript.info/coordinates#getCoords
function getDocumentCoords(elem) {
  let box = elem.getBoundingClientRect();
  return {
    top: box.top + pageYOffset,
    bottom: box.bottom + pageYOffset,
    left: box.left + pageXOffset,
    right: box.right + pageXOffset
  };
}

lastSelectedPanel = null;

function selectPanel(panelElt) {
    var s = document.getElementById("scrubber");
    var containerCoords = getDocumentCoords(panelElt.parentElement);
    var x = (containerCoords.left + containerCoords.right)/2 - s.offsetWidth/2;
    var y = getDocumentCoords(panelElt).bottom;

    s.style.visibility = "visible";
    s.style.left = x+"px";
    s.style.top = y+"px";

    if (lastSelectedPanel != null) {
        lastSelectedPanel.classList.remove('selected');
        lastSelectedPanel.classList.add('deselected');
    }

    panelElt.classList.remove('deselected');
    panelElt.classList.add('selected');

    lastSelectedPanel = panelElt;
}

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
