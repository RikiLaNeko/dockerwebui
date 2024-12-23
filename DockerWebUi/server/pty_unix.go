// +build !windows

package main

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os/exec"
)

func ensureWinPTY() {}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {

		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Starting PTY for container:", containerID)
	cmd := exec.Command("docker", "exec", "-it", containerID, "sh")
	pty, err := pty.Start(cmd)
	if err != nil {

		fmt.Println("PTY start error:", err)
		return
	}
	defer pty.Close()

	// Send previous logs to the client
	mu.Lock()
	if logs, exists := logStorage[containerID]; exists {
		for _, logLine := range logs {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(logLine)); err != nil {

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
			if _, err := pty.Write(message); err != nil {
				log.Println("Pty write error:", err)
				fmt.Println("PTY write error:", err)
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := pty.Read(buf)
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

	if err := cmd.Wait(); err != nil {
		log.Println("Wait error:", err)
		fmt.Println("Wait error:", err)
	}
	fmt.Println("PTY session ended for container:", containerID)
}
