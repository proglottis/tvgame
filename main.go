package main

import (
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var (
	index = template.Must(template.ParseFiles("index.html"))
	host  = template.Must(template.ParseFiles("host.html"))
)

func withLog(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		fn(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	csv, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	repo, err := NewQuestionRepo(csv)
	if err != nil {
		log.Fatal(err)
	}
	lobby := NewLobby(repo)
	go lobby.Run()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.HandleFunc("/", withLog(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", 404)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		index.Execute(w, r.Host)
	}))

	http.HandleFunc("/host", withLog(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/host" {
			http.Error(w, "Not found", 404)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		host.Execute(w, r.Host)
	}))

	http.HandleFunc("/ws", withLog(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade:", err)
			return
		}
		lobby.Handle(NewConn(conn))
	}))

	addr := ":" + port
	log.Printf("Starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
