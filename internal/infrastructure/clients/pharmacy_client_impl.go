package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/farmanexo/order-service/internal/domain/services"
	"go.uber.org/zap"
)

type PharmacyClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewPharmacyClient(baseURL string, logger *zap.Logger) *PharmacyClientImpl {
	return &PharmacyClientImpl{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * 1e9}, // 3s
		logger:     logger,
	}
}

func (c *PharmacyClientImpl) GetInventoryItem(ctx context.Context, pharmacyID, productID string) (*services.PharmacyInventoryItem, error) {
	url := fmt.Sprintf("%s/api/v1/pharmacies/%s/inventory/%s", c.baseURL, pharmacyID, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Pharmacy Service", zap.String("url", url), zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pharmacy service returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Data *services.PharmacyInventoryItem `json:"datos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (c *PharmacyClientImpl) GetPharmacyInfo(ctx context.Context, pharmacyID string) (*services.PharmacyInfo, error) {
	url := fmt.Sprintf("%s/api/v1/pharmacies/%s", c.baseURL, pharmacyID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Pharmacy Service", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pharmacy service returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Data *services.PharmacyInfo `json:"datos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

var _ services.PharmacyClient = (*PharmacyClientImpl)(nil)
