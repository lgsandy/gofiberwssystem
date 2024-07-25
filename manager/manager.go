package manager

import (
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WebSocketManager struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

func (manager *WebSocketManager) Run() {
	for {
		select {
		case conn := <-manager.Register:
			manager.mu.Lock()
			manager.Clients[conn] = true
			manager.mu.Unlock()
		case conn := <-manager.Unregister:
			manager.mu.Lock()
			if _, ok := manager.Clients[conn]; ok {
				delete(manager.Clients, conn)
				conn.Close()
			}
			manager.mu.Unlock()
		case message := <-manager.Broadcast:
			manager.mu.Lock()
			for conn := range manager.Clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					conn.Close()
					delete(manager.Clients, conn)
				}
			}
			manager.mu.Unlock()
		}
	}
}

func (manager *WebSocketManager) SendBroadcast(message []byte) {
	manager.Broadcast <- message
}
