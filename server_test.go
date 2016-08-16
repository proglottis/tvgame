package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"github.com/proglottis/tvgame/game"
)

func TestServer_joining(t *testing.T) {
	csv, err := os.Open("testdata/quiz.csv")
	if err != nil {
		t.Fatal(err)
	}
	repo, err := game.NewQuestionRepo(csv)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(repo)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ctx := context.Background()
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		if err := server.Handle(ctx, NewConn(ctx, conn)); err != nil {
			return
		}
	}))
	defer httpServer.Close()
	serverURL, err := url.Parse(httpServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	serverURL.Scheme = "ws"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			host, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := host.WriteMessage(websocket.TextMessage, []byte(`{"Type":"create"}`)); err != nil {
				t.Fatal(err)
			}
			_, msg, err := host.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}
			doc, err := simplejson.NewJson(msg)
			if err != nil {
				t.Fatal(err)
			}
			code, err := doc.GetPath("Data", "Code").String()
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 8; i++ {
				go func(i int) {
					player, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
					if err != nil {
						return
					}
					defer player.Close()
					if err := player.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"Type":"join","Data":{"Code":"%s","Name":"bob%d"}}`, code, i+1))); err != nil {
						return
					}
					_, _, err = player.ReadMessage()
					if err != nil {
						return
					}
				}(i)
			}

			for i := 0; i < 8; i++ {
				_, _, err = host.ReadMessage()
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := host.WriteMessage(websocket.TextMessage, []byte(`{"Type":"begin"}`)); err != nil {
				t.Fatal(err)
			}
			_, _, err = host.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}
			host.Close()
			wg.Done()
		}()
	}
	wg.Wait()
}
