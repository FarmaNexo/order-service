package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/farmanexo/order-service/internal/domain/services"
	"go.uber.org/zap"
)

type UserClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewUserClient(baseURL string, logger *zap.Logger) *UserClientImpl {
	return &UserClientImpl{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * 1e9},
		logger:     logger,
	}
}

func (c *UserClientImpl) GetAddress(ctx context.Context, userID, addressID, token string) (*services.UserAddress, error) {
	url := fmt.Sprintf("%s/api/v1/users/me/addresses/%s", c.baseURL, addressID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando User Service", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Data *services.UserAddress `json:"datos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

var _ services.UserClient = (*UserClientImpl)(nil)
