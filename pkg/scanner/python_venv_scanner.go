package scanner

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime" // Used only for initial OS detection in scanner
	"strings"
)

type PythonVenvScanner struct {
	TargetOS string // OS we are looking for artifacts for
}

// NewPythonVenvScanner creates a scanner for Python venvs.
func NewPythonVenvScanner(targetOS string) *PythonVenvScanner {
	return &PythonVenvScanner{TargetOS: targetOS}
}

// Scan finds compiled artifacts (.so for Linux) in a Python venv.
func (s *PythonVenvScanner) Scan(envPath string) ([]Artifact, error) {
	var artifacts []Artifact
	sitePackagesPaths, err := findSitePackages(envPath)
	if err != nil {
		return nil, fmt.Errorf("could not find site-packages in %s: %w", envPath, err)
	}
	if len(sitePackagesPaths) == 0 {
		return nil, fmt.Errorf("no site-packages directory found in %s or its subdirectories", envPath)
	}

	log.Printf("Found site-packages directories: %v", sitePackagesPaths)

	var targetSuffix []string
	switch s.TargetOS {
	case "linux":
		targetSuffix = append(targetSuffix, ".so")
	case "darwin":
		targetSuffix = append(targetSuffix, ".so", ".dylib") // or .dylib? Python extensions are often .so even on macOS
	// case "windows":
	//  targetSuffix = ".pyd" // or .dll? Python extensions are .pyd
	default:
		return nil, fmt.Errorf("unsupported target OS for Python venv scan: %s", s.TargetOS)
	}

	processedPaths := make(map[string]bool) // Avoid processing the same file multiple times if found via different site-packages paths

	for _, spPath := range sitePackagesPaths {
		log.Printf("Scanning %s...", spPath)
		err := filepath.WalkDir(spPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Log errors accessing specific paths but continue walking
				log.Printf("Warning: Error accessing path %s: %v", path, err)
				return nil // Continue walking if possible
			}
			// Check if it's a file and has the target suffix (e.g., .so)
			// Also check if it's likely executable/linkable (basic check for regular file)
			var isSuffix bool = false
			if !d.IsDir() {
				for _, suffix := range targetSuffix {
					if strings.HasSuffix(d.Name(), suffix) {
						isSuffix = true
						break
					}
				}
			}
			if isSuffix {
				absPath, _ := filepath.Abs(path) // Ignore error as path comes from WalkDir
				if !processedPaths[absPath] {
					// Basic check: ensure it's a regular file before adding
					info, infoErr := d.Info()
					if infoErr == nil && info.Mode().IsRegular() {
						// Add artifact - Arch detection could be added here or by parser
						artifacts = append(artifacts, Artifact{
							Path: absPath,
							Type: PythonExtensionSO,
							OS:   s.TargetOS, // Assume artifact matches target OS for now
							Arch: runtime.GOARCH, // Use current arch as placeholder, parser should confirm
						})
						processedPaths[absPath] = true
					} else if infoErr != nil {
						log.Printf("Warning: Could not get info for %s: %v", path, infoErr)
					}
				}
			}
			return nil // Continue walking
		})
		if err != nil {
			// Log error during walk but potentially continue with other site-package paths
			log.Printf("Warning: Error walking directory %s: %v", spPath, err)
		}
	}

	return artifacts, nil
}

// findSitePackages tries to locate site-packages directories within a venv path.
// This is a simplified heuristic and might need improvement.
func findSitePackages(envPath string) ([]string, error) {
	var paths []string
	// Common patterns: lib/pythonX.Y/site-packages
	libPath := filepath.Join(envPath, "lib")
	entries, err := os.ReadDir(libPath)
	if err != nil {
		// Check Lib for Windows venv structure
		libPath = filepath.Join(envPath, "Lib")
		entries, err = os.ReadDir(libPath)
		if err != nil {
			return nil, fmt.Errorf("could not read 'lib' or 'Lib' directory in %s: %w", envPath, err)
		}
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "python") {
			spPath := filepath.Join(libPath, entry.Name(), "site-packages")
			if info, err := os.Stat(spPath); err == nil && info.IsDir() {
				paths = append(paths, spPath)
			}
		}
	}

	// Also check top-level site-packages (less common for venv but possible)
	spPathTop := filepath.Join(envPath, "site-packages")
	if info, err := os.Stat(spPathTop); err == nil && info.IsDir() {
		paths = append(paths, spPathTop)
	}


	if len(paths) == 0 {
		// Fallback or further search logic could go here
		log.Printf("Warning: Could not find site-packages using common patterns in %s.", envPath)
		// As a simple fallback, maybe scan the whole envPath? Risky.
		// return []string{envPath}, nil // Or return error as above
	}

	return paths, nil
}