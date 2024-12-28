// +build windows

package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/iamacarpet/go-winpty"
)

const (
    dllURL         = "https://github.com/rprichard/winpty/releases/download/latest/winpty.dll"
    agentURL       = "https://github.com/rprichard/winpty/releases/download/latest/winpty-agent.exe"
    dllFilename    = "winpty.dll"
    agentFilename  = "winpty-agent.exe"
    downloadDir    = "."
)

// ensureWinPTY downloads winpty DLLs if they are missing
func ensureWinPTY() {
    // Download winpty.dll and winpty-agent.exe if not present in the current directory
    downloadFileIfMissing(filepath.Join(downloadDir, dllFilename), dllURL)
    downloadFileIfMissing(filepath.Join(downloadDir, agentFilename), agentURL)
}

func downloadFileIfMissing(filepath string, url string) {
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        fmt.Println("Downloading", filepath)
        out, err := os.Create(filepath)
        if err != nil {
            fmt.Println("Error creating file:", err)
            return
        }
        defer out.Close()

        resp, err := http.Get(url)
        if err != nil {
            fmt.Println("Error downloading file:", err)
            return
        }
        defer resp.Body.Close()

        _, err = io.Copy(out, resp.Body)
        if err != nil {
            fmt.Println("Error saving file:", err)
            return
        }
        fmt.Println("Downloaded", filepath)
    }
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Starting WinPTY for container:", containerID)

    // Open a WinPTY instance
    wp, err := winpty.Open("docker", "exec -it "+containerID+" cmd")
    if err != nil {
        fmt.Println("WinPTY start error:", err)
        return
    }
    defer wp.Close()

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
                fmt.Println("ReadMessage error:", err)
                break
            }
            if _, err := wp.StdIn.Write(message); err != nil {
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
                fmt.Println("PTY read error:", err)
                break
            }
            message := string(buf[:n])
            mu.Lock()
            logStorage[containerID] = append(logStorage[containerID], message)
            mu.Unlock()
            if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
                fmt.Println("WriteMessage error:", err)
                break
            }
        }
    }()

    fmt.Println("WinPTY session ended for container:", containerID)
}
