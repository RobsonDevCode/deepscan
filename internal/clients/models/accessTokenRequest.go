package models

type AccessTokenRequest struct {
	ClientId   string `json:"client_id"`
	GrantType  string `json:"grant_type"`
	DeviceCode string `json:"device_code"`
}
