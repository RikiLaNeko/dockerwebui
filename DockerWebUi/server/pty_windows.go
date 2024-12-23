// +build windows

package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/iamacarpet/go-winpty"
    "fmt"
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        fmt.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Starting WinPTY for container:", containerID)
    // Open a WinPTY instance
    wp, err := winpty.Open("docker", "exec -it "+containerID+" cmd")
    if err != nil {
        log.Println("Winpty start error:", err)
        fmt.Println("WinPTY start error:", err)
        return
    }
    defer wp.Close()

    // Send previous logs to the client
    mu.Lock()
    if logs, exists := logStorage[containerID]; exists {
        for _, logLine := range logs {
            if err := conn.WriteMessage(websocket.TextMessage, []byte(logLine)); err != nil {
                log.Println("WriteMessage error:", err)
                fmt.Println("WriteMessage error:", err)
                mu.Unlock()
                return
            }
        }
    }
    mu.Unlock()

    go func() {
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("ReadMessage error:", err)
                fmt.Println("ReadMessage error:", err)
                break
            }
            if _, err := wp.StdIn.Write(message); err != nil {
                log.Println("Pty write error:", err)
                fmt.Println("PTY write error:", err)
                break
            }
        }
    }()

    go func() {
        buf := make([]byte, 1024)
        for {
            n, err := wp.StdOut.Read(buf)
            if err != nil {
                log.Println("Pty read error:", err)
                fmt.Println("PTY read error:", err)
                break
            }
            message := string(buf[:n])
            mu.Lock()
            logStorage[containerID] = append(logStorage[containerID], message)
            mu.Unlock()
            if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
                log.Println("WriteMessage error:", err)
                fmt.Println("WriteMessage error:", err)
                break
            }
        }
    }()

    // No need to wait for the command, wp.Close() will clean up.
    fmt.Println("WinPTY session ended for container:", containerID)
}
