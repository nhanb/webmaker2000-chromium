/* jshint globalstrict: true */
/* globals window, document, console, alert, WebSocket */
"use strict";

// constants.js is not a real file. See server's http handler for details.
import constants from "./constants.js";

var appState = {};

window.addEventListener("DOMContentLoaded", function () {
  const ws = new WebSocket(constants.WEBSOCKET_URL);

  function request(msg) {
    ws.send(JSON.stringify(msg));
  }

  ws.onopen = (event) => {
    request({ proc: "getstate" });
  };

  // When server shuts down, the websocket will be closed,
  // in which case we should close the browser too:
  ws.onclose = (event) => {
    window.close();
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    switch (msg.type) {
      case "state":
        appState = msg.data;
        document.querySelector("#app").innerHTML =
          "<pre>" + JSON.stringify(appState, "", 2) + "</pre>";
        break;
      default:
        console.log("Unknown msg:", msg);
        break;
    }
  };
});
