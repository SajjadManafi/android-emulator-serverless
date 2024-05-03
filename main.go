package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}
	fmt.Printf("Docker version: %s", output)
}
