"use strict";

$(function() {
  var $players         = $(".players"),
      $start           = $(".start"),
      $question        = $(".question"),
      $answers         = $(".answers"),
      $join            = $(".join"),
      $timer           = $(".timer"),
      $lobby           = $(".lobby"),
      $place_your_vote = $('.place-your-vote'),
      $scoreboard = $('.scoreboard'),
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
        conn.send(JSON.stringify({type: "stop"}));
      } else {
        time_remaining--;
        seconds.text(time_remaining);
      }
    };

    this.reset = function () {
      this.stop();
      time_remaining = 30;
      seconds.text(time_remaining);
      interval = setInterval(showTime.bind(this), 1000);
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
        action = res["Type"];

    switch (action) {
    case "collected":
      // {"Type":"collected","Data":{"Player":{"ID":"948cce4fae","Name":"ff85"},"Complete":true}}
      if ( data["Complete"] ) {
        console.log("received all votes");
        conn.send(JSON.stringify({type: "stop"}));
        timer.stop();
        $timer.hide();
        $place_your_vote.hide();
        return lobby;
      }
      break;
    default:
      console.log("Uncaught message from voteCollection: " + event.data);
      return lobby(event);
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
          conn.send(JSON.stringify({type: "vote"}));
          return lobby;
      }
      break;
    default:
      console.log("Uncaught message from answerCollection: " + event.data);
      return lobby(event);
    }
    return answerCollection;
  }

  function lobby(event) {
    var res   = JSON.parse(event.data),
        data  = res["Data"],
        action = res["Type"];

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
      $scoreboard.hide();
      $join.hide();
      $place_your_vote.hide();
      $timer.show();
      $question.show().text(data["Question"]["Text"]);
      timer = new Timer($timer);
      return answerCollection;
    case "vote":
      var question_text = data["Question"]["Text"];
      var answers = $.map(data["Question"]["Answers"], function (answer) {
        return '<h2>' + answer["Text"] + '</h2>';
      });
      $answers.html(answers.join('')).show();
      $place_your_vote.show();
      timer.reset();
      // {"Type":"vote","Data":{"Question":{"Text":"In the city of Manchester (England) the Irk and Medlock join which river?","Multiplier":1,"Answers":[{"Correct":true,"Text":"IRWELL","Player":null,"Votes":null},{"Correct":false,"Text":"FOO","Player":{"ID":"04cdd7b5ca","Name":"25bb"},"Votes":null}]}}}
      return voteCollection;
    case "results":
      console.log('received scores');
      $place_your_vote.hide();
      $timer.hide();
      $question.hide();
      $answers.hide();
      var scores = $.map(data["Points"].sort(function(a, b) { return b["Total"] - a["Total"]; }), function (score) {
        return '<tr><td>' + score["Player"]["Name"] + '</td><td>' + score["Total"] + '</td></tr>';
      });
      $scoreboard.show().find('tbody').html(scores.join(''));
      setTimeout(function () { conn.send(JSON.stringify({type: "next"})) }, 5000);
      break;
      // {"Type":"results","Data":{"Points":[{"Player":{"ID":"XJWKFEUYLX","Name":"ALSAQ"},"Total":1500}],"Offsets":[{"Answer":{"Correct":true,"Text":"EGYPT","Player":null,"Votes":[{"ID":"XJWKFEUYLX","Name":"ALSAQ"}]},"Offsets":[{"Player":{"ID":"XJWKFEUYLX","Name":"ALSAQ"},"Offset":1500}]}]}}
    case "complete":
      console.log('complete');
      $('.game-title').text('Someone is the winner!');
      break;
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
  
  $().ready(function(){
  });
});


const game = new Phaser.Game(800,600, Phaser.AUTO, '');
game.state.add("start", new StartScene());
game.state.start("start");

function StartScene(){
  Phaser.State.call(this);
}
StartScene.prototype = Object.create(Phaser.State.prototype);

StartScene.prototype.create = function(){
  this.start_button = new Phaser.Text(this.game, 50,50, "Everybody IN!");
};