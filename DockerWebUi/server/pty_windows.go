// +build windows

package main

import (
    "fmt"
    "io"
    "log"
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
    downloadDir    = "winpty"
)

func main() {
    // Create download directory if it doesn't exist
    if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
        os.Mkdir(downloadDir, os.ModePerm)
    }

    // Download winpty.dll and winpty-agent.exe if not present
    downloadFileIfMissing(filepath.Join(downloadDir, dllFilename), dllURL)
    downloadFileIfMissing(filepath.Join(downloadDir, agentFilename), agentURL)

    // Add your other main logic here
}

func downloadFileIfMissing(filepath string, url string) {
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        fmt.Println("Downloading", filepath)
        out, err := os.Create(filepath)
        if err != nil {
            log.Fatal(err)
        }
        defer out.Close()

        resp, err := http.Get(url)
        if err != nil {
            log.Fatal(err)
        }
        defer resp.Body.Close()

        _, err = io.Copy(out, resp.Body)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println("Downloaded", filepath)
    }
}

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

    // Set DLL search path to the download directory
    winpty.SetDllSearchPath(downloadDir)

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
