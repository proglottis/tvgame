"use strict";

const QUESTION_STYLE = {
  font: 'bold 40pt Arial',
  fill: '#F5F5DC',
  align: 'center',
  stroke: '#000000',
  strokeThickness: 6,
  wordWrap: true,
}

function StartScene(){
  Phaser.State.call(this);
}

StartScene.prototype = Object.create(Phaser.State.prototype);

StartScene.prototype.init = function(conn) {
  this.conn = conn;
}

StartScene.prototype.preload = function() {
  this.scale.scaleMode = Phaser.ScaleManager.RESIZE;
  this.scale.setResizeCallback(function() {
    this.scale.setMaximum();
  }.bind(this));
}

StartScene.prototype.create = function(){
  console.log("StartScene");
  this.waiting = false;
};

StartScene.prototype.update = function(event) {
  if(this.conn.open && !this.waiting) {
    this.conn.send(JSON.stringify({type: "create"}));
    this.waiting = true;
  }
  var event = this.conn.get();
  if(event != null) {
    switch(event.Type) {
      case "create":
        this.state.start("lobby", true, false, this.conn, event.Data.Code);
        break;
      default:
        console.log(event);
    }
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
  this.load.image('rain', 'css/rain.png');
  this.load.audio('music', ['css/Funky Chunk.mp3']);
}

LobbyScene.prototype.create = function() {
  console.log("LobbyScene");

  const music = game.add.audio('music');
  music.play();

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;

  var emitter = game.add.emitter(game.world.centerX, 0, 100);
  emitter.width = game.world.width;
  emitter.makeParticles('rain');
  emitter.minParticleScale = 0.1;
  emitter.maxParticleScale = 5;
  emitter.setYSpeed(300, 500);
  emitter.setXSpeed(0, 0);
  emitter.minRotation = 0;
  emitter.maxRotation = 0;
  emitter.start(false, 1600, 5, 0);

  const heading = this.add.text(this.world.centerX, 10, "TVGame", {
    font: 'bold 72pt Arial',
    fill: '#F5F5DC',
  });
  heading.anchor.set(0.5, 0);

  const startBtn = this.add.text(this.world.centerX, this.world.centerY, "EVERYBODY'S IN!", {
    fill: "#ff0000"
  });
  startBtn.anchor.set(0.5, 0.5);
  startBtn.inputEnabled = true;
  startBtn.events.onInputDown.add(this.listener, this);

  const press = this.add.text(0,0, "Press", {
    fill: "#F5F5DC"
  });
  press.alignTo(startBtn, Phaser.TOP_CENTER);
  const tostart = this.add.text(0,0, "to start", {
    fill: "#F5F5DC"
  });
  tostart.alignTo(startBtn, Phaser.BOTTOM_CENTER);

  const credit = this.add.text(this.world.centerX, this.world.height - 10, "Funky Chunk Kevin MacLeod (incompetech.com) Licensed under Creative Commons: By Attribution 3.0 License http://creativecommons.org/licenses/by/3.0/", {
    font: "15pt",
    wordWrap: true,
    wordWrapWidth: this.world.width
  });
  credit.anchor.set(0.5, 1);

  const code = this.add.text(this.world.centerX, this.world.height - 10, this.code, {
    font: 'bold 72pt Arial',
    fill: "#F5F5DC"
  });
  code.alignTo(credit, Phaser.TOP_CENTER);

  const instructions = this.add.text(0, 0, "Join on your phone at tv.nothing.co.nz\nYour room code is", {
    align: 'center',
    fill: "#F5F5DC"
  });
  instructions.alignTo(code, Phaser.TOP_CENTER);

};

LobbyScene.prototype.listener = function(){
  this.conn.send(JSON.stringify({type: "begin"}));
};

LobbyScene.prototype.update = function(event) {
  var event = this.conn.get();
  if(event != null) {
    this.onMessage(event);
  }
}

LobbyScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "question":
      this.state.start("lie",true,false, this.conn, event.Data.Question);
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
  const question = this.add.text(this.world.centerX, 10, this.question.Text, QUESTION_STYLE);
  question.wordWrapWidth = this.world.width - 100;
  question.anchor.set(0.5, 0);

  this.last = question;
  
  const instructions = this.add.text(this.world.centerX, 10, "Tell your lies now:", {
    font: 'bold 28pt Arial',
    fill: '#F5F5DC',
    align: 'center',
    stroke: '#000000',
    strokeThickness: 6,
    wordWrap: true,
  });
  instructions.wordWrapWidth = this.world.width - 100;
  instructions.alignTo(this.last, Phaser.BOTTOM_CENTER);
  
  this.last = question;
  
  this.timer = game.time.create(true);
  this.timer.add(30000, this.endRound, this);
  this.timer.start();
};

