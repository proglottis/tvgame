package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

func generateCode() string {
	var buf [2]byte
	var msg [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	hex.Encode(msg[:], buf[:])
	return string(msg[:])
}

type RoomMessage struct {
	Code string
}

type Room struct {
	Code      string
	Game      *Game
	Host      *websocket.Conn
	LobbyDone chan<- string
	Join      chan *websocket.Conn
	Players   []*websocket.Conn
}

func NewRoom(repo *QuestionRepo, host *websocket.Conn) *Room {
	return &Room{
		Code: generateCode(),
		Game: NewGame(repo),
		Host: host,
		Join: make(chan *websocket.Conn),
	}
}

func (r *Room) Run() {
	r.Host.WriteJSON(RoomMessage{Code: r.Code})
	for {
		select {
		case conn, ok := <-r.Join:
			if !ok {
				r.Join = nil
				continue
			}
			log.Printf("Room: Join request")
			r.Players = append(r.Players, conn)
		}
	}
}

type LobbyMessage struct {
	Type string
	Code string
}

type Lobby struct {
	Repo *QuestionRepo
	done <-chan string

	mu    sync.RWMutex
	rooms map[string]chan<- *websocket.Conn
}

func NewLobby(repo *QuestionRepo) *Lobby {
	return &Lobby{
		Repo:  repo,
		done:  make(chan string),
		rooms: make(map[string]chan<- *websocket.Conn),
	}
}

func (l *Lobby) Run() {
	for code := range l.done {
		l.detatch(code)
	}
}

func (l *Lobby) create(conn *websocket.Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	room := NewRoom(l.Repo, conn)
	go room.Run()
	l.rooms[room.Code] = room.Join
}

func (l *Lobby) detatch(code string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	room, ok := l.rooms[code]
	if !ok {
		log.Printf("Lobby: No such room to detach: %s", code)
		return
	}
	close(room)
	delete(l.rooms, code)
	log.Printf("Lobby: Detached room: %s", code)
}

func (l *Lobby) join(conn *websocket.Conn, code string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	room, ok := l.rooms[code]
	if !ok {
		log.Printf("Lobby: No such room to join: %s", code)
		conn.Close()
		return
	}
	room <- conn
}

func (l *Lobby) Handle(conn *websocket.Conn) {
	// runs within the HTTP handler go routine
	var msg LobbyMessage
	if err := conn.ReadJSON(&msg); err != nil {
		log.Printf("Lobby: %s", err)
		conn.Close()
	}
	switch msg.Type {
	case "create":
		l.create(conn)
	case "join":
		l.join(conn, msg.Code)
	default:
		log.Printf("Lobby: Unknown message type: %s", msg.Type)
		conn.Close()
	}
}
