package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
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

type PlayerText struct {
	Text   string
	Player *RoomPlayer `json:"-"`
}

type RoomMessage struct {
	Code   string `json:",omitempty"`
	Player Player `json:",omitempty"`
}

type RoomHost struct {
	Conn *Conn
}

func (h *RoomHost) Joined(player Player) {
	var err error
	msg := ConnMessage{Type: "joined"}
	msg.Data, err = json.Marshal(&RoomMessage{Player: player})
	if err != nil {
		panic(err)
	}
	h.Conn.Send <- msg
}

func (h *RoomHost) Question(question *Question) {
}

type RoomPlayer struct {
	Name string
	Conn *Conn `json:"-"`
}

func (p *RoomPlayer) RequestAnswer(question string) {
}

func (p *RoomPlayer) RequestVote(text string, answers []string) {
}

type Room struct {
	Code           string
	Game           *Game
	Host           *Conn
	LobbyDone      chan<- string
	Join           chan *RoomPlayer
	playerMessages chan PlayerText
}

func NewRoom(repo *QuestionRepo, host *Conn) *Room {
	h := &RoomHost{Conn: host}
	return &Room{
		Code: generateCode(),
		Game: NewGame(repo, h),
		Host: host,
		Join: make(chan *RoomPlayer),
	}
}

func (r *Room) Run() {
	var err error
	msg := ConnMessage{Type: "start"}
	msg.Data, err = json.Marshal(RoomMessage{Code: r.Code})
	if err != nil {
		panic(err)
	}
	r.Host.Send <- msg

	var playerMessages chan PlayerText
	for {
		select {
		case msg, ok := <-r.Host.Recv:
			if !ok {
				// TODO: host quit
			}
			switch msg.Type {
			case "begin":
				r.Game.Begin()
			case "next":
				r.Game.Next()
			case "vote":
				r.Game.Vote()
			case "stop":
				r.Game.Stop()
			}
		case msg, ok := <-playerMessages:
			if !ok {
				playerMessages = nil
				continue
			}
			if err := r.Game.Collect(msg.Player, msg.Text); err != nil {
				// TODO: send err to player
			}
		case player, ok := <-r.Join:
			if !ok {
				r.Join = nil
				continue
			}
			r.Game.AddPlayer(player)
			go func() {
				for msg := range player.Conn.Recv {
					var text PlayerText
					if err := json.Unmarshal(msg.Data, &text); err != nil {
						panic(err)
					}
					r.playerMessages <- text
				}
			}()
		}
	}
}

type LobbyMessage struct {
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

func (l *Lobby) create(conn *Conn) {
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

func (l *Lobby) join(conn *Conn, name, code string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	room, ok := l.rooms[code]
	if !ok {
		log.Printf("Lobby: No such room to join: %s", code)
		close(conn.Send)
		return
	}
	room <- &RoomPlayer{Name: name, Conn: conn}
}

func (l *Lobby) Handle(conn *Conn) {
	// runs within the HTTP handler go routine
	msg := <-conn.Recv
	switch msg.Type {
	case "create":
		l.create(conn)
	case "join":
		var lobby LobbyMessage
		if err := json.Unmarshal(msg.Data, &lobby); err != nil {
			log.Printf("Lobby: Unmarshal: %s", err)
			close(conn.Send)
			return
		}
		l.join(conn, lobby.Name, lobby.Code)
	default:
		log.Printf("Lobby: Unknown message type: %s", msg.Type)
		close(conn.Send)
	}
}
