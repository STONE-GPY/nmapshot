package handler

// ScanRequest defines the request body for Nmap scan
type ScanRequest struct {
	Targets []string `json:"targets" binding:"required,gt=0" example:"127.0.0.1,scanme.nmap.org"`
	Ports   []string `json:"ports" binding:"required,gt=0" example:"80,443"`
}

// ErrorResponse defines the error response structure
type ErrorResponse struct {
	Error string `json:"error" example:"failed to run nmap scan: ..."`
}
