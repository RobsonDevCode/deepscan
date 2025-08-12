package authenticaionmodels

type DeviceResposnse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUrl string `json:"verification_uri"`
	ExpriesIn       int16  `json:"expires_in"`
	Interval        int8   `json:"interval"`
}
