package service

import (
	"strings"
)

type XrayErrorType string

const (
	ErrorTypeInbound  XrayErrorType = "inbound"
	ErrorTypeOutbound XrayErrorType = "outbound"
	ErrorTypeRouting  XrayErrorType = "routing"
	ErrorTypeGeneral  XrayErrorType = "general"
)

type XrayError struct {
	Type    XrayErrorType `json:"type"`
	Tag     string        `json:"tag,omitempty"`
	Message string        `json:"message"`
	Raw     string        `json:"raw"`
}

// ParseXrayError categorizes Xray errors based on their content
func (s *XrayService) ParseXrayError() *XrayError {
	errorMsg := s.GetXrayResult()
	if errorMsg == "" {
		return nil
	}

	xrayErr := &XrayError{
		Type:    ErrorTypeGeneral,
		Message: errorMsg,
		Raw:     errorMsg,
	}

	// Convert to lowercase for case-insensitive matching
	lowerMsg := strings.ToLower(errorMsg)

	// Determine error type based on keywords
	switch {
	case strings.Contains(lowerMsg, "outbound"):
		xrayErr.Type = ErrorTypeOutbound
		// Extract outbound tag if present
		if strings.Contains(errorMsg, "tag") {
			xrayErr.Tag = extractTag(errorMsg, "outbound")
		}
		xrayErr.Message = cleanErrorMessage(errorMsg)

	case strings.Contains(lowerMsg, "inbound"):
		xrayErr.Type = ErrorTypeInbound
		// Extract inbound tag if present
		if strings.Contains(errorMsg, "tag") {
			xrayErr.Tag = extractTag(errorMsg, "inbound")
		}
		xrayErr.Message = cleanErrorMessage(errorMsg)

	case strings.Contains(lowerMsg, "routing") || strings.Contains(lowerMsg, "router"):
		xrayErr.Type = ErrorTypeRouting
		xrayErr.Message = cleanErrorMessage(errorMsg)

	default:
		// General configuration error
		xrayErr.Message = cleanErrorMessage(errorMsg)
	}

	return xrayErr
}

// extractTag attempts to extract the tag name from error message
func extractTag(errorMsg, configType string) string {
	// Look for patterns like "tag xxx" or "with tag xxx"
	patterns := []string{
		"tag " + configType + "-",
		"with tag ",
		"tag out-",
		"tag in-",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(errorMsg, pattern); idx != -1 {
			start := idx + len(pattern)
			// Find the end of the tag (space, > or newline)
			end := strings.IndexAny(errorMsg[start:], " >\n")
			if end == -1 {
				end = len(errorMsg) - start
			}
			if end > 0 {
				tag := errorMsg[start : start+end]
				// Remove trailing punctuation
				tag = strings.TrimRight(tag, " >.,;:")
				return tag
			}
		}
	}

	return ""
}

// cleanErrorMessage extracts the most relevant part of the error message
func cleanErrorMessage(errorMsg string) string {
	// Split by common separators and get the most specific error
	if idx := strings.LastIndex(errorMsg, ">"); idx != -1 && idx < len(errorMsg)-2 {
		// Get everything after the last ">"
		cleaned := strings.TrimSpace(errorMsg[idx+1:])
		if cleaned != "" {
			return cleaned
		}
	}

	// If no ">" found or result is empty, look for "Failed to" or "failed to"
	if idx := strings.Index(errorMsg, "failed to"); idx != -1 {
		return strings.TrimSpace(errorMsg[idx:])
	}

	if idx := strings.Index(errorMsg, "Failed to"); idx != -1 {
		return strings.TrimSpace(errorMsg[idx:])
	}

	// Return the original message if no better alternative found
	return strings.TrimSpace(errorMsg)
}
