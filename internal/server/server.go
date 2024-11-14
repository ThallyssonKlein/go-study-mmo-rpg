package server

import (
    "net/http"
    "thallyssonklein.dev/mmorpggolang/internal/websocket"
)

// Run configura e inicia o servidor HTTP
func Run() {
    http.HandleFunc("/ws", websocket.WSHandler)
    http.ListenAndServe(":8080", nil)
}
