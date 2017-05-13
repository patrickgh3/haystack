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

// Update magnifier to a given thumb element
function magnify(thumbElt, hasVod) {
    // Grab magnifier elements
    var m = document.getElementById("magnifier");
    var img = m.children[0];

    // Calculate new coordinates
    var x = event.pageX;
    x -= img.width/2;
    x = Math.max(10, Math.min(document.body["scrollWidth"]-(img.width+10),x));
    var y = getDocumentCoords(thumbElt).bottom+0;

    // Move magnifier
    m.style.visibility = "visible";
    m.style.left = x+"px";
    m.style.top = y+"px";

    // Set image source
    img.src = thumbElt.src;

    // Adjust opacity depending on if VOD exists
    if (hasVod) {
        img.style.opacity = 1;
    }
    else {
        img.style.opacity = .7;
    }
}

// Hide magnifier
function unmagnify() {
    var m = document.getElementById("magnifier");
    m.style.visibility = "hidden";
}
