package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "sync"

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

var (
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
    logStorage = make(map[string][]string)
    mu         sync.Mutex
)

func main() {
    // Ensure winpty DLLs are downloaded on Windows (no-op on non-Windows platforms)
    ensureWinPTY()

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
    fmt.Println("Fetching containers...")
    cmd := exec.Command("docker", "ps", "-a", "--format", "{{json .}}")
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error fetching containers", http.StatusInternalServerError)
        fmt.Println("Error fetching containers:", err)
        return
    }

    var containers []Container
    lines := splitLines(output)
    for _, line := range lines {
        var container Container
        if err := json.Unmarshal([]byte(line), &container); err != nil {
            http.Error(w, "Error parsing container data", http.StatusInternalServerError)
            fmt.Println("Error parsing container data:", err)
            return
        }
        containers = append(containers, container)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(containers)
    fmt.Println("Containers fetched successfully")
}

func createContainer(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Creating container...")
    var requestBody struct {
        Image string `json:"Image"`
        Name  string `json:"Names"`
    }
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        fmt.Println("Invalid request body:", err)
        return
    }

    cmd := exec.Command("docker", "create", "--name", requestBody.Name, requestBody.Image)
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error creating container", http.StatusInternalServerError)
        fmt.Println("Error creating container:", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(output)
    fmt.Println("Container created successfully")
}

  func startContainer(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    fmt.Printf("Starting container: %s...\n", containerID)

    // Log command to be executed
    cmd := exec.Command("docker", "start", containerID)
    fmt.Printf("Executing command: %v\n", cmd.Args)

    // Capture stdout and stderr
    var out, stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    // Run the command
    err := cmd.Run()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error starting container: %v", stderr.String()), http.StatusInternalServerError)
        fmt.Printf("Error starting container %s: %v\n", containerID, err)
        return
    }

    // Log successful start
    w.WriteHeader(http.StatusNoContent)
    fmt.Printf("Container %s started successfully. Output: %s\n", containerID, out.String())
}

func stopContainer(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    containerID := vars["id"]

    fmt.Printf("Stopping container: %s...\n", containerID)

    // Log command to be executed
    cmd := exec.Command("docker", "stop", containerID)
    fmt.Printf("Executing command: %v\n", cmd.Args)

    // Capture stdout and stderr
    var out, stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    // Run the command
    err := cmd.Run()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error stopping container: %v", stderr.String()), http.StatusInternalServerError)
        fmt.Printf("Error stopping container %s: %v\n", containerID, err)
        return
    }

    // Log successful stop
    w.WriteHeader(http.StatusNoContent)
    fmt.Printf("Container %s stopped successfully. Output: %s\n", containerID, out.String())
}

    // Log successful stop
func getContainerLogs(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Fetching container logs...")
    vars := mux.Vars(r)
    containerID := vars["id"]

    cmd := exec.Command("docker", "logs", containerID)
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error fetching logs", http.StatusInternalServerError)
        fmt.Println("Error fetching logs:", err)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write(output)
    fmt.Println("Container logs fetched successfully")
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
