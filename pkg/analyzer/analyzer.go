package analyzer

import (
	"log"
	"github.com/alexdconf/osdeps/pkg/config" // Use alias if needed due to package name conflict
	"sort"
)

// AnalyzeDependencies filters and deduplicates the raw dependency lists.
func AnalyzeDependencies(rawDeps [][]string, cfg *config.ScanConfig, targetOS string) []string {
	if cfg == nil {
		log.Println("Warning: Analyzer received nil config, using empty defaults.")
		cfg = &config.ScanConfig{IgnoreLists: make(map[string][]string)}
	}
	// Flatten the list of lists into a single list
	flatDeps := []string{}
	for _, depsList := range rawDeps {
		flatDeps = append(flatDeps, depsList...)
	}

	// Create map for unique dependencies and filtering
	uniqueDeps := make(map[string]bool)
	ignoreMap := make(map[string]bool)

	// Populate ignore map for the target OS
	if ignoreList, ok := cfg.IgnoreLists[targetOS]; ok {
		for _, lib := range ignoreList {
			ignoreMap[lib] = true
		}
	} else {
		log.Printf("Analyzer: No ignore list found for OS '%s'", targetOS)
	}

	// Filter and deduplicate
	for _, dep := range flatDeps {
		if _, shouldIgnore := ignoreMap[dep]; !shouldIgnore {
			uniqueDeps[dep] = true
		}
	}

	// Convert map keys back to a slice
	finalList := make([]string, 0, len(uniqueDeps))
	for dep := range uniqueDeps {
		finalList = append(finalList, dep)
	}

	// Sort for consistent output
	sort.Strings(finalList)

	return finalList
}