// Update magnifier to a given thumb element
function magnify(event, thumbElt, hasVod) {
    // Grab magnifier elements
    var m = document.getElementById("magnifier");
    var img = m.children[0];

    // Set X centered on mouse cursor.
    var x = event.pageX;
    x -= img.width/2;
    // Clamp x on screen
    //x = Math.max(10, Math.min(document.body["scrollWidth"]-(img.width+10),x));

    // Set Y above thumb, or below if that's off the top of the screen.
    var thumbTop = thumbElt.getBoundingClientRect().top;
    // Extra 1 px for Firefox
    var y = thumbTop + window.pageYOffset - m.clientHeight + 1;
    if (y < window.pageYOffset) {
        y = thumbElt.getBoundingClientRect().bottom + window.pageYOffset;
    }

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
