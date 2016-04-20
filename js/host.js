"use strict";

$(function() {
  var $log     = $("#log"),
      $players = $(".players"),
      $lobby   = $(".lobby");

  function appendLog(msg) {
    $log.append($("<div>").text(msg));
  }

  var conn = new WebSocket($('body').data('url'));

  conn.onopen = function(event) {
    appendLog("Connection opened");
    conn.send(JSON.stringify({type: "create"}));
  };

  conn.onmessage = function(event) {
    var data   = JSON.parse(event.data),
        action = data["Type"];

    switch(action) {
    case "create":
      $lobby.append(data["Data"]["Code"]);
      break;
    case "joined":
      $players.append("<li>" + data["Data"]["Player"]["Name"] + "</li>");
      break;
    default:
      appendLog("Message: " + event.data);
    }
  };

  conn.onclose = function(event) {
    appendLog("Connection closed");
  };

  conn.onerror = function(event) {
    appendLog("Error: " + event.data);
  };

  $('form').submit(function(event) {
    event.preventDefault();
    conn.send(JSON.stringify({type: "begin"}));
  });
});
