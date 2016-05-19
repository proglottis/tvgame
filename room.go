package main

import (
	"errors"
	"sync"

	"github.com/proglottis/tvgame/game"
)

type Room struct {
	Code string

	mu   sync.Mutex
	game *game.Game
}

func NewRoom(repo *game.QuestionRepo, host *Conn) *Room {
	return &Room{game: game.New(repo, &RoomHost{Conn: host})}
}

type roomMessage struct {
	Code string
}

func (r *Room) Host() *RoomHost {
	return r.game.Host.(*RoomHost)
}

func (r *Room) AddPlayer(player *RoomPlayer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	player.Name = game.CleanText(player.Name)
	if len(player.Name) < 1 {
		return errors.New("Name is too short (min 1)")
	}
	if len(player.Name) > 10 {
		return errors.New("Name is too long (max 10)")
	}
	for other := range r.game.Players {
		if other.(*RoomPlayer).Name == player.Name {
			return errors.New("Name is taken")
		}
	}
	if err := r.game.AddPlayer(player); err != nil {
		return err
	}
	return nil
}

func (r *Room) Begin() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.game.Begin()
}

func (r *Room) Vote() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.game.Vote()
}

func (r *Room) Collect(player *RoomPlayer, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.game.Collect(player, text)
}

func (r *Room) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.game.Stop()
}

func (r *Room) Next() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.game.Next()
}
