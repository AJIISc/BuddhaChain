package models

import (
	"time"

	"github.com/google/uuid"
)

// Data represents uploaded data
type Data struct {
	ID         uuid.UUID              `json:"id"`
	Type       string                 `json:"type"`
	RawDataURL string                 `json:"raw_data_url"`
	Metadata   map[string]interface{} `json:"metadata"`
	UploadedBy string                 `json:"uploaded_by,omitempty"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// AILabel represents AI-generated labels
type AILabel struct {
	ID               uuid.UUID              `json:"id"`
	DataID           uuid.UUID              `json:"data_id"`
	Labels           map[string]interface{} `json:"labels"`
	Confidence       float64                `json:"confidence"`
	ModelVersion     string                 `json:"model_version,omitempty"`
	ProcessingTimeMs int                    `json:"processing_time_ms,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// HumanLabel represents human validator labels
type HumanLabel struct {
	ID          uuid.UUID              `json:"id"`
	DataID      uuid.UUID              `json:"data_id"`
	Labels      map[string]interface{} `json:"labels"`
	Confidence  float64                `json:"confidence,omitempty"`
	ValidatorID string                 `json:"validator_id"`
	Action      string                 `json:"action"`
	Notes       string                 `json:"notes,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// FinalLabel represents the consensus label
type FinalLabel struct {
	ID          uuid.UUID              `json:"id"`
	DataID      uuid.UUID              `json:"data_id"`
	FinalLabels map[string]interface{} `json:"final_labels"`
	Confidence  float64                `json:"confidence"`
	Sources     []string               `json:"sources"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Validator represents a human validator
type Validator struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	ReputationScore  float64   `json:"reputation_score"`
	TotalValidations int       `json:"total_validations"`
	CreatedAt        time.Time `json:"created_at"`
}

// UsageRecord tracks API usage
type UsageRecord struct {
	ID        uuid.UUID `json:"id"`
	APIKeyID  uuid.UUID `json:"api_key_id"`
	Endpoint  string    `json:"endpoint"`
	DataID    uuid.UUID `json:"data_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
