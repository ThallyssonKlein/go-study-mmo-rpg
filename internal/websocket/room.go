package websocket

import (
    "net/http"
    "sync"
    "github.com/gorilla/websocket"
	"encoding/json"
)

type Room struct {
    connections map[*websocket.Conn]bool
    broadcast   chan []byte
    sync.Mutex
    characters  map[*websocket.Conn]Character
}

var rooms = make(map[string]*Room)
var roomsMutex = &sync.Mutex{}

type Action string
type Character string

// warrior, archer, wizard
const (
	Warrior Character = "warrior"
	Archer  Character = "archer"
	Wizard  Character = "wizard"
)
const (
    Attack  Action = "attack"
    Defend  Action = "defend"
    UseItem Action = "use_item"
    Dodge   Action = "dodge"
)

type MessageChooseCharacter struct {
	Character Character `json:"character"`
}

type MessageAction struct {
    Action Action
}

type MessageConfirmation struct {
    Message string `json:"message"`
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
    roomID := r.URL.Query().Get("room")
    if roomID == "" {
        http.Error(w, "Room ID is required", http.StatusBadRequest)
        return
    }

    conn, err := Upgrader.Upgrade(w, r, nil)
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

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			continue
		}

		var msg MessageChooseCharacter
		err = json.Unmarshal(message, &msg)
		if err != nil {
            var msg MessageAction
            err = json.Unmarshal(message, &msg)
            if err != nil {
			    continue
            } else {
                confirmationMessage := MessageConfirmation{
                    Message: "Character chosen successfully",
                }
                confirmationBytes, _ := json.Marshal(confirmationMessage)
                conn.WriteMessage(websocket.TextMessage, confirmationBytes)
            }
		} else {
            room.Lock()
            room.characters[conn] = msg.Character
            room.Unlock()

            confirmationMessage := MessageConfirmation{
                Message: "Character chosen successfully",
            }
            confirmationBytes, _ := json.Marshal(confirmationMessage)
            conn.WriteMessage(websocket.TextMessage, confirmationBytes)
        }

		room.broadcast <- message
	}
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
