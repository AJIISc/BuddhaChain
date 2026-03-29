package service

import "strings"

// RoutingEngine decides whether data needs human validation
type RoutingEngine struct {
	confidenceThreshold float64
	highRiskLevels      map[string]bool
}

type RoutingDecision struct {
	NeedsHumanReview bool   `json:"needs_human_review"`
	Reason           string `json:"reason"`
}

func NewRoutingEngine(threshold float64, riskLevels []string) *RoutingEngine {
	levels := make(map[string]bool)
	for _, l := range riskLevels {
		levels[strings.ToLower(strings.TrimSpace(l))] = true
	}
	return &RoutingEngine{
		confidenceThreshold: threshold,
		highRiskLevels:      levels,
	}
}

func (r *RoutingEngine) Decide(labels map[string]interface{}, confidence float64) RoutingDecision {
	// Check confidence threshold
	if confidence < r.confidenceThreshold {
		return RoutingDecision{
			NeedsHumanReview: true,
			Reason:           "confidence below threshold",
		}
	}

	// Check for high-risk labels
	if risk, ok := labels["risk"]; ok {
		if riskStr, ok := risk.(string); ok {
			if r.highRiskLevels[strings.ToLower(riskStr)] {
				return RoutingDecision{
					NeedsHumanReview: true,
					Reason:           "high risk level detected",
				}
			}
		}
	}

	return RoutingDecision{
		NeedsHumanReview: false,
		Reason:           "auto-accepted",
	}
}
