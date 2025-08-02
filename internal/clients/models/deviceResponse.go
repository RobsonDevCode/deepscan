package models

type DeviceResposnse struct {
	DeviceCode      string `json:"deviceCode"`
	UserCode        string `json:"user_code"`
	VerificationUrl string `json:"verification_uri"`
	ExpriesIn       int16  `json:"expires_in"`
	Interval        int8   `json:"interval"`
}
