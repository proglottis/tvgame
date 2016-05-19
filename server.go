package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/proglottis/tvgame/game"
	"golang.org/x/net/context"
)

type JoinRequest struct {
	Name string
	Code string
}

type Server struct {
	Repo *game.QuestionRepo

	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewServer(repo *game.QuestionRepo) *Server {
	return &Server{
		Repo:  repo,
		rooms: make(map[string]*Room),
	}
}

func (s *Server) Handle(ctx context.Context, conn *Conn) error {
	var msg ConnMessage
	if err := conn.Read(&msg); err != nil {
		return err
	}
	switch msg.Type {
	case "create":
		return s.CreateRoom(ctx, conn)
	case "join":
		var req JoinRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return err
		}
		return s.JoinRoom(ctx, conn, &req)
	default:
		return fmt.Errorf("Unknown message: %s", msg.Type)
	}
}

func (s *Server) detachRoom(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("Server: Detaching room: %s", code)
	delete(s.rooms, code)
}

func (s *Server) createRoom(conn *Conn) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	room := NewRoom(s.Repo, conn)
	for {
		room.Code = generateCode(4)
		if _, ok := s.rooms[room.Code]; !ok {
			break
		}
	}
	s.rooms[room.Code] = room
	return room
}

func (s *Server) CreateRoom(ctx context.Context, conn *Conn) error {
	room := s.createRoom(conn)
	log.Printf("Server: room %s created", room.Code)
	detach := func() {
		s.detachRoom(room.Code)
	}
	return room.Host().Run(ctx, room, detach)
}

func (s *Server) JoinRoom(ctx context.Context, conn *Conn, msg *JoinRequest) error {
	msg.Code = game.CleanText(msg.Code)
	s.mu.RLock()
	room, ok := s.rooms[msg.Code]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("No such room: %s", msg.Code)
	}
	player := &RoomPlayer{ID: generateCode(10), Name: game.CleanText(msg.Name), Conn: conn}
	if err := room.AddPlayer(player); err != nil {
		player.SendError(err.Error())
		return err
	}
	log.Printf("Server: joined player to room %s", msg.Code)
	player.SendAck()
	return player.Run(ctx, room)
}

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateCode(n int) string {
	msg := make([]byte, n)
	for i := range msg {
		msg[i] = letters[rand.Intn(len(letters))]
	}
	return game.CleanText(string(msg))
}
