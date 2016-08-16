package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/proglottis/tvgame/game"
)

type RoomHost struct {
	Conn *Conn
}

type joinedMessage struct {
	Player game.Player
}

func (h *RoomHost) Joined(player game.Player) {
	var err error
	msg := ConnMessage{Type: "joined"}
	msg.Data, err = json.Marshal(joinedMessage{Player: player})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		h.Conn.Close()
		return
	}
	h.Conn.Write(&msg)
}

type questionMessage struct {
	Question *game.Question
}

func (h *RoomHost) Question(question *game.Question) {
	var err error
	msg := ConnMessage{Type: "question"}
	msg.Data, err = json.Marshal(questionMessage{Question: question})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		h.Conn.Close()
		return
	}
	h.Conn.Write(&msg)
}

func (h *RoomHost) Vote(question *game.Question) {
	var err error
	msg := ConnMessage{Type: "vote"}
	msg.Data, err = json.Marshal(questionMessage{Question: question})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		h.Conn.Close()
		return
	}
	h.Conn.Write(&msg)
}

type collectedMessage struct {
	Player   game.Player
	Complete bool
}

func (h *RoomHost) Collected(player game.Player, complete bool) {
	var err error
	msg := ConnMessage{Type: "collected"}
	msg.Data, err = json.Marshal(collectedMessage{Player: player, Complete: complete})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		h.Conn.Close()
		return
	}
	h.Conn.Write(&msg)
}

type resultPoints struct {
	Player game.Player
	Total  int
}

type resultOffsets struct {
	Answer  *game.Answer
	Offsets []game.Result
}

type resultsMessage struct {
	Points  []resultPoints
	Offsets []resultOffsets `json:",omitempty"`
}

func (h *RoomHost) Results(game *game.Game, results game.ResultSet) {
	var err error
	data := &resultsMessage{}
	for player, total := range game.Players {
		data.Points = append(data.Points, resultPoints{Player: player, Total: total})
	}
	for answer, result := range results {
		data.Offsets = append(data.Offsets, resultOffsets{Answer: answer, Offsets: result})
	}
	msg := ConnMessage{Type: "results"}
	msg.Data, err = json.Marshal(data)
	if err != nil {
		log.Printf("RoomHost: %s", err)
		h.Conn.Close()
		return
	}
	h.Conn.Write(&msg)
}

func (h *RoomHost) Complete(game *game.Game) {
	var err error
	data := &resultsMessage{}
	for player, total := range game.Players {
		data.Points = append(data.Points, resultPoints{Player: player, Total: total})
	}
	msg := ConnMessage{Type: "complete"}
	msg.Data, err = json.Marshal(data)
	if err != nil {
		log.Printf("RoomHost: %s", err)
		h.Conn.Close()
		return
	}
	h.Conn.Write(&msg)
}

type roomMessage struct {
	Code string
}

func (h *RoomHost) Run(ctx context.Context, room *Room, detach func()) error {
	var err error
	msg := ConnMessage{Type: "create"}
	msg.Data, err = json.Marshal(roomMessage{Code: room.Code})
	if err != nil {
		return err
	}
	if err := h.Conn.Write(&msg); err != nil {
		return err
	}
	for {
		var msg ConnMessage
		if err := h.Conn.Read(&msg); err != nil {
			return err
		}
		switch msg.Type {
		case "begin":
			detach()
			room.Begin()
		case "next":
			room.Next()
		case "vote":
			room.Vote()
		case "stop":
			room.Stop()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}
