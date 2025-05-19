package main

import (
	"log"

	"github.com/abhiii71/clean-code-abhi/pkg/config"
	di "github.com/abhiii71/clean-code-abhi/pkg/dependencies"
)

func main() {
	//loadConfig
	cnf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load environments")
	}

	// server Initialization
	server, err := di.InitializeEvents(cnf)
	if err != nil {
		log.Fatal("Failed to initialize the files")
	}
	server.Start(cnf)
}
