package handler

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
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
	clientIP := c.ClientIP()
	log.Printf("[Scan API] Received scan request from IP: %s", clientIP)

	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Scan API] Invalid request body from %s: %v", clientIP, err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	if len(req.Targets) == 0 {
		log.Printf("[Scan API] Missing targets in request from %s", clientIP)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "At least one target is required"})
		return
	}

	// Validate that requested ports are within the allowed range
	if err := ValidatePorts(req.Ports, h.AllowedPorts); err != nil {
		log.Printf("[Scan API] Port validation failed for %s: %v. Requested ports: %v", clientIP, err, req.Ports)
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
		log.Printf("[Scan API] Failed to create nmap scanner: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create nmap scanner: %v", err)})
		return
	}

	tracer := otel.Tracer("nmapshot/handler")
	scanCtx, span := tracer.Start(ctx, "nmap.Run")
	defer span.End()

	log.Printf("[Scan API] Starting Nmap scan for targets: %v, ports: %v (Requested by %s)", req.Targets, req.Ports, clientIP)
	startTime := time.Now()

	// Run the scan
	result, err := scanner.Run(scanCtx)
	duration := time.Since(startTime)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[Scan API] Scan timed out for %s after %v. Targets: %v", clientIP, duration, req.Targets)
			c.JSON(http.StatusGatewayTimeout, ErrorResponse{Error: "Scan timed out"})
			return
		}
		if ctx.Err() == context.Canceled {
			log.Printf("[Scan API] Scan canceled by client %s after %v. Targets: %v", clientIP, duration, req.Targets)
			c.JSON(499, ErrorResponse{Error: "Request canceled by client"})
			return
		}
		log.Printf("[Scan API] Nmap execution failed for %s: %v", clientIP, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Nmap execution failed: %v", err)})
		return
	}

	log.Printf("[Scan API] Scan completed successfully for %s in %v. Found %d hosts", clientIP, duration, len(result.Hosts))

	// Convert result back to XML
	xmlBytes, err := xml.Marshal(result)
	if err != nil {
		log.Printf("[Scan API] Failed to generate XML response for %s: %v", clientIP, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to generate XML response: %v", err)})
		return
	}

	c.Data(http.StatusOK, "application/xml", append([]byte(xml.Header), xmlBytes...))
}
