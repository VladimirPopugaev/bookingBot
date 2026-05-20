package domain

type SiteInfoResponse struct {
	URL                     string `json:"url"`
	Title                   string `json:"title,omitempty"`
	IsRegistrationAvailable bool   `json:"is_registration_available"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
