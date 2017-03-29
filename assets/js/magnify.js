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

function magnify(elt, hasVod) {
    var m = document.getElementById("magnifier");
    var img = m.children[0];

    var x = event.pageX;
    x -= img.width/2;
    x = Math.max(10, Math.min(document.body["scrollWidth"]-(img.width+10),x));
    var y = getDocumentCoords(elt).bottom+0;
    
    m.style.visibility = "visible";
    m.style.left = x+"px";
    m.style.top = y+"px";
    img.src = elt.src;
    
    if (hasVod) {
        img.style.opacity = 1;
    }
    else {
        img.style.opacity = .7;
    }
}

function unmagnify() {
    var m = document.getElementById("magnifier");
    m.style.visibility = "hidden";
}
