/* jshint globalstrict: true */
/* globals window, document, console, alert, WebSocket */
"use strict";

// constants.js is not a real file. See server's http handler for details.
import constants from "./constants.js";

window.addEventListener("DOMContentLoaded", function () {
  document.querySelector("h1").innerHTML = "I <em>am</em> the JavaScript.";

  const ws = new WebSocket(constants.WEBSOCKET_URL);
  ws.onopen = (event) => {
    ws.send(JSON.stringify({ foo: "bar" }));
    ws.send(JSON.stringify({ foo2: "bar2" }));
  };

  // When server shuts down, the websocket will be closed,
  // in which case we should close the browser too:
  ws.onclose = (event) => {
    window.close();
  };
});
