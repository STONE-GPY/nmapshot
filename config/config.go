package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PortEntry represents a single allowed port or a contiguous port range.
type PortEntry struct {
	Min int
	Max int // same as Min for a single port
}

// String returns the human-readable representation of the entry.
func (p PortEntry) String() string {
	if p.Min == p.Max {
		return strconv.Itoa(p.Min)
	}
	return fmt.Sprintf("%d-%d", p.Min, p.Max)
}

// Contains reports whether port falls within this entry.
func (p PortEntry) Contains(port int) bool {
	return port >= p.Min && port <= p.Max
}

// AllowedPorts holds the list of allowed port entries for scan requests.
type AllowedPorts struct {
	Entries []PortEntry
}

// String returns the human-readable representation of all allowed ports.
func (a *AllowedPorts) String() string {
	parts := make([]string, len(a.Entries))
	for i, e := range a.Entries {
		parts[i] = e.String()
	}
	return strings.Join(parts, ",")
}

// Contains reports whether the given port is within any allowed entry.
func (a *AllowedPorts) Contains(port int) bool {
	for _, e := range a.Entries {
		if e.Contains(port) {
			return true
		}
	}
	return false
}

const (
	// AllowedPortsEnvVar is the environment variable that specifies the allowed ports.
	// Format: comma-separated list of ports and/or ranges.
	// Example: "80,100,125-256,1088-1090"
	AllowedPortsEnvVar = "ALLOWED_PORTS"
	//
	RestAPIPort = "REST_API_PORT"
)

// LoadAllowedPorts reads the allowed port list from the ALLOWED_PORTS environment variable.
// If not set, all ports (1-65535) are allowed.
// Supported format: "80,100,125-256,1088-1090"
func LoadAllowedPorts() (*AllowedPorts, error) {
	raw := os.Getenv(AllowedPortsEnvVar)
	if raw == "" {
		return &AllowedPorts{
			Entries: []PortEntry{{Min: 1, Max: 65535}},
		}, nil
	}

	return ParseAllowedPorts(raw)
}

// ParseAllowedPorts parses a comma-separated port specification string
// into an AllowedPorts struct. Each token can be a single port ("80") or
// a range ("125-256").
func ParseAllowedPorts(raw string) (*AllowedPorts, error) {
	tokens := strings.Split(raw, ",")
	entries := make([]PortEntry, 0, len(tokens))

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if strings.Contains(token, "-") {
			parts := strings.SplitN(token, "-", 2)
			min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid port range start in %q: %w", token, err)
			}
			max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid port range end in %q: %w", token, err)
			}
			if min > max {
				return nil, fmt.Errorf("port range start (%d) exceeds end (%d) in %q", min, max, token)
			}
			if err := validatePort(min); err != nil {
				return nil, fmt.Errorf("in range %q: %w", token, err)
			}
			if err := validatePort(max); err != nil {
				return nil, fmt.Errorf("in range %q: %w", token, err)
			}
			entries = append(entries, PortEntry{Min: min, Max: max})
		} else {
			port, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("invalid port %q: %w", token, err)
			}
			if err := validatePort(port); err != nil {
				return nil, err
			}
			entries = append(entries, PortEntry{Min: port, Max: port})
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("%s is set but contains no valid port entries", AllowedPortsEnvVar)
	}

	return &AllowedPorts{Entries: entries}, nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d is out of valid range (1-65535)", port)
	}
	return nil
}

func LoadRestAPIPort() string {
	port := os.Getenv(RestAPIPort)
	if p, e := strconv.Atoi(port); e != nil {
		return "8082"
	} else if p > 0 && p < 65535 {
		return port
	} else {
		return "8082"
	}
}
