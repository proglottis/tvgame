"use strict";

$(function() {
  var $players  = $(".players"),
      $start    = $(".start"),
      $question = $(".question"),
      $answers  = $question.find('.answers'),
      $join     = $(".join"),
      $timer     = $(".timer"),
      $lobby    = $(".lobby"),
      timer;

  var Timer = function (el) {
    var time_remaining = 30,
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
    console.log("Connection opened");
    conn.send(JSON.stringify({type: "create"}));
  };

  var state = lobby;
  conn.onmessage = function(event) {
    state = state(event);
  }

  function voteCollection(event) {
    var res   = JSON.parse(event.data),
        data  = res["Data"],
        action = res["Type"],
        mode  = 'answer';

    switch (action) {
    case "vote":
      var html = $.map(data["Question"]["Answers"], function (answer) {
        return '<li>' + answer["Text"] + '</li>';
      });
      $answers.html(html.join('')).show();
      // {"Type":"vote","Data":{"Question":{"Text":"In the city of Manchester (England) the Irk and Medlock join which river?","Multiplier":1,"Answers":[{"Correct":true,"Text":"IRWELL","Player":null,"Votes":null},{"Correct":false,"Text":"FOO","Player":{"ID":"04cdd7b5ca","Name":"25bb"},"Votes":null}]}}}
      break;
    case "collected":
      // {"Type":"collected","Data":{"Player":{"ID":"948cce4fae","Name":"ff85"},"Complete":true}}
      if ( data["Complete"] ) {

      }
      break;
    default:
      console.log("Uncaught message from voteCollection: " + event.data);
    }
    return voteCollection;
  }

  function answerCollection(event) {
    var res   = JSON.parse(event.data),
        data  = res["Data"],
        action = res["Type"];

    switch (action) {
    case "collected":
      // {"Type":"collected","Data":{"Player":{"ID":"948cce4fae","Name":"ff85"},"Complete":true}}
      if ( data["Complete"] ) {
          timer.stop();
          conn.send(JSON.stringify({type: "vote"}));
          return voteCollection;
      }
      break;
    default:
      console.log("Uncaught message from answerCollection: " + event.data);
    }
    return answerCollection;
  }

  function lobby(event) {
    var res   = JSON.parse(event.data),
        data  = res["Data"],
        action = res["Type"],
        mode  = 'answer';

    switch (action) {
    case "create":
      $lobby.append(data["Code"]);
      break;
    case "joined":
      $players.find('.blank').first().text(data["Player"]["Name"]).removeClass('blank');
      break;
    case "question":
      // {"Type":"question","Data":{"Question":{"Text":"In which year were premium bonds first issued in Britain?","Multiplier":1,"Answers":[{"Correct":true,"Text":"1956","Player":null,"Votes":null}]}}}
      $start.hide();
      $join.hide();
      $timer.show();
      $question.show().find('h1').text(data["Question"]["Text"]);
      timer = new Timer($timer);
      return answerCollection;
    default:
      console.log("Uncaught message from lobby: " + event.data);
    }
    return lobby;
  }

  conn.onclose = function(event) {
    console.log("Connection closed");
  };

  conn.onerror = function(event) {
    console.log("Error: " + event.data);
  };

  $('form').submit(function(event) {
    event.preventDefault();
    conn.send(JSON.stringify({type: "begin"}));
  });
});
