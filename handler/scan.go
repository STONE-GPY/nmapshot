package handler

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v4"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"

	"nmapshot/config"
)

// ScanHandler holds dependencies for the scan endpoint.
type ScanHandler struct {
	AllowedPorts *config.AllowedPorts
}

// NewScanHandler creates a new ScanHandler with the given allowed ports configuration.
func NewScanHandler(allowed *config.AllowedPorts) *ScanHandler {
	return &ScanHandler{AllowedPorts: allowed}
}

// Handle runs an nmap scan and returns the raw XML output.
// @Summary Run Nmap Scan
// @Description Execute an nmap scan on specified targets and ports, returning raw Nmap XML output. Ports must be within the server-configured allowed range.
// @Tags scan
// @Accept json
// @Produce xml
// @Param request body ScanRequest true "Scan Parameters"
// @Success 200 {string} string "Nmap XML output"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /scan [post]
func (h *ScanHandler) Handle(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	if len(req.Targets) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "At least one target is required"})
		return
	}

	// Validate that requested ports are within the allowed range
	if err := ValidatePorts(req.Ports, h.AllowedPorts); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Port validation failed: %v", err)})
		return
	}

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// Prepare nmap options
	options := []nmap.Option{
		nmap.WithTargets(req.Targets...),
		nmap.WithPorts(strings.Join(req.Ports, ",")),
	}

	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create nmap scanner: %v", err)})
		return
	}

	tracer := otel.Tracer("nmapshot/handler")
	scanCtx, span := tracer.Start(ctx, "nmap.Run")
	defer span.End()

	// Run the scan
	result, err := scanner.Run(scanCtx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, ErrorResponse{Error: "Scan timed out"})
			return
		}
		if ctx.Err() == context.Canceled {
			c.JSON(499, ErrorResponse{Error: "Request canceled by client"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Nmap execution failed: %v", err)})
		return
	}

	// Convert result back to XML
	xmlBytes, err := xml.Marshal(result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to generate XML response: %v", err)})
		return
	}

	c.Data(http.StatusOK, "application/xml", append([]byte(xml.Header), xmlBytes...))
}