LieScene.prototype.endRound = function(){
  conn.send(JSON.stringify({type: "stop"}));
}

LieScene.prototype.update = function(event) {
  var event = this.conn.get();
  if(event != null) {
    this.onMessage(event);
  }
}

LieScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "collected":
      if(event.Data.Complete) {
        this.endRound();
      }
      const answer = this.add.text(50, 100+this.players.length*50, "${this.players.length+1}. ${event.Data.Player.Name}", {fill: "#ff0000"});
      answer.alignTo(this.last, Phaser.BOTTOM_CENTER);
      this.last = answer;
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
  const question = this.add.text(this.world.centerX, 10, this.question.Text, QUESTION_STYLE);
  question.wordWrapWidth = this.world.width - 100;
  question.anchor.set(0.5, 0);
  var last = question;
  
  for(var i = 0; i < this.question.Answers.length; i++) {
    const answer = this.add.text(50, i*50+100, this.question.Answers[i].Text, {
      font: 'bold 28pt Arial',
      fill: "#ff0000",
      align: 'center',
      stroke: '#000000',
      strokeThickness: 6,
      wordWrap: true,
    });
    answer.alignTo(last, Phaser.BOTTOM_CENTER);
    last = answer;
  }

  this.timer = game.time.create(true);
  this.timer.add(30000, this.endRound, this);
  this.timer.start();
};

VoteScene.prototype.endRound = function(){
  conn.send(JSON.stringify({type: "stop"}));
}

VoteScene.prototype.update = function(event) {
  var event = this.conn.get();
  if(event != null) {
    this.onMessage(event);
  }
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
  
  this.question.Answers.sort(function(a,b) {
    if (!this.offsets){
      return 0;
    }
    
    var a_votes = 0;
    var b_votes = 0;
    for (var i = 0; i < this.offsets.length; i++){
      if (this.offsets[i].Answer.Text == a.Text) { a_votes = this.offsets[i].Answer.Votes.length }
      if (this.offsets[i].Answer.Text == b.Text) { b_votes = this.offsets[i].Answer.Votes.length }
    }

    return b_votes - a_votes;
  }.bind(this));
};

ScoreScene.prototype.create = function() {
  console.log("ScoreScene");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;
  const question = this.add.text(this.world.centerX, 10, this.question.Text, QUESTION_STYLE);
  question.wordWrapWidth = this.world.width - 100;
  question.anchor.set(0.5, 0);


  const goodStyle = {fill: "#ffd800"};
  const badStyle = {fill: "#00a9ff"};

  var last = question;
  
  for (var i = 0; i < this.question.Answers.length; i++) {
    const answer = this.question.Answers[i];
    var votes = 0;
    for(var j = 0; j < this.offsets.length; j++) {
      const offset = this.offsets[j];
      if(offset.Answer.Text === answer.Text) {
        votes = offset.Answer.Votes.length;
        break;
      }
    }
    var answer_text;
    if (answer.Correct) {
      answer_text = this.add.text(game.world.centerX, i*50+100, `${answer.Text} - ${votes} votes`, goodStyle);
    } else {
      answer_text = this.add.text(game.world.centerX, i*50+100, `${answer.Text} - ${votes} votes`, badStyle);
    }
    answer_text.alpha = 0;
    answer_text.anchor.set(0.5);
    answer_text.font = 'Arial Black';
    answer_text.fontSize = 50;
    answer_text.fontWeight = 'bold';
    answer_text.stroke = '#000000';
    answer_text.strokeThickness = 6;
    answer_text.alignTo(last, Phaser.BOTTOM_CENTER);
    
    game.add.tween(answer_text).to( { alpha: 1 }, 1000, Phaser.Easing.Linear.None, true );
    
    last = answer_text;
  }

  
  this.timer = game.time.create(true);
  this.timer.add(5000, this.endRound, this);
  this.timer.start();
};

ScoreScene.prototype.endRound = function(){
  this.state.start("summary", true, false, this.conn, this.points);
}

function SummaryScene() {
  Phaser.State.call(this);
}

SummaryScene.prototype = Object.create(Phaser.State.prototype);

SummaryScene.prototype.preload = function() {
  this.load.image("bg", "css/green-background.jpg");
}

