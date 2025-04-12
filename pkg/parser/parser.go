package parser

// Parser defines the interface for artifact parsers.
type Parser interface {
	// ParseDependencies extracts the names of required shared libraries.
	ParseDependencies(artifactPath string) ([]string, error)
}