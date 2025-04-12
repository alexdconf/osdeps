package parser

import (
	"debug/elf"
	"fmt"
	"strings"
)

// ELFParser implements the Parser interface for ELF files.
type ELFParser struct{}

// NewELFParser creates a new ELF parser.
func NewELFParser() *ELFParser {
	return &ELFParser{}
}

// ParseDependencies extracts DT_NEEDED entries from an ELF file.
func (p *ELFParser) ParseDependencies(artifactPath string) ([]string, error) {
	f, err := elf.Open(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("error opening ELF file %s: %w", artifactPath, err)
	}
	defer f.Close()

	// Use ImportedLibraries which reads DT_NEEDED entries
	libs, err := f.ImportedLibraries()
	if err != nil {
		// This can happen if there's no dynamic section, which is fine (static lib?)
		// Check for specific error types if needed, otherwise treat as no deps
		if _, ok := err.(*elf.FormatError); ok && strings.Contains(err.Error(), "no dynamic section") {
             // log.Printf("Debug: No dynamic section in %s", filepath.Base(artifactPath))
			 return []string{}, nil // No dynamic dependencies found
		}
		// Treat other errors as actual parsing problems
		return nil, fmt.Errorf("error reading imported libraries from %s: %w", artifactPath, err)
	}

	// TODO: Could also extract f.Machine here to populate Artifact.Arch accurately

	return libs, nil
}