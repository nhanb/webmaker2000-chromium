/* jshint globalstrict: true */
/* globals window, document, console, alert, WebSocket */
"use strict";
import conf from "./config.js";

window.addEventListener("DOMContentLoaded", function () {
  document.querySelector("h1").innerHTML = "I <em>am</em> the JavaScript.";

  const ws = new WebSocket(conf.WEBSOCKET_URL);
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
