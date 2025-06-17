package main

import (
	"log"

	_ "github.com/abhiii71/clean-code-abhi/docs"
	"github.com/abhiii71/clean-code-abhi/pkg/config"
	di "github.com/abhiii71/clean-code-abhi/pkg/dependencies"
)

// @title	Clean-Code-Arch
// @version 1.0
// @description	This is clean code architecture.
// @contact.name	Abhishek
// @contact.url		https://linktr.ee/abhiii71
// @contact.email	abhishek.work71@gmail.com
// @host	localhost:8080
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
