"use strict";

$(function() {
  var $join_form   = $('#join'),
      $error       = $('.error'),
      $waiting     = $('.waiting'),
      $question    = $('.question'),
      $answer_form = $('#answer-form'),
      $answers     = $question.find('.answers');

  function appendLog(msg) {
    console.log(msg);
  }

  var conn, state = stateJoined;

  function waiting() {
    $waiting.show();
    $error.hide();
    $question.hide();
    $answer_form.hide();
    $answers.hide();
  }

  function stopWaiting() {
    waiting();
    $waiting.hide();
  }

  function stateJoined(action, data) {
    switch(action) {
      case "ok":
        waiting();
        return stateWaiting;
      case "error":
        $waiting.hide();
        $error.show().text(data["Data"]["Text"]);
        $join_form.show()
        conn.close()
        break;
      default:
        appendLog("stateJoined: " + action + ": " + JSON.stringify(data));
    }
    return stateJoined;
  }

  function stateWaiting(action, data) {
    switch(action) {
      case "answer":
        stopWaiting();
        $answer_form.show();
        $question.show().find('h2').text(data["Data"]["Text"]);
        return stateAnswering;
      case "vote":
        // {"Type":"vote","Data":{"Text":"A phlebotomist extracts what from the human body?","Answers":["BLOOD"]}}
        stopWaiting();
        $question.show().find('h2').text(data["Data"]["Text"]);
        $answers.show().html($.map(data["Data"]["Answers"], function(text){ return $('<button>').text(text).wrap('<li>'); }));
        return stateVoting;
      case "results":
        waiting();
        break;
      default:
        appendLog("stateWaiting: " + action + ": " + JSON.stringify(data));
    }
    return stateWaiting;
  }

  function stateAnswering(action, data) {
    switch(action) {
      case "ok":
        waiting();
        return stateWaiting;
      case "error":
        $error.show().text(data["Data"]["Text"]);
        break;
      default:
        appendLog("stateAnswering: " + action + ": " + JSON.stringify(data));
        return stateWaiting(action, data);
    }
    return stateAnswering;
  }

  function stateVoting(action, data) {
    switch(action) {
      case "ok":
        waiting();
        return stateWaiting;
      case "error":
        $error.show().text(data["Data"]["Text"]);
        break;
      default:
        appendLog("stateVoting: " + action + ": " + JSON.stringify(data));
        return stateWaiting(action, data);
    }
    return stateVoting;
  }

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
      state = state(action, data);
    };

    conn.onclose = function(event) {
      stopWaiting();
      $error.show().text("Connection closed");
      $join_form.show();
      state = stateJoined;
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
    $('textarea[name=answer]').val('');
  });

  $answers.click(function (event) {
    conn.send(JSON.stringify({Type: 'vote', Data: {
      Text: event.target.innerText
    }}));
  });
});
