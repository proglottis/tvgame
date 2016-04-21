"use strict";

$(function() {
  var $log      = $("#log"),
      $players  = $(".players"),
      $start    = $(".start"),
      $question = $(".question"),
      $lobby    = $(".lobby");

  function appendLog(msg) {
    $log.append($("<div>").text(msg));
  }

  var Timer = function (el) {
    var time_remaining = 10,
        seconds        = el.find('.seconds'),
        interval;

    seconds.text(time_remaining);
    interval = setInterval(showTime, 1000);

    function showTime() {
      if ( time_remaining === 0 ) {
        clearInterval(interval);
        conn.send(JSON.stringify({type: "vote"}));
      } else {
        time_remaining--;
        seconds.text(time_remaining);
      }
    }
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
      $start.show();
      $players.append("<li>" + data["Data"]["Player"]["Name"] + "</li>");
      break;
    case "question":
      $start.hide();
      $question.show().find('h1').text(data["Data"]["Question"]["Text"]);
      new Timer($question.find('.timer'));
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
