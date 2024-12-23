package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "github.com/creack/pty"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/rs/cors"
    
)

type Container struct {
    ID     string `json:"ID"`
    Names  string `json:"Names"`
    Image  string `json:"Image"`
    Status string `json:"Status"`
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/api/containers/json", getContainers).Methods("GET")
    r.HandleFunc("/api/containers/create", createContainer).Methods("POST")
    r.HandleFunc("/api/containers/{id}/start", startContainer).Methods("POST")
    r.HandleFunc("/api/containers/{id}/stop", stopContainer).Methods("POST")
    r.HandleFunc("/api/containers/{id}/logs", getContainerLogs).Methods("GET")
    r.HandleFunc("/ws/{id}", handleWebSocket)

    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type"},
    })

    handler := c.Handler(r)

    fmt.Println("Server running at http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", handler))
}

func getContainers(w http.ResponseWriter, r *http.Request) {
    cmd := exec.Command("docker", "ps", "-a", "--format", "{{json .}}")
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error fetching containers", http.StatusInternalServerError)
        return
    }

    var containers []Container
    lines := splitLines(output)
    for _, line := range lines {
        var container Container
        if err := json.Unmarshal([]byte(line), &container); err != nil {
            http.Error(w, "Error parsing container data", http.StatusInternalServerError)
            return
        }
        containers = append(containers, container)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(containers)
}

func createContainer(w http.ResponseWriter, r *http.Request) {
    var requestBody struct {
        Image string `json:"Image"`
        Name  string `json:"Names"`
    }
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    cmd := exec.Command("docker", "create", "--name", requestBody.Name, requestBody.Image)
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error creating container", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(output)
}

func startContainer(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    cmd := exec.Command("docker", "start", containerID)
    if err := cmd.Run(); err != nil {
        http.Error(w, "Error starting container", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func stopContainer(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    cmd := exec.Command("docker", "stop", containerID)
    if err := cmd.Run(); err != nil {
        http.Error(w, "Error stopping container", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func getContainerLogs(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    cmd := exec.Command("docker", "logs", containerID)
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error fetching logs", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write(output)
}

func splitLines(data []byte) []string {
    var lines []string
    start := 0
    for i, b := range data {
        if b == '\n' {
            lines = append(lines, string(data[start:i]))
            start = i + 1
        }
    }
    if start < len(data) {
        lines = append(lines, string(data[start:]))
    }
    return lines
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    cmd := exec.Command("docker", "exec", "-it", containerID, "sh")
    pty, err := pty.Start(cmd)
    if err != nil {
        log.Println("Pty start error:", err)
        return
    }
    defer pty.Close()

    go func() {
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("ReadMessage error:", err)
                break
            }
            if _, err := pty.Write(message); err != nil {
                log.Println("Pty write error:", err)
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
                break
            }
            if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
                log.Println("WriteMessage error:", err)
                break
            }
        }
    }()

    if err := cmd.Wait(); err != nil {
        log.Println("Wait error:", err)
    }
}
