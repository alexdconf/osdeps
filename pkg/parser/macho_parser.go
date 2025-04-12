package parser

import (
	"debug/macho"
	"fmt"
)

// MachOParser implements the Parser interface for Mach-O files (macOS).
type MachOParser struct{}

// NewMachOParser creates a new Mach-O parser.
func NewMachOParser() *MachOParser {
	return &MachOParser{}
}

// ParseDependencies extracts dynamically linked library paths (LC_LOAD_DYLIB, etc.)
// from a Mach-O file.
func (p *MachOParser) ParseDependencies(artifactPath string) ([]string, error) {
	// Open the Mach-O file
	machoFile, err := macho.Open(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("error opening Mach-O file %s: %w", artifactPath, err)
	}
	defer machoFile.Close() // Ensure the file is closed

	var dependencies []string

	// Iterate through the load commands in the Mach-O file
	for _, loadCmd := range machoFile.Loads {
		// Check if the load command is one of the types that imports a dynamic library.
		// The macho.Dylib struct represents commands like LC_LOAD_DYLIB,
		// LC_LOAD_WEAK_DYLIB, LC_REEXPORT_DYLIB, etc.
		if dylibCmd, ok := loadCmd.(*macho.Dylib); ok {
			// dylibCmd.Name contains the path to the linked library
			// (e.g., /usr/lib/libSystem.B.dylib or @rpath/MyFramework.framework/Versions/A/MyFramework)
			dependencies = append(dependencies, dylibCmd.Name)
		}
	}

	// Note: machoFile.ImportedLibraries() is specific to ELF and does not exist for Mach-O.
	// We must iterate through the load commands as shown above.

	return dependencies, nil
}