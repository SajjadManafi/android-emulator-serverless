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
	Port       string `json:"port"`
	DeviceName string `json:"DeviceName"`
	AndroidAPI string `json:"AndroidAPI"`
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

		portStr := android.Port
		cmd := exec.Command("docker", "run", "-d", "-p", portStr+":"+portStr, "-e", "EMULATOR_DEVICE="+android.DeviceName, "-e", "WEB_VNC=true", "--device", "/dev/kvm", "budtmo/docker-android:"+android.AndroidAPI)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Emulator started successfully")
	})

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
