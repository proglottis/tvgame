"use strict";

$(function() {
  var $log      = $("#log"),
      $form     = $('form'),
      $waiting  = $('.waiting'),
      $question = $('.question');

  function appendLog(msg) {
    $log.append($("<div>").text(msg));
  }

  var conn;

  $form.submit(function(event) {
    event.preventDefault();
    conn = new WebSocket($('body').data('url'));


    conn.onopen = function(event) {
      $form.hide();
      $waiting.show();
      conn.send(JSON.stringify({Type: 'join', Data: {
        Name: $('input[name=name]').val(),
        Code: $('input[name=code]').val()
      }}))
    };

    conn.onmessage = function(event) {
      var data   = JSON.parse(event.data),
          action = data["Type"];

      switch(action) {
      case "answer":
        $waiting.hide();
        $question.text(data["Data"]["Text"]).show();
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
  });
});
