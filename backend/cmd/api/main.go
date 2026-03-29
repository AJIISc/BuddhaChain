package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/synapsechain/backend/internal/config"
	"github.com/synapsechain/backend/internal/db"
	"github.com/synapsechain/backend/internal/handler"
	"github.com/synapsechain/backend/internal/repository"
	"github.com/synapsechain/backend/internal/service"
)

func main() {
	cfg := config.Load()

	// Database
	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Repositories
	dataRepo := repository.NewDataRepo(pool)
	labelRepo := repository.NewLabelRepo(pool)

	// Services
	aiClient := service.NewAIClient(cfg.AIServiceURL)
	routingEngine := service.NewRoutingEngine(cfg.ConfidenceThreshold, cfg.HighRiskLevels)
	consensusEngine := service.NewConsensusEngine(labelRepo, dataRepo)

	// Handlers
	dataHandler := handler.NewDataHandler(dataRepo, labelRepo, aiClient, routingEngine, consensusEngine, cfg)
	labelHandler := handler.NewLabelHandler(labelRepo, dataRepo, consensusEngine)

	// Router
	r := mux.NewRouter()
	handler.RegisterRoutes(r, dataHandler, labelHandler)

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("SynapseChain API starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, c.Handler(r)))
}
