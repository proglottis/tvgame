package main

import (
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/proglottis/tvgame/game"
	"golang.org/x/net/context"
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
	ctx := context.Background()
	csv, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("Load CSV: %s", err)
	}
	repo, err := game.NewQuestionRepo(csv)
	if err != nil {
		log.Fatal("Parse CSV: %s", err)
	}
	server := NewServer(repo)
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
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade: %s", err)
			return
		}
		if err := server.Handle(ctx, NewConn(ctx, conn)); err != nil {
			log.Printf("Server: %s", err)
			return
		}
	}))

	addr := ":" + port
	log.Printf("Starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
