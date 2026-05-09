package main

import (
	"log"

	"github.com/choken/llm-proxy/internal/api"
	"github.com/choken/llm-proxy/internal/config"
	"github.com/choken/llm-proxy/internal/database"
)

// @title LLM Proxy API
// @version 1.0
// @description A high-performance LLM API Proxy aggregator.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

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
