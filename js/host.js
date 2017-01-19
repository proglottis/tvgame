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

  // var conn = new WebSocket($('body').data('url'));
  // 
  // conn.onopen = function(event) {
  //   console.log("Connection opened");
  //   conn.send(JSON.stringify({type: "create"}));
  // };
  // 
  // var state = lobby;
  // conn.onmessage = function(event) {
  //   state = state(event);
  // }

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
});

function StartScene(){
  Phaser.State.call(this);
}
StartScene.prototype = Object.create(Phaser.State.prototype);

StartScene.prototype.init = function(conn) {
  this.conn = conn;
}

StartScene.prototype.create = function(){
  console.log("StartScene");
};

StartScene.prototype.onConnection = function() {
  this.conn.send(JSON.stringify({type: "create"}));
}

StartScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "create":
      this.state.start("lobby", true, false, this.conn, event.Data.Code);
      break;
    default:
      console.log(event);
  }
}

function LobbyScene(){
  Phaser.State.call(this);
}
LobbyScene.prototype = Object.create(Phaser.State.prototype);

LobbyScene.prototype.init = function(conn, code) {
  this.conn = conn;
  this.code = code;
  this.players = [];
};

LobbyScene.prototype.preload = function() {
  this.load.image("bg", "css/green-background.jpg");
}

LobbyScene.prototype.create = function() {
  console.log("LobbyScene");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;
  this.add.text(50,0, this.code, {fill: "#ff0000"});
  this.start_button = this.add.text(50,50, "Everybody IN!", {fill: "#ff0000"});
  this.start_button.inputEnabled = true;
  this.start_button.events.onInputDown.add(this.listener, this);
};

LobbyScene.prototype.listener = function(){
  this.conn.send(JSON.stringify({type: "begin"}));
};

LobbyScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "question":
      this.state.start("lie",true,false,this.conn, event.Data.Question);
      break;
    case "joined":
      const player = event.Data.Player;
      this.add.text(50,100+(this.players.length*50), player.Name, {fill: "#ff0000"});
      this.players.push(player);
      break;
    default:
      console.log(event);
  }
};

function LieScene() {
  Phaser.State.call(this);
}

LieScene.prototype = Object.create(Phaser.State.prototype);

LieScene.prototype.preload = function() {
  this.load.image("bg", "css/green-background.jpg");
}

LieScene.prototype.init = function(conn, question) {
  this.conn = conn;
  this.players = [];
  this.question = question;
};

LieScene.prototype.create = function() {
  console.log("LieScene");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;
  const question = this.add.text(50,0, this.question.Text, {fill: "#ff0000", wordWrap: true, wordWrapWidth: this.world.width - 100});

  this.timer = game.time.create(true);
  this.timer.add(30000, this.endRound, this);
  this.timer.start();
};

LieScene.prototype.endRound = function(){
  conn.send(JSON.stringify({type: "stop"}));
}

LieScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "collected":
      if(event.Data.Complete) {
        this.endRound();
      }
      this.add.text(50, 100+this.players.length*50, event.Data.Player.Name, {fill: "#ff0000"});
      break;
    case "vote":
      this.state.start("vote", true, false, this.conn, event.Data.Question);
      break;
    default:
      console.log(event);
  }
};

function VoteScene() {
  Phaser.State.call(this);
}

VoteScene.prototype = Object.create(Phaser.State.prototype);

VoteScene.prototype.preload = function() {
  this.load.image("bg", "css/green-background.jpg");
}

VoteScene.prototype.init = function(conn, question) {
  this.conn = conn;
  this.question = question;
};

VoteScene.prototype.create = function() {
  console.log("VoteScene");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;
  const question = this.add.text(50,0, this.question.Text, {fill: "#ff0000", wordWrap: true, wordWrapWidth: this.world.width - 100});

  for(var i = 0; i < this.question.Answers.length; i++) {
    const answer = this.question.Answers[i];
    this.add.text(50, i*50+100, answer.Text, {fill: "#ff0000"});
  }

  this.timer = game.time.create(true);
  this.timer.add(30000, this.endRound, this);
  this.timer.start();
};

VoteScene.prototype.endRound = function(){
  conn.send(JSON.stringify({type: "stop"}));
}

VoteScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "collected":
      if(event.Data.Complete) {
        this.endRound();
      }
      break;
    case "results":
      this.state.start("score", true, false, this.conn, this.question, event.Data.Offsets, event.Data.Points);
      break;
    default:
      console.log(event);
  }
};

function ScoreScene() {
  Phaser.State.call(this);
}

ScoreScene.prototype = Object.create(Phaser.State.prototype);

ScoreScene.prototype.preload = function() {
  this.load.image("bg", "css/green-background.jpg");
}

ScoreScene.prototype.init = function(conn, question, offsets, points) {
  this.conn = conn;
  this.question = question;
  this.offsets = offsets;
  this.points = points;
};

ScoreScene.prototype.create = function() {
  console.log("ScoreScene");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;
  const question = this.add.text(50,0, this.question.Text, {fill: "#ff0000", wordWrap: true, wordWrapWidth: this.world.width - 100});

  const goodStyle = {fill: "#0000ff"};
  const badStyle = {fill: "#ff0000"};

  for(var i = 0; i < this.offsets.length; i++) {
    const answer = this.offsets[i].Answer;
    if (answer.Correct) {
      this.add.text(50, i*50+100, `${answer.Text} - ${answer.Votes.length} votes`, goodStyle);
    } else {
      this.add.text(50, i*50+100, `${answer.Text} - ${answer.Votes.length} votes`, badStyle);
    }
  }

  this.timer = game.time.create(true);
  this.timer.add(30000, this.endRound, this);
  this.timer.start();
};

ScoreScene.prototype.endRound = function(){
}

ScoreScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    default:
      console.log(event);
  }
};

const game = new Phaser.Game(800,600, Phaser.AUTO, '');
game.state.add("start", new StartScene());
game.state.add("lobby", new LobbyScene());
game.state.add("lie", new LieScene());
game.state.add("vote", new VoteScene());
game.state.add("score", new ScoreScene());

const conn = new WebSocket($('body').data('url'));
conn.onopen = function(event) {
  const state = game.state.getCurrentState();
  state.onConnection.call(state);
};
conn.onmessage = function(event) {
  const state = game.state.getCurrentState();
  const msg = JSON.parse(event.data);
  console.log(msg.Type);
  state.onMessage.call(state, msg);
};

game.state.start("start", true, false, conn);