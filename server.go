package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
)

type JSONConn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}

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
	Type   string
	Code   string      `json:",omitempty"`
	Player *RoomPlayer `json:",omitempty"`
}

type RoomPlayer struct {
	Name  string
	Score int
	Conn  JSONConn `json:"-"`
}

type Room struct {
	Code      string
	Game      *Game
	Host      JSONConn
	LobbyDone chan<- string
	Join      chan *RoomPlayer
	Players   []*RoomPlayer
}

func NewRoom(repo *QuestionRepo, host JSONConn) *Room {
	return &Room{
		Code: generateCode(),
		Game: NewGame(repo),
		Host: host,
		Join: make(chan *RoomPlayer),
	}
}

func (r *Room) Run() {
	r.Host.WriteJSON(RoomMessage{Type: "start", Code: r.Code})
	for {
		select {
		case player, ok := <-r.Join:
			if !ok {
				r.Join = nil
				continue
			}
			log.Printf("Room: Join request")
			r.Players = append(r.Players, player)
			r.Host.WriteJSON(RoomMessage{Type: "join", Player: player})
		}
	}
}

type LobbyMessage struct {
	Type string
	Name string
	Code string
}

type Lobby struct {
	Repo *QuestionRepo
	done <-chan string

	mu    sync.RWMutex
	rooms map[string]chan<- *RoomPlayer
}

func NewLobby(repo *QuestionRepo) *Lobby {
	return &Lobby{
		Repo:  repo,
		done:  make(chan string),
		rooms: make(map[string]chan<- *RoomPlayer),
	}
}

func (l *Lobby) Run() {
	for code := range l.done {
		l.detatch(code)
	}
}

func (l *Lobby) create(conn JSONConn) {
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

func (l *Lobby) join(conn JSONConn, name, code string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	room, ok := l.rooms[code]
	if !ok {
		log.Printf("Lobby: No such room to join: %s", code)
		conn.Close()
		return
	}
	room <- &RoomPlayer{Name: name, Conn: conn}
}

func (l *Lobby) Handle(conn JSONConn) {
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
		l.join(conn, msg.Name, msg.Code)
	default:
		log.Printf("Lobby: Unknown message type: %s", msg.Type)
		conn.Close()
	}
}
