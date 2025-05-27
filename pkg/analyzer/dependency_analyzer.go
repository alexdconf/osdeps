package analyzer

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/alexdconf/osdeps/pkg/config"
	"github.com/alexdconf/osdeps/pkg/parser"
)

// DependencyAnalyzer analyzes dependencies from parsed artifacts
type DependencyAnalyzer struct {
	Config *config.ScanConfig
	TargetOS string
}

// NewDependencyAnalyzer creates a new DependencyAnalyzer
func NewDependencyAnalyzer(cfg *config.ScanConfig, targetOS string) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		Config:   cfg,
		TargetOS: targetOS,
	}
}

// AnalyzeDependencies analyzes and returns all dependencies from parsed artifacts
func (da *DependencyAnalyzer) AnalyzeDependencies(rawDeps [][]string, artifactPaths []string) []string {
	var allDeps []string
	var seenDeps = make(map[string]bool)

	// Process each artifact's dependencies
	for _, deps := range rawDeps {
		for _, dep := range deps {
			if !seenDeps[dep] {
				seenDeps[dep] = true
				allDeps = append(allDeps, dep)
			}
		}
	}

	// For rpath dependencies, try to find the actual file
	for i, dep := range allDeps {
		if strings.HasPrefix(dep, "@rpath/") {
			for _, artifactPath := range artifactPaths {
				// Get the directory of the artifact
				artifactDir := filepath.Dir(artifactPath)
				// Try to find the library in the same directory
				libPath := filepath.Join(artifactDir, strings.TrimPrefix(dep, "@rpath/"))
				if _, err := parser.NewMachOParser().ParseDependencies(libPath); err == nil {
					allDeps[i] = libPath
					break
				}
			}
		}
	}

	// Sort dependencies
	sort.Strings(allDeps)
	return allDeps
}
