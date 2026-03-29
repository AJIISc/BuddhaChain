package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/synapsechain/backend/internal/config"
	"github.com/synapsechain/backend/internal/models"
	"github.com/synapsechain/backend/internal/repository"
	"github.com/synapsechain/backend/internal/service"
)

type DataHandler struct {
	dataRepo  *repository.DataRepo
	labelRepo *repository.LabelRepo
	aiClient  *service.AIClient
	router    *service.RoutingEngine
	consensus *service.ConsensusEngine
	cfg       *config.Config
}

func NewDataHandler(
	dataRepo *repository.DataRepo,
	labelRepo *repository.LabelRepo,
	aiClient *service.AIClient,
	router *service.RoutingEngine,
	consensus *service.ConsensusEngine,
	cfg *config.Config,
) *DataHandler {
	return &DataHandler{
		dataRepo:  dataRepo,
		labelRepo: labelRepo,
		aiClient:  aiClient,
		router:    router,
		consensus: consensus,
		cfg:       cfg,
	}
}

// POST /data/upload — upload event JSON
func (h *DataHandler) Upload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type     string                 `json:"type"`
		Data     map[string]interface{} `json:"data"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Type != "event" && req.Type != "content" {
		respondError(w, http.StatusBadRequest, "type must be 'event' or 'content'")
		return
	}

	dataID := uuid.New()

	// Store raw data as a file
	rawPath := filepath.Join(h.cfg.UploadDir, dataID.String()+".json")
	if err := os.MkdirAll(h.cfg.UploadDir, 0750); err != nil {
		respondError(w, http.StatusInternalServerError, "storage error")
		return
	}
	rawFile, err := os.Create(rawPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "storage error")
		return
	}
	defer rawFile.Close()
	json.NewEncoder(rawFile).Encode(req.Data)

	d := &models.Data{
		ID:         dataID,
		Type:       req.Type,
		RawDataURL: rawPath,
		Metadata:   req.Metadata,
		Status:     "received",
	}

	if err := h.dataRepo.Create(r.Context(), d); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to store data")
		return
	}

	// Trigger async AI labeling (fire-and-forget in MVP; production would use a queue)
	go h.processLabel(dataID, req.Type, req.Data, req.Metadata)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data_id": dataID.String(),
		"status":  "received",
	})
}

// POST /data/upload/file — upload binary file (video, etc.)
func (h *DataHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadSize)

	if err := r.ParseMultipartForm(h.cfg.MaxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "file too large")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	dataType := r.FormValue("type")
	if dataType != "event" && dataType != "content" {
		respondError(w, http.StatusBadRequest, "type must be 'event' or 'content'")
		return
	}

	dataID := uuid.New()
	ext := filepath.Ext(header.Filename)
	rawPath := filepath.Join(h.cfg.UploadDir, dataID.String()+ext)

	if err := os.MkdirAll(h.cfg.UploadDir, 0750); err != nil {
		respondError(w, http.StatusInternalServerError, "storage error")
		return
	}
	dst, err := os.Create(rawPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "storage error")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	d := &models.Data{
		ID:         dataID,
		Type:       dataType,
		RawDataURL: rawPath,
		Status:     "received",
	}

	if err := h.dataRepo.Create(r.Context(), d); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to store data")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data_id": dataID.String(),
		"status":  "received",
	})
}

// GET /data/{id}
func (h *DataHandler) GetData(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid data ID")
		return
	}

	d, err := h.dataRepo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "data not found")
		return
	}

	respondJSON(w, http.StatusOK, d)
}

// processLabel runs AI labeling & routing asynchronously
func (h *DataHandler) processLabel(dataID uuid.UUID, dataType string, rawData, metadata map[string]interface{}) {
	ctx := r_context()

	if err := h.dataRepo.UpdateStatus(ctx, dataID, "processing"); err != nil {
		log.Printf("ERROR: update status for %s: %v", dataID, err)
		return
	}

	aiResp, err := h.aiClient.Label(ctx, &service.AILabelRequest{
		DataID:   dataID.String(),
		Type:     dataType,
		RawData:  rawData,
		Metadata: metadata,
	})
	if err != nil {
		log.Printf("ERROR: AI labeling for %s: %v", dataID, err)
		h.dataRepo.UpdateStatus(ctx, dataID, "failed")
		return
	}

	aiLabel := &models.AILabel{
		ID:               uuid.New(),
		DataID:           dataID,
		Labels:           aiResp.Labels,
		Confidence:       aiResp.Confidence,
		ModelVersion:     aiResp.ModelVersion,
		ProcessingTimeMs: aiResp.ProcessingTimeMs,
	}
	if err := h.labelRepo.CreateAILabel(ctx, aiLabel); err != nil {
		log.Printf("ERROR: store AI label for %s: %v", dataID, err)
		return
	}

	decision := h.router.Decide(aiResp.Labels, aiResp.Confidence)

	if !decision.NeedsHumanReview {
		// Auto-accept: run consensus immediately
		if _, err := h.consensus.Resolve(ctx, dataID); err != nil {
			log.Printf("ERROR: consensus for %s: %v", dataID, err)
		}
	}
	// If needs human review, it stays in 'processing' for human validation UI
	fmt.Printf("Data %s: AI confidence=%.2f, routing=%s (%s)\n",
		dataID, aiResp.Confidence, map[bool]string{true: "HUMAN", false: "AUTO"}[decision.NeedsHumanReview], decision.Reason)
}

func r_context() context.Context {
	return context.Background()
}
