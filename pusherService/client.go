package main

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	id       string
	socket   *websocket.Conn
	outbound chan []byte
}

// Constructor de Client
func NewClient(hub *Hub, socket *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		socket:   socket,
		outbound: make(chan []byte),
	}
}

// Funcion que se encargara de transmitir post que se crean en tiempo real
func (c *Client) Write() {
	for {
		select {
		case message, ok := <-c.outbound:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{}) // Se envia un mesaje al cliente por que la coneccion se cerro
				return
			}
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}
