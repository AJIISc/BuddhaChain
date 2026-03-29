package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synapsechain/backend/internal/models"
)

type DataRepo struct {
	pool *pgxpool.Pool
}

func NewDataRepo(pool *pgxpool.Pool) *DataRepo {
	return &DataRepo{pool: pool}
}

func (r *DataRepo) Create(ctx context.Context, d *models.Data) error {
	metaJSON, _ := json.Marshal(d.Metadata)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO data (id, type, raw_data_url, metadata, uploaded_by, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		d.ID, d.Type, d.RawDataURL, metaJSON, d.UploadedBy, d.Status,
	)
	return err
}

func (r *DataRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Data, error) {
	d := &models.Data{}
	var metaJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, type, raw_data_url, metadata, uploaded_by, status, created_at, updated_at
		 FROM data WHERE id = $1`, id,
	).Scan(&d.ID, &d.Type, &d.RawDataURL, &metaJSON, &d.UploadedBy, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metaJSON, &d.Metadata)
	return d, nil
}

func (r *DataRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE data SET status = $1, updated_at = NOW() WHERE id = $2`, status, id,
	)
	return err
}

func (r *DataRepo) ListPending(ctx context.Context, limit int) ([]*models.Data, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, type, raw_data_url, metadata, uploaded_by, status, created_at, updated_at
		 FROM data WHERE status IN ('received', 'processing')
		 ORDER BY created_at ASC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.Data
	for rows.Next() {
		d := &models.Data{}
		var metaJSON []byte
		if err := rows.Scan(&d.ID, &d.Type, &d.RawDataURL, &metaJSON, &d.UploadedBy, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		json.Unmarshal(metaJSON, &d.Metadata)
		results = append(results, d)
	}
	return results, nil
}
