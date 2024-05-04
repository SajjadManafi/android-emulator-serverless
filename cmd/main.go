package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// AndroidConfig holds the configuration for the Android emulator
type AndroidConfig struct {
	ContainerName string `json:"containerName"`
	Port          int    `json:"port"`
	DeviceName    string `json:"DeviceName"`
	AndroidAPI    string `json:"AndroidAPI"`
}

func main() {
	http.HandleFunc("/run-emulator", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/stop-emulator", func(w http.ResponseWriter, r *http.Request) {
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

		// Stop the container
		stopCmd := exec.Command("docker", "stop", android.ContainerName)
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		err := stopCmd.Run()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Remove the container
		rmCmd := exec.Command("docker", "rm", android.ContainerName)
		rmCmd.Stdout = os.Stdout
		rmCmd.Stderr = os.Stderr
		err = rmCmd.Run()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Emulator stopped and deleted successfully")
	})
	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
