package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"

	"github.com/SajjadManafi/android-emulator-serverless/internal/config"
	"github.com/SajjadManafi/android-emulator-serverless/internal/token"
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

var DevicesPortMap = map[string]string{}

var TokenMaker token.Maker

func main() {

	config, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	TokenMaker, err = token.NewPasetoMaker(config.Token.SecretKey)
	if err != nil {
		log.Fatalf("failed to init token maker: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/run-emulator", RunEmulator)
	r.HandleFunc("/stop-emulator", StopEmulator)
	r.HandleFunc("/device-status", DeviceStatus)

	//TODO: maybe ned to run this in another port
	r.PathPrefix("/").Handler(HandleProxy())

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

// NewProxy creates a new reverse proxy for the given target.
func NewProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}

func HandleProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the token from the Authorization header
		auth := r.Header.Get("Authorization")

		if auth == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// validate token
		claims, err := TokenMaker.VerifyAccessToken(auth)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		port := DevicesPortMap[claims.Username+"-Device"]

		if port == "" {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		proxy := NewProxy("http://localhost:" + port)
		log.Println("Proxying request to: ", "http://localhost:"+port)
		proxy.ServeHTTP(w, r)
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

	DevicesPortMap[android.ContainerName] = portStr

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

	delete(DevicesPortMap, android.ContainerName)

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
