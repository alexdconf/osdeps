package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

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
	workers := flag.Int("workers", -1, "Number of worker goroutines to use for dependency analysis (-1 for auto-detect)")
	debug := flag.Bool("debug", false, "Show all dependencies before filtering")
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
	}

	log.Printf("Found %d potential artifacts.", len(artifacts))

	// Parse artifacts for dependencies
	log.Println("Parsing artifacts for dependencies...")
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
		artifactParser = parser.NewMachOParser()
	default:
		log.Fatalf("Error: Unsupported OS '%s'", *targetOS)
		os.Exit(1)
	}

	allRawDependencies := make([][]string, 0)

	for _, artifact := range artifacts {
		deps, err := artifactParser.ParseDependencies(artifact.Path)
		if err != nil {
			log.Printf("Error parsing %s: %v", artifact.Path, err)
			continue
		}
		allRawDependencies = append(allRawDependencies, deps)
	}

	log.Printf("Successfully parsed %d artifacts, skipped %d.", len(artifacts), len(artifacts)-len(allRawDependencies))

	// Analyze dependencies using worker pool
	log.Println("Analyzing dependencies...")
	artifactPaths := make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactPaths[i] = artifact.Path
	}
	wp := analyzer.NewWorkerPool(artifactPaths, config.DefaultConfig(), *targetOS, *workers)
	depsChan := wp.Start()

	// Collect dependencies from all workers and deduplicate
	depsMap := make(map[string]bool)
	for deps := range depsChan {
		for _, dep := range deps {
			depsMap[dep] = true
		}
	}

	// Convert map keys back to slice
	allDeps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		allDeps = append(allDeps, dep)
	}

	sort.Strings(allDeps)

	log.Printf("Found %d unique dependencies.", len(allDeps))

	// Print dependencies
	log.Println("\nDependencies:")
	for _, dep := range allDeps {
		log.Printf("- %s", dep)
	}

	// Format output
	log.Println("Formatting output...")

	// Format and print dependencies
	fmt.Println("\n--- OS Dependencies ---")
	if len(allDeps) > 0 {
		outputString, err := output.Format(allDeps, *outputFormat)
		if err != nil {
			log.Fatalf("Error formatting output: %v", err)
			os.Exit(1)
		}
		fmt.Println(outputString)
	} else {
		fmt.Println("(None identified)")
	}

	// If debug mode is enabled, show all raw dependencies
	if *debug {
		fmt.Println("\n--- All Raw Dependencies ---")
		var allRawDeps []string
		for _, deps := range allRawDependencies {
			allRawDeps = append(allRawDeps, deps...)
		}
		sort.Strings(allRawDeps)
		outputString, err := output.Format(allRawDeps, *outputFormat)
		if err != nil {
			log.Fatalf("Error formatting output: %v", err)
			os.Exit(1)
		}
		fmt.Println(outputString)
	}

	log.Println("Scan complete.")
}
