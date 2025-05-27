package config

// ScanConfig holds configuration for the scanner.
type ScanConfig struct {
	// Add other config options as needed
}

// DefaultConfig provides a basic default configuration.
func DefaultConfig() *ScanConfig {
	return &ScanConfig{}
}