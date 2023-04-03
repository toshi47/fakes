package main

import (
	"api/auth_manager"
	"api/rest"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed to load enviroment: %v", err)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	authmgr, err := auth_manager.New()
	if err != nil {
		log.Fatal("failed to create auth mgr: %v", err)
	}

	server, err := rest.NewServer(os.Getenv("API_PORT"), authmgr)
	if err != nil {
		log.Fatal("failed to create server: %v", err)
	}
	server.Start()
}

// TODO: api to check if text is fake (mb +link)
// POST /check_news

// TODO: api to check if image is fake
// POST /check_image
