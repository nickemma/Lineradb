package main

import (
	"fmt"
	"log"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("LineraDB v%s\n", version)
		return
	}

	log.Println("LineraDB starting...")
	fmt.Println("🚀 LineraDB - Globally Distributed SQL Database")
	fmt.Printf("Version: %s\n", version)
	log.Println("Server initialized successfully")
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}
