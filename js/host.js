"use strict";

$(function() {
  var $log      = $("#log"),
      $players  = $(".players"),
      $start    = $(".start"),
      $question = $(".question"),
      $lobby    = $(".lobby"),
      timer;

  function appendLog(msg) {
    $log.append($("<div>").text(msg));
  }

  var Timer = function (el) {
    var time_remaining = 10,
        seconds        = el.find('.seconds'),
        interval;

    seconds.text(time_remaining);
    interval = setInterval(showTime.bind(this), 1000);

    function showTime() {
      if ( time_remaining === 0 ) {
        this.stop();
        conn.send(JSON.stringify({type: "vote"}));
      } else {
        time_remaining--;
        seconds.text(time_remaining);
      }
    };

    this.stop = function () {
      clearInterval(interval);
    };
  }

  var conn = new WebSocket($('body').data('url'));

  conn.onopen = function(event) {
    appendLog("Connection opened");
    conn.send(JSON.stringify({type: "create"}));
  };

  conn.onmessage = function(event) {
    var res   = JSON.parse(event.data),
        data  = res["Data"],
        action = res["Type"];

    switch (action) {
    case "create":
      $lobby.append(data["Code"]);
      break;
    case "joined":
      $start.show();
      $players.append("<li>" + data["Player"]["Name"] + "</li>");
      break;
    case "question":
      // {"Type":"question","Data":{"Question":{"Text":"In which year were premium bonds first issued in Britain?","Multiplier":1,"Answers":[{"Correct":true,"Text":"1956","Player":null,"Votes":null}]}}}
      $start.hide();
      $question.show().find('h1').text(data["Question"]["Text"]);
      timer = new Timer($question.find('.timer'));
      break;
    case "collected":
      // {"Type":"collected","Data":{"Player":{"ID":"948cce4fae","Name":"ff85"},"Complete":true}}
      if ( data["Complete"] ) {
        timer.stop();
        conn.send(JSON.stringify({type: "vote"}));
      }
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
