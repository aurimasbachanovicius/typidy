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

        if (!msg.value) {
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

    conn = new WebSocket("ws://" + document.location.host + "/chat-socket");

    conn.onmessage = function (event) {
        let message = JSON.parse(event.data);

        let writingSpotID = "writing-" + message["client_id"];
        let writingSpot = document.getElementById(writingSpotID);

        let isTyping = message["typing"] === true;

        if (message["message"].length === 0) {
            writingSpot.remove();

            return;
        }

        if (writingSpot && isTyping) {
            writingSpot.innerText = namedMessage(message);

            return;
        }

        if (isTyping) {
            let newWritingSpot = document.createElement("div");
            newWritingSpot.id = "writing-" + message["client_id"];

            appendMessage(writingMsg, newWritingSpot, message)
        } else {
            let committedMessage = document.createElement("div");
            appendMessage(committedMsg, committedMessage, message);
        }
    };

    function appendMessage(element, messageElement, message) {
        let doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;

        messageElement.innerText = namedMessage(message);
        element.appendChild(messageElement);

        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }

    function namedMessage(msg) {
        return msg["client_id"] + ": " + msg["message"];
    }
};