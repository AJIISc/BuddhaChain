package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "synapsechain-api",
	})
}

func RegisterRoutes(
	r *mux.Router,
	dataH *DataHandler,
	labelH *LabelHandler,
) {
	// Health
	r.HandleFunc("/health", HealthCheck).Methods("GET")

	// Data ingestion
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/data/upload", dataH.Upload).Methods("POST")
	api.HandleFunc("/data/upload/file", dataH.UploadFile).Methods("POST")
	api.HandleFunc("/data/{id}", dataH.GetData).Methods("GET")

	// Labels / Output API
	api.HandleFunc("/label/{data_id}", labelH.GetFinalLabel).Methods("GET")
	api.HandleFunc("/label/{data_id}/ai", labelH.GetAILabel).Methods("GET")

	// Human validation
	api.HandleFunc("/validation/pending", labelH.ListPendingValidation).Methods("GET")
	api.HandleFunc("/validation/submit", labelH.SubmitValidation).Methods("POST")
}
