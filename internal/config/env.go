package config

import (
	"os"
	"strings"
)

func InternalAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("INTERNAL_API_KEY")); value != "" {
		return value
	}

	return strings.TrimSpace(os.Getenv("PDF_INTERNAL_API_KEY"))
}
