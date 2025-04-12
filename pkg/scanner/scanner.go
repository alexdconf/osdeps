package scanner

// Artifact represents a compiled file found within an environment.
type Artifact struct {
	Path string       // Absolute path to the artifact
	Type ArtifactType // Type of artifact (e.g., PythonExtensionSO)
	OS   string       // Target OS (e.g., "linux")
	Arch string       // Target Arch (e.g., "amd64")
}

// ArtifactType defines the kind of compiled artifact.
type ArtifactType string

const (
	PythonExtensionSO ArtifactType = "python-ext-so"
	// Future: NodeAddonNode, MachOExecutable, PEExecutable, etc.
)

// Scanner defines the interface for environment scanners.
type Scanner interface {
	// Scan traverses the environment path and returns discovered artifacts.
	Scan(envPath string) ([]Artifact, error)
}