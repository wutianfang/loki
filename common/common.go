package common


type CommonResponse struct {
	Error string `json:"error"`
	Errno int    `json:"errno"`
}