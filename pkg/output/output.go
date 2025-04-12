package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Format takes the list of dependencies and formats it.
func Format(dependencies []string, formatType string) (string, error) {
	switch strings.ToLower(formatType) {
	case "list":
		return formatList(dependencies), nil
	case "json":
		return formatJSON(dependencies)
	// case "docker-apt":
	//  return formatDockerApt(dependencies) // Future implementation
	default:
		return "", fmt.Errorf("unsupported output format: %s", formatType)
	}
}

func formatList(dependencies []string) string {
	if len(dependencies) == 0 {
		return ""
	}
	return strings.Join(dependencies, "\n")
}

func formatJSON(dependencies []string) (string, error) {
	// Handle empty list explicitly for cleaner JSON [] instead of null
	if dependencies == nil {
		dependencies = []string{}
	}
	jsonData, err := json.MarshalIndent(dependencies, "", "  ") // Pretty print
	if err != nil {
		return "", fmt.Errorf("error marshalling dependencies to JSON: %w", err)
	}
	return string(jsonData), nil
}