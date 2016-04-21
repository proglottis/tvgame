"use strict";

$(function() {
  var $log         = $("#log"),
      $join_form   = $('#join'),
      $waiting     = $('.waiting'),
      $question    = $('.question'),
      $answer_form = $('#answer');

  function appendLog(msg) {
    $log.append($("<div>").text(msg));
  }

  var conn;

  $join_form.submit(function(event) {
    event.preventDefault();
    conn = new WebSocket($('body').data('url'));


    conn.onopen = function(event) {
      $join_form.hide();
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
        $question.show().find('h2').text(data["Data"]["Text"]);
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

  $answer_form.submit(function (event) {
    event.preventDefault();
    conn.send(JSON.stringify({Type: 'answer', Data: {
      Text: $('textarea[name=answer]').val()
    }}))
  });
});
