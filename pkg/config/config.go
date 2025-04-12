package config

// ScanConfig holds configuration for the scanner.
type ScanConfig struct {
	IgnoreLists map[string][]string // Map OS name to list of libraries to ignore
	// Add other config options as needed
}

// DefaultConfig provides a basic default configuration.
func DefaultConfig() *ScanConfig {
	// Basic list for common Linux libs - THIS LIST IS NOT EXHAUSTIVE!
	// Needs refinement based on specific Ubuntu versions and common false positives.
	linuxStdLibs := []string{
		"linux-vdso.so.1",
		"libc.so.6",
		"libm.so.6",
		"libdl.so.2",
		"libpthread.so.0",
		"ld-linux-x86-64.so.2", // Specific to x86_64 glibc linker
		"ld-linux-aarch64.so.1", // Specific to aarch64 glibc linker
		"libgcc_s.so.1",
		"libstdc++.so.6",
		"librt.so.1", // Often part of glibc/core system
		// Add more common system/compiler runtime libs as needed
	}
	// Basic list for common macOS libs - Needs refinement!
	darwinStdLibs := []string{
		"/usr/lib/libSystem.B.dylib",
		"/usr/lib/libobjc.A.dylib",
		"/usr/lib/libc++.1.dylib",
		// Frameworks are often specified differently (@rpath, @executable_path, /System/Library/Frameworks/...)
		// Simple path filtering might be complex for frameworks.
		// Add more common low-level dylibs if needed.
	}

	return &ScanConfig{
		IgnoreLists: map[string][]string{
			"linux": linuxStdLibs,
			"darwin": darwinStdLibs, // <-- Add the macOS list
			// "windows": {"kernel32.dll", "user32.dll", ...}, // Future
		},
	}
}