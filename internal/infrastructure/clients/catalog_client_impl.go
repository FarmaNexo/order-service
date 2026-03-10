package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/farmanexo/order-service/internal/domain/services"
	"go.uber.org/zap"
)

type CatalogClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewCatalogClient(baseURL string, logger *zap.Logger) *CatalogClientImpl {
	return &CatalogClientImpl{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * 1e9},
		logger:     logger,
	}
}

func (c *CatalogClientImpl) GetProduct(ctx context.Context, productID string) (*services.ProductInfo, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Catalog Service", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog service returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Data *services.ProductInfo `json:"datos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

var _ services.CatalogClient = (*CatalogClientImpl)(nil)