SummaryScene.prototype.init = function(conn, points) {
  this.conn = conn;
  this.points = points;
  this.points.sort(function(a, b){ return b.Total - a.Total });
};

SummaryScene.prototype.create = function() {
  console.log("SummaryScene");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;
  const title = this.add.text(this.world.centerX, 10, "Total Scores:", QUESTION_STYLE);
  title.anchor.set(0.5, 0);
  
  var last = title;
  
  for (var i = 0; i < this.points.length; i++) {
    const player = this.points[i].Player.Name;
    const total = this.points[i].Total;
    var fill = "#ff0000";
    if (i === 0) { fill = "#ffd700" }
    else if (i === 1) { fill = "#c0c0c0" }
    else if (i === 2) { fill = "#cd7f32" }
    
    const score = this.add.text(50, i*50+100, `${player}: ${total} points`, {
      font: 'bold 28pt Arial',
      fill: fill,
      align: 'center',
      stroke: '#000000',
      strokeThickness: 6,
    });
    
    score.alignTo(last, Phaser.BOTTOM_CENTER);
    
    last = score;
  }
  
  this.timer = game.time.create(true);
  this.timer.add(4000, this.endRound, this);
  this.timer.start();
};

SummaryScene.prototype.endRound = function(){
  conn.send(JSON.stringify({type: "next"}));
}

SummaryScene.prototype.update = function(event) {
  var event = this.conn.get();
  if(event != null) {
    this.onMessage(event);
  }
}

SummaryScene.prototype.onMessage = function(event) {
  switch(event.Type) {
    case "question":
      this.state.start("lie",true,false,this.conn, event.Data.Question);
      break;
    case "complete":
      game.state.start("end", true, false, this.conn, this.points);
      break;
    default:
      console.log(event);
  }
};

function EndScene() {
  Phaser.State.call(this);
}

EndScene.prototype = Object.create(Phaser.State.prototype);

EndScene.prototype.preload = function() {
  this.load.image("bg", "css/green-background.jpg");
  this.load.image('rain', 'css/rain.png');
}

EndScene.prototype.init = function(conn, points) {
  this.conn = conn;
  this.points = points;
  this.points.sort(function(a, b){ return b.Total - a.Total });
};

EndScene.prototype.create = function() {
  console.log("EndScreen");

  const bg = this.add.image(0, 0, "bg");
  bg.width = this.world.width;
  bg.height = this.world.height;

  var emitter = game.add.emitter(game.world.centerX, 0, 100);
  emitter.width = game.world.width;
  emitter.makeParticles('rain');
  emitter.minParticleScale = 0.1;
  emitter.maxParticleScale = 5;
  emitter.setYSpeed(300, 500);
  emitter.setXSpeed(0, 0);
  emitter.minRotation = 0;
  emitter.maxRotation = 0;
  emitter.start(false, 1600, 5, 0);

  const title = this.add.text(this.world.centerX, 0, "Winners!", {
    font: "bold 72pt Arial",
    fill: '#F5F5DC',
  });
  title.anchor.set(0.5, 0);

  var max = this.points.length;
  if (max > 3) {
    max = 3;
  }
  var last = title;
  for (var i = 0; i < max; i++) {
    const player = this.points[i].Player.Name;
    const total = this.points[i].Total;
    const row = this.add.text(0, 0, `${player} ${total}`, {
      font: "bold 40pt Arial",
      fill: '#F5F5DC',
    });
    row.alignTo(last, Phaser.BOTTOM_CENTER);
    last = row;
  }
}

function Conn(url) {
  this.conn = new WebSocket(url);
  this.pending = [];
  this.open = false;

  this.conn.onmessage = this.onMessage.bind(this);
  this.conn.onopen = this.onConnection.bind(this);
}

Conn.prototype.onConnection = function() {
  this.open = true;
}

Conn.prototype.onMessage = function(event) {
  this.pending.push(JSON.parse(event.data));
}

Conn.prototype.get = function() {
  return this.pending.shift();
}

Conn.prototype.send = function(event) {
  this.conn.send(event);
}

const conn = new Conn($('body').data('url'));
const game = new Phaser.Game(800,600, Phaser.AUTO, '');
game.state.add("start", new StartScene());
game.state.add("lobby", new LobbyScene());
game.state.add("lie", new LieScene());
game.state.add("vote", new VoteScene());
game.state.add("score", new ScoreScene());
game.state.add("summary", new SummaryScene());
game.state.add("end", new EndScene());
game.state.start("start", true, false, conn);
