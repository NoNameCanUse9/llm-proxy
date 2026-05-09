package main

import (
	"log"

	"github.com/choken/llm-proxy/internal/api"
	"github.com/choken/llm-proxy/internal/config"
	"github.com/choken/llm-proxy/internal/database"
)

func main() {
	// 1. Load config (Env based)
	config.LoadConfig("")

	// 2. Initialize database
	database.InitDB(config.GlobalConfig.DBPath)

	// 3. Start server
	server := api.NewServer()
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
