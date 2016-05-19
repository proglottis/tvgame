package main

import (
	"encoding/json"
	"log"

	"github.com/proglottis/tvgame/game"
	"golang.org/x/net/context"
)

type RoomPlayer struct {
	ID   string
	Name string
	Conn *Conn `json:"-"`
}

type errorMessage struct {
	Text string
}

func (p *RoomPlayer) SendError(text string) {
	var err error
	msg := ConnMessage{Type: "error"}
	msg.Data, err = json.Marshal(errorMessage{Text: text})
	if err != nil {
		log.Printf("RoomPlayer: %s", err)
		return
	}
	p.Conn.Write(&msg)
}

func (p *RoomPlayer) SendAck() {
	msg := ConnMessage{Type: "ok"}
	p.Conn.Write(&msg)
}

type requestAnswerMessage struct {
	Text string
}

func (p *RoomPlayer) RequestAnswer(text string) {
	var err error
	msg := ConnMessage{Type: "answer"}
	msg.Data, err = json.Marshal(requestAnswerMessage{Text: text})
	if err != nil {
		log.Printf("RoomPlayer: %s", err)
		return
	}
	p.Conn.Write(&msg)
}

type playerText struct {
	Text string
}

func (p *RoomPlayer) Run(ctx context.Context, room *Room) error {
	defer p.Conn.Close()
	for {
		var msg ConnMessage
		if err := p.Conn.Read(&msg); err != nil {
			return err
		}
		var text playerText
		if err := json.Unmarshal(msg.Data, &text); err != nil {
			return err
		}
		if err := room.Collect(p, text.Text); err != nil {
			p.SendError(err.Error())
		} else {
			p.SendAck()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

type requestVoteMessage struct {
	Text    string
	Answers []string
}

func (p *RoomPlayer) RequestVote(text string, answers []string) {
	var err error
	msg := ConnMessage{Type: "vote"}
	msg.Data, err = json.Marshal(requestVoteMessage{Text: text, Answers: answers})
	if err != nil {
		log.Printf("RoomPlayer: %s", err)
		return
	}
	p.Conn.Write(&msg)
}

func (p *RoomPlayer) Complete(game *game.Game) {
	p.Conn.Write(&ConnMessage{Type: "complete"})
}
