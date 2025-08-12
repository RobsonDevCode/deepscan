package authenticaionmodels

type AuthenticationError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
	Interval         int16  `json:"interval"`
}
