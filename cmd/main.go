package main

import (
	"log"
	"net/http"

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

	r.HandleFunc("/run-emulator", runEmulator)
	r.HandleFunc("/stop-emulator", stopEmulator)
	r.HandleFunc("/device-status", deviceStatus)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
