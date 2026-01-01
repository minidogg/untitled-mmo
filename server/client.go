package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientID = string
type Client struct {
	ID     ClientID
	Socket *websocket.Conn
	Entity *Entity
}

type ClientMap map[ClientID]*Client
type ClientStore struct {
	ClientMap ClientMap
}

var Clients = ClientStore{
	ClientMap: make(ClientMap),
}

func (cs *ClientStore) GenerateClientFromSocket(socket *websocket.Conn) *Client {
	c := &Client{
		ID:     ClientID(uuid.NewString()),
		Socket: socket,
	}
	cs.ClientMap[c.ID] = c

	return c
}

func (cs *ClientStore) RemoveClient(id ClientID) {
	client, exists := cs.ClientMap[id]
	if !exists {
		return
	}

	if client.Socket != nil {
		_ = client.Socket.Close()
	}

	if client.Entity != nil {
		client.Entity.Client = nil
		client.Entity.Remove = true
	}

	delete(cs.ClientMap, id)
}

func (cl *ClientStore) SendPacketToID(id ClientID, packet Packet) bool {
	client, found := cl.ClientMap[id]
	if found {
		client.Socket.WriteJSON(packet)
	}
	return found
}
