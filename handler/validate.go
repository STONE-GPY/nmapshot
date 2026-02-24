package handler

import (
	"fmt"
	"strconv"
	"strings"

	"nmapshot/config"
)

// ValidatePorts checks that all ports in the request are within the allowed port entries.
// Supported request formats per entry: single port ("80"), port range ("80-443").
func ValidatePorts(ports []string, allowed *config.AllowedPorts) error {
	for _, entry := range ports {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			return fmt.Errorf("empty port entry is not allowed")
		}

		// Check if it's a range (e.g. "80-443")
		if strings.Contains(entry, "-") {
			parts := strings.SplitN(entry, "-", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid port range format: %q", entry)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return fmt.Errorf("invalid port range start %q: %w", parts[0], err)
			}
			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Errorf("invalid port range end %q: %w", parts[1], err)
			}

			if start > end {
				return fmt.Errorf("port range start (%d) must not exceed end (%d)", start, end)
			}

			// Every port in the requested range must be within the allowed list
			for p := start; p <= end; p++ {
				if !allowed.Contains(p) {
					return fmt.Errorf("port %d (in range %s) is not in the allowed ports [%s]", p, entry, allowed)
				}
			}
		} else {
			// Single port (e.g. "80")
			port, err := strconv.Atoi(entry)
			if err != nil {
				return fmt.Errorf("invalid port %q: %w", entry, err)
			}
			if !allowed.Contains(port) {
				return fmt.Errorf("port %d is not in the allowed ports [%s]", port, allowed)
			}
		}
	}
	return nil
}
