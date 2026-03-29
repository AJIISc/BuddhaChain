package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/synapsechain/backend/internal/models"
	"github.com/synapsechain/backend/internal/repository"
	"github.com/synapsechain/backend/internal/service"
)

type LabelHandler struct {
	labelRepo *repository.LabelRepo
	dataRepo  *repository.DataRepo
	consensus *service.ConsensusEngine
}

func NewLabelHandler(labelRepo *repository.LabelRepo, dataRepo *repository.DataRepo, consensus *service.ConsensusEngine) *LabelHandler {
	return &LabelHandler{
		labelRepo: labelRepo,
		dataRepo:  dataRepo,
		consensus: consensus,
	}
}

// GET /label/{data_id} — get final label
func (h *LabelHandler) GetFinalLabel(w http.ResponseWriter, r *http.Request) {
	dataID, err := uuid.Parse(mux.Vars(r)["data_id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid data ID")
		return
	}

	final, err := h.labelRepo.GetFinalLabel(r.Context(), dataID)
	if err != nil {
		respondError(w, http.StatusNotFound, "label not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"final_label": final.FinalLabels,
		"confidence":  final.Confidence,
		"source":      final.Sources,
	})
}

// GET /label/{data_id}/ai — get AI label
func (h *LabelHandler) GetAILabel(w http.ResponseWriter, r *http.Request) {
	dataID, err := uuid.Parse(mux.Vars(r)["data_id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid data ID")
		return
	}

	aiLabel, err := h.labelRepo.GetAILabel(r.Context(), dataID)
	if err != nil {
		respondError(w, http.StatusNotFound, "AI label not found")
		return
	}

	respondJSON(w, http.StatusOK, aiLabel)
}

// GET /validation/pending — list items needing human review
func (h *LabelHandler) ListPendingValidation(w http.ResponseWriter, r *http.Request) {
	items, err := h.labelRepo.ListNeedingHumanReview(r.Context(), 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch pending items")
		return
	}

	type pendingItem struct {
		DataID     uuid.UUID              `json:"data_id"`
		Labels     map[string]interface{} `json:"labels"`
		Confidence float64                `json:"confidence"`
	}

	result := make([]pendingItem, 0, len(items))
	for _, item := range items {
		result = append(result, pendingItem{
			DataID:     item.DataID,
			Labels:     item.Labels,
			Confidence: item.Confidence,
		})
	}

	respondJSON(w, http.StatusOK, result)
}

// POST /validation/submit — submit human validation
func (h *LabelHandler) SubmitValidation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DataID      string                 `json:"data_id"`
		ValidatorID string                 `json:"validator_id"`
		Action      string                 `json:"action"`
		Labels      map[string]interface{} `json:"labels"`
		Confidence  float64                `json:"confidence"`
		Notes       string                 `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Action != "accept" && req.Action != "modify" && req.Action != "reject" {
		respondError(w, http.StatusBadRequest, "action must be 'accept', 'modify', or 'reject'")
		return
	}

	dataID, err := uuid.Parse(req.DataID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid data_id")
		return
	}

	// If action is "accept", use the AI labels
	if req.Action == "accept" {
		aiLabel, err := h.labelRepo.GetAILabel(r.Context(), dataID)
		if err != nil {
			respondError(w, http.StatusNotFound, "AI label not found")
			return
		}
		req.Labels = aiLabel.Labels
		req.Confidence = aiLabel.Confidence
	}

	humanLabel := &models.HumanLabel{
		ID:          uuid.New(),
		DataID:      dataID,
		Labels:      req.Labels,
		Confidence:  req.Confidence,
		ValidatorID: req.ValidatorID,
		Action:      req.Action,
		Notes:       req.Notes,
	}

	if err := h.labelRepo.CreateHumanLabel(r.Context(), humanLabel); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to store validation")
		return
	}

	// Run consensus to produce final label
	final, err := h.consensus.Resolve(r.Context(), dataID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "consensus failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "validated",
		"final_label": final.FinalLabels,
		"confidence":  final.Confidence,
		"sources":     final.Sources,
	})
}
