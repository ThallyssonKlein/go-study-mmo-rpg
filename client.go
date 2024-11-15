package main

import (
    "log"
    "net/url"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/websocket"
)

func main() {
    // Defina a URL do servidor WebSocket
    u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "room=sala"}

    // Conecte-se ao servidor WebSocket
    log.Printf("Conectando a %s", u.String())
    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("Erro ao conectar:", err)
    }
    defer c.Close()

    // Canal para capturar sinais de interrupção (Ctrl+C)
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    done := make(chan struct{})

    // Goroutine para ler mensagens do servidor
    go func() {
        defer close(done)
        for {
            _, message, err := c.ReadMessage()
            if err != nil {
                log.Println("Erro ao ler mensagem:", err)
                return
            }
            log.Printf("Mensagem recebida: %s", message)
        }
    }()
	c.WriteMessage(websocket.TextMessage, []byte(`{"action":"choose_character","character":"warrior"}`))

    // Loop principal para manter a conexão ativa
    for {
        select {
        case <-done:
            return
        case <-interrupt:
            log.Println("Interrompido pelo usuário, fechando conexão...")

            // Feche a conexão WebSocket
            err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            if err != nil {
                log.Println("Erro ao fechar conexão:", err)
                return
            }
            select {
            case <-done:
            case <-time.After(time.Second):
            }
            return
        }
    }
}