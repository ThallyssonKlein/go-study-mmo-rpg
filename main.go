package main

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/websocket"
    "sync"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type Response struct {
    Message string `json:"message"`
}

type Room struct {
    connections map[*websocket.Conn]bool
    broadcast   chan []byte
    sync.Mutex
}

var rooms = make(map[string]*Room)
var roomsMutex = &sync.Mutex{}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    response := Response{Message: "Hello, World!"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    roomID := r.URL.Query().Get("room")
    if roomID == "" {
        http.Error(w, "Room ID is required", http.StatusBadRequest)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
        return
    }

    roomsMutex.Lock()
    room, ok := rooms[roomID]
    if !ok {
        room = &Room{
            connections: make(map[*websocket.Conn]bool),
            broadcast:   make(chan []byte),
        }
        rooms[roomID] = room
        go room.run()
    }
    roomsMutex.Unlock()

    room.Lock()
    room.connections[conn] = true
    room.Unlock()

    defer func() {
        room.Lock()
        delete(room.connections, conn)
        room.Unlock()
        conn.Close()
    }()
}

func (room *Room) run() {
    for {
        message := <-room.broadcast
        room.Lock()
        for conn := range room.connections {
            if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
                conn.Close()
                delete(room.connections, conn)
            }
        }
        room.Unlock()
    }
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    http.HandleFunc("/ws", wsHandler)
    http.ListenAndServe(":8080", nil)
}