// WebSocket objects - created when window is loaded.
let sockEcho = null,
    // Websocket server address.
    wsServerAddress = "ws://127.0.0.1:4050",
    canvas = document.getElementById("canvas"),
    ctx = canvas.getContext("2d"),
    ptr = new Image(20, 20);

ptr.src = "assets/pointer.png";

window.onload = () => {
    // Connect the WebSocket to the server and register callbacks on it.
    sockEcho = new WebSocket(wsServerAddress + "/wsecho");

    sockEcho.onopen = () => {
        console.log("connected");
    }

    sockEcho.onclose = e => {
        console.log("connection closed (" + e.code + ")");
    }

    sockEcho.onmessage = e => {
        var msg = JSON.parse(e.data);
        var coordMsg = "Coordinates: (" + msg.x + "," + msg.y + ")";
        document.getElementById("output").innerHTML = coordMsg;
    }

    sockTime = new WebSocket(wsServerAddress + "/wstime");
    sockTime.onmessage = e => {
        let cursors = JSON.parse(e.data).cursors;
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        if (cursors) {
            cursors.forEach(cursor => {
                ctx.drawImage(ptr,
                              cursor.x - 5,
                              cursor.y - 2,
                              20, 20);
            });
        }
    }
};

canvas.onmousemove = e => {
    // When a "mouse moved" event is invoked, send it on the socket.
    socketSend({x: e.offsetX, y: e.offsetY});
}

canvas.onmouseout = () => {
    document.getElementById("output").innerHTML = "";
}

// Send the msg object, encoded with JSON, on the websocket if it's open.
function socketSend(msg) {
    if (sockEcho != null && sockEcho.readyState == WebSocket.OPEN) {
        sockEcho.send(JSON.stringify(msg));
    } else {
        console.log("Socket isn't OPEN");
    }
}
