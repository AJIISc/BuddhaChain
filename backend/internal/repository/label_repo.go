package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synapsechain/backend/internal/models"
)

type LabelRepo struct {
	pool *pgxpool.Pool
}

func NewLabelRepo(pool *pgxpool.Pool) *LabelRepo {
	return &LabelRepo{pool: pool}
}

// --- AI Labels ---

func (r *LabelRepo) CreateAILabel(ctx context.Context, l *models.AILabel) error {
	labelsJSON, _ := json.Marshal(l.Labels)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ai_labels (id, data_id, labels, confidence, model_version, processing_time_ms)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (data_id) DO UPDATE SET labels = $3, confidence = $4, model_version = $5, processing_time_ms = $6`,
		l.ID, l.DataID, labelsJSON, l.Confidence, l.ModelVersion, l.ProcessingTimeMs,
	)
	return err
}

func (r *LabelRepo) GetAILabel(ctx context.Context, dataID uuid.UUID) (*models.AILabel, error) {
	l := &models.AILabel{}
	var labelsJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, data_id, labels, confidence, model_version, processing_time_ms, created_at
		 FROM ai_labels WHERE data_id = $1`, dataID,
	).Scan(&l.ID, &l.DataID, &labelsJSON, &l.Confidence, &l.ModelVersion, &l.ProcessingTimeMs, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(labelsJSON, &l.Labels)
	return l, nil
}

// --- Human Labels ---

func (r *LabelRepo) CreateHumanLabel(ctx context.Context, l *models.HumanLabel) error {
	labelsJSON, _ := json.Marshal(l.Labels)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO human_labels (id, data_id, labels, confidence, validator_id, action, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		l.ID, l.DataID, labelsJSON, l.Confidence, l.ValidatorID, l.Action, l.Notes,
	)
	return err
}

func (r *LabelRepo) GetHumanLabels(ctx context.Context, dataID uuid.UUID) ([]*models.HumanLabel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, data_id, labels, confidence, validator_id, action, notes, created_at
		 FROM human_labels WHERE data_id = $1 ORDER BY created_at DESC`, dataID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.HumanLabel
	for rows.Next() {
		l := &models.HumanLabel{}
		var labelsJSON []byte
		if err := rows.Scan(&l.ID, &l.DataID, &labelsJSON, &l.Confidence, &l.ValidatorID, &l.Action, &l.Notes, &l.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(labelsJSON, &l.Labels)
		results = append(results, l)
	}
	return results, nil
}

// --- Final Labels ---

func (r *LabelRepo) CreateFinalLabel(ctx context.Context, l *models.FinalLabel) error {
	labelsJSON, _ := json.Marshal(l.FinalLabels)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO final_labels (id, data_id, final_labels, confidence, sources)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (data_id) DO UPDATE SET final_labels = $3, confidence = $4, sources = $5, updated_at = NOW()`,
		l.ID, l.DataID, labelsJSON, l.Confidence, l.Sources,
	)
	return err
}

func (r *LabelRepo) GetFinalLabel(ctx context.Context, dataID uuid.UUID) (*models.FinalLabel, error) {
	l := &models.FinalLabel{}
	var labelsJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, data_id, final_labels, confidence, sources, created_at, updated_at
		 FROM final_labels WHERE data_id = $1`, dataID,
	).Scan(&l.ID, &l.DataID, &labelsJSON, &l.Confidence, &l.Sources, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(labelsJSON, &l.FinalLabels)
	return l, nil
}

// --- Pending for Human Review ---

func (r *LabelRepo) ListNeedingHumanReview(ctx context.Context, limit int) ([]*models.AILabel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.data_id, a.labels, a.confidence, a.model_version, a.processing_time_ms, a.created_at
		 FROM ai_labels a
		 JOIN data d ON d.id = a.data_id
		 WHERE d.status = 'processing'
		   AND a.data_id NOT IN (SELECT data_id FROM human_labels)
		   AND a.data_id NOT IN (SELECT data_id FROM final_labels)
		 ORDER BY a.created_at ASC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.AILabel
	for rows.Next() {
		l := &models.AILabel{}
		var labelsJSON []byte
		if err := rows.Scan(&l.ID, &l.DataID, &labelsJSON, &l.Confidence, &l.ModelVersion, &l.ProcessingTimeMs, &l.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(labelsJSON, &l.Labels)
		results = append(results, l)
	}
	return results, nil
}
