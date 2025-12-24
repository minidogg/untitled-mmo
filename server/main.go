// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	ip := r.RemoteAddr
	ip_header := r.Header.Get("cf-connecting-ip")
	if ip_header != "" {
		ip = ip_header
	}
	handshaked := false
	for {
		var msg Packet
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}
		fmt.Println(msg)
		switch msg.Cmd {
		case "handshake":
			if handshaked == true {
				break
			}
			handshaked = true
			c.WriteJSON(Packet{
				Cmd: "server-info",
				Val: "untitled-mmo-cl-0.1.0",
			})
			c.WriteJSON(Packet{
				Cmd: "client-ip",
				Val: ip,
			})
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse((`<h1>Server Shard</h1>`)))
