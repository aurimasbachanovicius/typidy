package main

import (
	"flag"
	"github.com/3auris/typidy/chat"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func main() {
	apiPort := flag.String("apiPort", "8080", "Port of opening API / WS / Static")

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	hub := chat.NewHub()
	go hub.Run()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/chat-socket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		client := chat.NewClient(hub, conn)
		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	})
	err := http.ListenAndServe(":"+*apiPort, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
