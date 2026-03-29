package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIClient communicates with the Python AI labeling service
type AIClient struct {
	baseURL    string
	httpClient *http.Client
}

type AILabelRequest struct {
	DataID   string                 `json:"data_id"`
	Type     string                 `json:"type"`
	RawData  map[string]interface{} `json:"raw_data"`
	Metadata map[string]interface{} `json:"metadata"`
}

type AILabelResponse struct {
	Labels           map[string]interface{} `json:"labels"`
	Confidence       float64                `json:"confidence"`
	ModelVersion     string                 `json:"model_version"`
	ProcessingTimeMs int                    `json:"processing_time_ms"`
}

func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AIClient) Label(ctx context.Context, req *AILabelRequest) (*AILabelResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/label", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ai service request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ai service error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result AILabelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
