setInterval(function() {
    document.body.style.backgroundColor = "#"+((1<<24)*Math.random()|0).toString(16);
}, 2000);