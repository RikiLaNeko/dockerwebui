package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"

    "github.com/gorilla/mux"
)

type Container struct {
    ID     string   `json:"Id"`
    Names  []string `json:"Names"`
    Image  string   `json:"Image"`
    Status string   `json:"Status"`
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/api/containers/json", getContainers).Methods("GET")
    r.HandleFunc("/api/containers/create", createContainer).Methods("POST")
    r.HandleFunc("/api/containers/{id}/start", startContainer).Methods("POST")
    r.HandleFunc("/api/containers/{id}/stop", stopContainer).Methods("POST")

    fmt.Println("Server running at http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", r))
}

func getContainers(w http.ResponseWriter, r *http.Request) {
    cmd := exec.Command("docker", "ps", "-a", "--format", "{{json .}}")
    output, err := cmd.Output()
    if err != nil {
        http.Error(w, "Error fetching containers", http.StatusInternalServerError)
        return
    }

    var containers []Container
    // 'output' is a slice of bytes representing multiple lines, one per container
    // It's best to split that into lines first to unmarshal each container JSON.
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
        Name  string `json:"name"`
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

// splitLines splits the command output into lines for parsing each JSON record.
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
