"use strict";

$(function() {
  var $log = $("#log");

  function appendLog(msg) {
    $log.append($("<div>").text(msg));
  }

  var conn;

  $('form').submit(function(event) {
    event.preventDefault();
    conn = new WebSocket($('body').data('url'));


    conn.onopen = function(event) {
      appendLog("Connection opened");
      conn.send(JSON.stringify({Type: 'join', Name: $('input[name=name]').val(), Code: $('input[name=code]').val()}))
    };

    conn.onmessage = function(event) {
      appendLog("Message: " + event.data);
    };

    conn.onclose = function(event) {
      appendLog("Connection closed");
    };

    conn.onerror = function(event) {
      appendLog("Error: " + event.data);
    };
  });
});
