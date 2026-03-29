package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/synapsechain/backend/internal/models"
	"github.com/synapsechain/backend/internal/repository"
)

// ConsensusEngine determines the final label from AI and human inputs
type ConsensusEngine struct {
	labelRepo *repository.LabelRepo
	dataRepo  *repository.DataRepo
}

func NewConsensusEngine(labelRepo *repository.LabelRepo, dataRepo *repository.DataRepo) *ConsensusEngine {
	return &ConsensusEngine{
		labelRepo: labelRepo,
		dataRepo:  dataRepo,
	}
}

// Resolve produces a final label for the given data_id.
// MVP Logic:
//   - If human label exists → use human label
//   - Otherwise → use AI label
func (c *ConsensusEngine) Resolve(ctx context.Context, dataID uuid.UUID) (*models.FinalLabel, error) {
	humanLabels, err := c.labelRepo.GetHumanLabels(ctx, dataID)
	if err != nil {
		return nil, err
	}

	aiLabel, err := c.labelRepo.GetAILabel(ctx, dataID)
	if err != nil {
		return nil, err
	}

	final := &models.FinalLabel{
		ID:     uuid.New(),
		DataID: dataID,
	}

	if len(humanLabels) > 0 {
		// Use the most recent human label
		latest := humanLabels[0]
		final.FinalLabels = latest.Labels
		final.Confidence = latest.Confidence
		final.Sources = []string{"AI", "Human"}
	} else if aiLabel != nil {
		final.FinalLabels = aiLabel.Labels
		final.Confidence = aiLabel.Confidence
		final.Sources = []string{"AI"}
	}

	if err := c.labelRepo.CreateFinalLabel(ctx, final); err != nil {
		return nil, err
	}

	if err := c.dataRepo.UpdateStatus(ctx, dataID, "labeled"); err != nil {
		return nil, err
	}

	return final, nil
}
