package main

import (
	"flag"
	"html/template"
	"log"
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
	flag.Parse()
	log.Printf("TV Game")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	csv, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	_, err = NewQuestionRepo(csv)
	if err != nil {
		log.Fatal(err)
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
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
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade:", err)
			return
		}
		// TODO: handle connection
	}))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
