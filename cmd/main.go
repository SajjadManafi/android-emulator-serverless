package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

// AndroidConfig holds the configuration for the Android emulator
type AndroidConfig struct {
	ContainerName string `json:"containerName"`
	Port          int    `json:"port"`
	DeviceName    string `json:"DeviceName"`
	AndroidAPI    string `json:"AndroidAPI"`
	Status        string `json:"status"`
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/run-emulator", RunEmulator)
	r.HandleFunc("/stop-emulator", StopEmulator)
	r.HandleFunc("/device-status", DeviceStatus)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func RunEmulator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var android AndroidConfig
	if err := json.NewDecoder(r.Body).Decode(&android); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("got android config: ", android)

	portStr := fmt.Sprintf("%d", android.Port)
	servicePort := fmt.Sprintf("%d", 6080)
	cmd := exec.Command("docker", "run", "-d", "-p", portStr+":"+servicePort, "-e", "EMULATOR_DEVICE="+android.DeviceName, "-e", "WEB_VNC=true", "--device", "/dev/kvm", "--name", android.ContainerName, "budtmo/docker-android:"+android.AndroidAPI)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Emulator started successfully")
}

func StopEmulator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var android AndroidConfig
	if err := json.NewDecoder(r.Body).Decode(&android); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("stopping android container: ", android.ContainerName)

	// Run stop and remove commands in a goroutine
	go func(containerName string) {
		// Stop the container
		stopCmd := exec.Command("docker", "stop", containerName)
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		err := stopCmd.Run()
		if err != nil {
			log.Printf("Error stopping container %s: %s", containerName, err)
			return
		}

		// Remove the container
		rmCmd := exec.Command("docker", "rm", containerName)
		rmCmd.Stdout = os.Stdout
		rmCmd.Stderr = os.Stderr
		err = rmCmd.Run()
		if err != nil {
			log.Printf("Error removing container %s: %s", containerName, err)
			return
		}

		log.Printf("Emulator %s stopped and deleted successfully", containerName)
	}(android.ContainerName)

	// Immediately respond to the request
	fmt.Fprintf(w, "Emulator stop and delete initiated successfully")
}

func DeviceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	containerName := r.URL.Query().Get("containerName")
	if containerName == "" {
		http.Error(w, "Container name is required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("docker", "exec", "-i", containerName, "cat", "device_status")
	// Removing the '-t' option because it's not suitable for non-interactive sessions like this
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command for container %s: %s", containerName, err)
		http.Error(w, "Failed to get device status", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Device Status: %s", out.String())
}
