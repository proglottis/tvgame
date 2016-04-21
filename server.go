package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateCode(n int) string {
	msg := make([]byte, n)
	for i := range msg {
		msg[i] = letters[rand.Intn(len(letters))]
	}
	return CleanText(string(msg))
}

type PlayerText struct {
	Text   string
	Player *RoomPlayer `json:"-"`
}

type RoomHost struct {
	Conn *Conn
}

type joinedMessage struct {
	Player Player
}

func (h *RoomHost) Joined(player Player) {
	var err error
	msg := ConnMessage{Type: "joined"}
	msg.Data, err = json.Marshal(joinedMessage{Player: player})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		close(h.Conn.Send)
		return
	}
	h.Conn.Send <- msg
}

type questionMessage struct {
	Question *Question
}

func (h *RoomHost) Question(question *Question) {
	var err error
	msg := ConnMessage{Type: "question"}
	msg.Data, err = json.Marshal(questionMessage{Question: question})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		close(h.Conn.Send)
		return
	}
	h.Conn.Send <- msg
}

func (h *RoomHost) Vote(question *Question) {
	var err error
	msg := ConnMessage{Type: "vote"}
	msg.Data, err = json.Marshal(questionMessage{Question: question})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		close(h.Conn.Send)
		return
	}
	h.Conn.Send <- msg
}

type collectedMessage struct {
	Player   Player
	Complete bool
}

func (h *RoomHost) Collected(player Player, complete bool) {
	var err error
	msg := ConnMessage{Type: "collected"}
	msg.Data, err = json.Marshal(collectedMessage{Player: player, Complete: complete})
	if err != nil {
		log.Printf("RoomHost: %s", err)
		close(h.Conn.Send)
		return
	}
	h.Conn.Send <- msg
}

func (h *RoomHost) Results(results *ResultSet) {
	var err error
	msg := ConnMessage{Type: "results"}
	msg.Data, err = json.Marshal(results)
	if err != nil {
		log.Printf("RoomHost: %s", err)
		close(h.Conn.Send)
		return
	}
	h.Conn.Send <- msg
}

type RoomPlayer struct {
	ID   string
	Name string
	Conn *Conn `json:"-"`
}

func (p *RoomPlayer) SendError(text string) {
	var err error
	msg := ConnMessage{Type: "error"}
	msg.Data, err = json.Marshal(errorMessage{Text: text})
	if err != nil {
		log.Printf("RoomPlayer: %s", err)
		return
	}
	p.Conn.Send <- msg
}

func (p *RoomPlayer) SendAck() {
	msg := ConnMessage{Type: "ok"}
	p.Conn.Send <- msg
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
	p.Conn.Send <- msg
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
	p.Conn.Send <- msg
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
		Game:           NewGame(repo, h),
		Host:           host,
		Join:           make(chan *RoomPlayer),
		playerMessages: make(chan PlayerText),
	}
}

type roomMessage struct {
	Code string
}

func (r *Room) Run() {
	var err error
	msg := ConnMessage{Type: "create"}
	msg.Data, err = json.Marshal(roomMessage{Code: r.Code})
	if err != nil {
		panic(err)
	}
	r.Host.Send <- msg

	for {
		select {
		case msg, ok := <-r.Host.Recv:
			if !ok {
				// TODO: host quit
				continue
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
		case msg, ok := <-r.playerMessages:
			if !ok {
				r.playerMessages = nil
				continue
			}
			if err := r.Game.Collect(msg.Player, msg.Text); err != nil {
				msg.Player.SendError(err.Error())
				continue
			}
			msg.Player.SendAck()
		case player, ok := <-r.Join:
			if !ok {
				r.Join = nil
				continue
			}
			player.Name = CleanText(player.Name)
			if len(player.Name) < 1 {
				player.SendError("Name is required")
				continue
			}
			nameTaken := false
			for other := range r.Game.Players {
				if other.(*RoomPlayer).Name == player.Name {
					nameTaken = true
					break
				}
			}
			if nameTaken {
				player.SendError("Name is taken")
				continue
			}
			if err := r.Game.AddPlayer(player); err != nil {
				player.SendError(err.Error())
				continue
			}
			player.SendAck()
			go func() {
				for msg := range player.Conn.Recv {
					var text PlayerText
					if err := json.Unmarshal(msg.Data, &text); err != nil {
						panic(err)
					}
					text.Player = player
					r.playerMessages <- text
				}
			}()
		}
	}
}

type lobbyMessage struct {
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
	for {
		room.Code = generateCode(5)
		if _, ok := l.rooms[room.Code]; !ok {
			break
		}
	}
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
	code = CleanText(code)
	room, ok := l.rooms[code]
	if !ok {
		log.Printf("Lobby: No such room to join: %s", code)
		close(conn.Send)
		return
	}
	room <- &RoomPlayer{ID: generateCode(10), Name: name, Conn: conn}
}

func (l *Lobby) Handle(conn *Conn) {
	// runs within the HTTP handler go routine
	for msg := range conn.Recv {
		switch msg.Type {
		case "create":
			l.create(conn)
			return
		case "join":
			var lobby lobbyMessage
			if err := json.Unmarshal(msg.Data, &lobby); err != nil {
				var e error
				msg := ConnMessage{Type: "error"}
				msg.Data, e = json.Marshal(errorMessage{Text: err.Error()})
				if e != nil {
					log.Printf("Lobby: %s", e)
					close(conn.Send)
					return
				}
				conn.Send <- msg
				continue
			}
			l.join(conn, lobby.Name, lobby.Code)
			return
		}
	}
}

type errorMessage struct {
	Text string
}
