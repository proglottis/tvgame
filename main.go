package main

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/proglottis/tvgame/game"
	"golang.org/x/crypto/acme/autocert"
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

func websocketURL(r *http.Request) template.URL {
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	ws := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "ws",
	}
	return template.URL(ws.String())
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	portTLS := os.Getenv("PORT_TLS")
	if portTLS == "" {
		portTLS = "8081"
	}
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
		index.Execute(w, websocketURL(r))
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
		host.Execute(w, websocketURL(r))
	}))

	http.HandleFunc("/ws", withLog(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade: %s", err)
			return
		}
		if err := server.Handle(r.Context(), NewConn(r.Context(), conn)); err != nil {
			log.Printf("Server: %s", err)
			return
		}
	}))

	go func() {
		addr := ":" + portTLS
		log.Printf("Starting TLS on %s", addr)
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("cache"),
			HostPolicy: autocert.HostWhitelist("tv.nothing.co.nz"),
		}
		s := &http.Server{
			Addr:      addr,
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		if err := s.ListenAndServeTLS("", ""); err != nil {
			log.Fatal("ListenAndServeTLS:", err)
		}
	}()

	addr := ":" + port
	log.Printf("Starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
