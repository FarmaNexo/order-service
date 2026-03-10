package services

import "context"

type UserAddress struct {
	ID         string `json:"id"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Reference  string `json:"reference"`
	IsDefault  bool   `json:"is_default"`
}

type UserClient interface {
	GetAddress(ctx context.Context, userID, addressID, token string) (*UserAddress, error)
}
