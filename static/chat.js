window.onload = function () {
    let conn;
    let msg = document.getElementById("msg"),
        writingMsg = document.getElementById("writing-messages"),
        committedMsg = document.getElementById("committed-messages");

    document.getElementById("form").onsubmit = function (event) {
        event.preventDefault();

        if (!conn) {
            return false;
        }

        conn.send(msg.value + "\n");

        msg.value = "";

        return false;
    };

    document.getElementById("form").onkeyup = function () {
        if (!conn) {
            return false;
        }

        conn.send(msg.value);

        return false;
    };

    if (!window["WebSocket"]) {
        let item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }

    conn = new WebSocket("ws://" + document.location.host + "/ws");
    
    conn.onmessage = function (event) {
        let message = JSON.parse(event.data);

        let writingSpotID = "writing-" + message["client_id"];
        let writingSpot = document.getElementById(writingSpotID);

        console.log(message);

        let isTyping = message["typing"] === true;

        if (writingSpot && isTyping) {
            writingSpot.innerText = message["message"];

            return;
        }

        if (isTyping) {
            let newWritingSpot = document.createElement("div");
            newWritingSpot.innerText = message["message"];
            newWritingSpot.id = "writing-" + message["client_id"];

            writingMsg.appendChild(newWritingSpot);
        } else {
            let committedMessage = document.createElement("div");
            committedMessage.innerText = message["message"];

            committedMsg.appendChild(committedMessage)
        }
    };
};