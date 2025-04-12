package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alexdconf/osdeps/pkg/analyzer"
	"github.com/alexdconf/osdeps/pkg/config"
	"github.com/alexdconf/osdeps/pkg/output"
	"github.com/alexdconf/osdeps/pkg/parser"
	"github.com/alexdconf/osdeps/pkg/scanner"
)

// Base directory for packages relative to the main file.
// Adjust if your project structure is different.
const basePkgDir = "pkg"

func main() {
	// --- CLI Flag Parsing ---
	envPath := flag.String("env-path", "", "Path to the environment directory (e.g., Python venv)")
	envType := flag.String("env-type", "python-venv", "Type of environment ('python-venv')")
	targetOS := flag.String("os", "linux", "Target operating system ('linux')")
	outputFormat := flag.String("output-format", "list", "Output format ('list', 'json')")
	filterLevel := flag.String("filter-level", "basic", "Level of standard library filtering ('basic', 'none')")
	// Add flags for architecture if needed in the future, e.g., targetArch

	flag.Parse()

	// Basic validation
	if *envPath == "" {
		log.Fatal("Error: --env-path flag is required")
		os.Exit(1)
	}
	absEnvPath, err := filepath.Abs(*envPath)
	if err != nil {
		log.Fatalf("Error getting absolute path for --env-path: %v", err)
		os.Exit(1)
	}

	log.Printf("Starting scan for environment: %s (Type: %s, OS: %s)", absEnvPath, *envType, *targetOS)

	// --- Configuration ---
	// Load configuration (including filter lists)
	// For now, using a default basic filter for Linux
	cfg := config.DefaultConfig()
	if *filterLevel == "none" {
		cfg.IgnoreLists[*targetOS] = []string{} // Clear ignore list if filter level is none
	} else if _, ok := cfg.IgnoreLists[*targetOS]; !ok {
		log.Printf("Warning: No basic filter list defined for OS '%s'. No filtering applied.", *targetOS)
		cfg.IgnoreLists[*targetOS] = []string{}
	}

	// --- Select Scanner ---
	var envScanner scanner.Scanner
	switch *envType {
	case "python-venv":
		// Construct path relative to where the executable might be run from
		// This assumes a standard Go project layout where main.go is in cmd/scanner-cli
		// and packages are in pkg/
		// A more robust solution might involve embedding or better path resolution.
		scannerPath := filepath.Join(basePkgDir, "scanner", "python_venv_scanner.go") // Placeholder path
		log.Printf("Using PythonVenvScanner (source assumed near %s)", scannerPath)
		envScanner = scanner.NewPythonVenvScanner(*targetOS) // Pass target OS
	// case "node-modules":
	//  envScanner = scanner.NewNodeModulesScanner(*targetOS) // Future implementation
	default:
		log.Fatalf("Error: Unsupported environment type '%s'", *envType)
		os.Exit(1)
	}

	// --- Scan Environment ---
	log.Println("Scanning environment for artifacts...")
	artifacts, err := envScanner.Scan(absEnvPath)
	if err != nil {
		log.Fatalf("Error scanning environment: %v", err)
		os.Exit(1)
	}
	if len(artifacts) == 0 {
		log.Println("No compiled artifacts found to analyze.")
		os.Exit(0)
	}
	log.Printf("Found %d potential artifacts.", len(artifacts))

	// --- Select Parser & Parse Artifacts ---
	log.Println("Parsing artifacts for dependencies...")
	allRawDependencies := make([][]string, 0, len(artifacts))
	var artifactParser parser.Parser

	// Select parser based on OS (extend this for cross-platform)
	switch *targetOS {
	case "linux":
		// Construct path relative to where the executable might be run from
		parserPath := filepath.Join(basePkgDir, "parser", "elf_parser.go") // Placeholder path
		log.Printf("Using ELFParser (source assumed near %s)", parserPath)
		artifactParser = parser.NewELFParser()
	case "darwin": // <-- Add this case for macOS
		parserPath := filepath.Join(basePkgDir, "parser", "macho_parser.go") // Placeholder path
		log.Printf("Using MachOParser (source assumed near %s)", parserPath)
		artifactParser = parser.NewMachOParser() // Use the new parser
	// case "windows":
	//  artifactParser = parser.NewPEParser() // Future implementation
	default:
		log.Fatalf("Error: Unsupported OS '%s' for parsing", *targetOS)
		os.Exit(1)
	}

	// Parse each artifact
	parsedCount := 0
	skippedCount := 0
	for _, artifact := range artifacts {
		// Basic check if artifact OS matches target OS (can be refined)
		if artifact.OS != *targetOS {
			log.Printf("Skipping artifact %s (OS mismatch: %s != %s)", filepath.Base(artifact.Path), artifact.OS, *targetOS)
			skippedCount++
			continue
		}

		// log.Printf("Parsing: %s", filepath.Base(artifact.Path)) // Can be verbose
		deps, err := artifactParser.ParseDependencies(artifact.Path)
		if err != nil {
			// Log non-fatal parsing errors (e.g., wrong format, permissions)
			// Don't stop the whole process unless it's critical
			log.Printf("Warning: Could not parse dependencies for %s: %v", filepath.Base(artifact.Path), err)
			continue // Skip this artifact
		}
		if len(deps) > 0 {
			allRawDependencies = append(allRawDependencies, deps)
			parsedCount++
		}
	}
	log.Printf("Successfully parsed %d artifacts, skipped %d.", parsedCount, skippedCount)

	// --- Analyze Dependencies ---
	log.Println("Analyzing and filtering dependencies...")
	// Construct path relative to where the executable might be run from
	analyzerPath := filepath.Join(basePkgDir, "analyzer", "analyzer.go") // Placeholder path
	log.Printf("Using Analyzer (source assumed near %s)", analyzerPath)
	finalDependencies := analyzer.AnalyzeDependencies(allRawDependencies, cfg, *targetOS)
	log.Printf("Found %d unique, non-standard OS dependencies.", len(finalDependencies))

	// --- Output Results ---
	log.Println("Formatting output...")
	// Construct path relative to where the executable might be run from
	outputPath := filepath.Join(basePkgDir, "output", "output.go") // Placeholder path
	log.Printf("Using Output formatter (source assumed near %s)", outputPath)
	outputString, err := output.Format(finalDependencies, *outputFormat)
	if err != nil {
		log.Fatalf("Error formatting output: %v", err)
		os.Exit(1)
	}

	// Print final result to stdout
	fmt.Println("\n--- Required OS Libraries ---")
	if outputString == "" && *outputFormat == "list" {
		fmt.Println("(None identified)")
	} else {
		fmt.Println(outputString)
	}
	log.Println("Scan complete.")
}
