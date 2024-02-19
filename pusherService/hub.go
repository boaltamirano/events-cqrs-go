package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    []*Client
	register   chan *Client
	unregister chan *Client
	mutex      *sync.Mutex
}

// Contructor de Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make([]*Client, 0), //Creamos un nuevo slice para clients de longitud 0
		register:   make(chan *Client), // Creamos un canal para register
		unregister: make(chan *Client),
		mutex:      &sync.Mutex{},
	}
}

// Definimos la ruta que va a manejar los websockets
// funcion deltro del Hub como un metodo de nombre HandlerWebSocket
func (hub *Hub) HandlerWebSocket(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error en HandlerWebSocket -> ", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	client := NewClient(hub, socket)
	hub.register <- client // al hub le registramos el cliente

	// Activamos go routina que se encarge de escribir los mensajes al web socket
	go client.Write()
}

// Reciver function que le permite correr o ejecutarce
func (hub *Hub) Run() {
	for {
		select { // Multiplexacion de los channels ya registrados
		case client := <-hub.register:
			hub.onConnect(client)
		case client := <-hub.unregister:
			hub.onDisconnect(client)
		}
	}
}

func (hub *Hub) onConnect(client *Client) {
	log.Println("Client Connected", client.socket.RemoteAddr()) // imprimimos que un cliente se conecto con la impresion de la direccion de conexion
	hub.mutex.Lock()                                            // Bloqueamos el programa para evitar condicion de carrera
	defer hub.mutex.Unlock()
	client.id = client.socket.RemoteAddr().String() // Le asignamos un id a los clientes registrados correctamente
	hub.clients = append(hub.clients, client)       // agregamos al cliente a los clientes del hub
}

func (hub *Hub) onDisconnect(client *Client) {
	log.Println("Client Connected", client.socket.RemoteAddr()) // imprimimos que un cliente se conecto con la impresion de la direccion de conexion
	client.socket.Close()                                       // Cerramos la conexion del cliente que se desconecto
	hub.mutex.Lock()                                            // Bloqueamos el programa para evitar condicion de carrera
	defer hub.mutex.Unlock()
	i := -1
	for j, c := range hub.clients { // iterar a traves de los clientes para buscar el cliente que se desconecto
		if c.id == client.id {
			i = j
		}
	}

	//proceso de eliminar clientes
	copy(hub.clients[i:], hub.clients[i+1:])
	hub.clients[len(hub.clients)-1] = nil
	hub.clients = hub.clients[:len(hub.clients)-1]
}

func (hub *Hub) Brodcast(message any, ignore *Client) {
	data, _ := json.Marshal(message)
	for _, client := range hub.clients {
		if client != ignore {
			client.outbound <- data
		}
	}
}
