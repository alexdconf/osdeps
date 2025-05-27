package analyzer

import (
	"log"
	"runtime"
	"sync"

	"github.com/alexdconf/osdeps/pkg/config"
	"github.com/alexdconf/osdeps/pkg/parser"
)

// WorkerPool manages a pool of worker goroutines for dependency analysis
type WorkerPool struct {
	artifactPaths []string
	cfg           *config.ScanConfig
	targetOS      string
	workers       int
	depsChan      chan []string
	wg            sync.WaitGroup
}

// NewWorkerPool creates a new WorkerPool
func NewWorkerPool(artifactPaths []string, cfg *config.ScanConfig, targetOS string, workers int) *WorkerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	// Ensure we don't have more workers than artifacts
	if workers > len(artifactPaths) {
		workers = len(artifactPaths)
	}
	return &WorkerPool{
		artifactPaths: artifactPaths,
		cfg:          cfg,
		targetOS:     targetOS,
		workers:      workers,
		depsChan:     make(chan []string, workers),
	}
}

// Start starts the worker pool and returns a channel of dependencies
func (wp *WorkerPool) Start() <-chan []string {
	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Close channel when all workers are done
	go func() {
		wp.wg.Wait()
		close(wp.depsChan)
	}()

	return wp.depsChan
}

// worker processes artifacts and sends dependencies to the channel
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	// Calculate which artifacts this worker should process
	start := (id * len(wp.artifactPaths)) / wp.workers
	end := ((id + 1) * len(wp.artifactPaths)) / wp.workers
	if end > len(wp.artifactPaths) {
		end = len(wp.artifactPaths)
	}

	// Process artifacts
	for i := start; i < end; i++ {
		artifactPath := wp.artifactPaths[i]
		log.Printf("Worker %d processing %s", id, artifactPath)

		parser := parser.NewMachOParser()
		deps, err := parser.ParseDependencies(artifactPath)
		if err != nil {
			log.Printf("Error parsing %s: %v", artifactPath, err)
			continue
		}

		// Analyze dependencies
		analyzer := NewDependencyAnalyzer(wp.cfg, wp.targetOS)
		uniqueDeps := analyzer.AnalyzeDependencies([][]string{deps}, []string{artifactPath})

		// Send dependencies to channel
		wp.depsChan <- uniqueDeps
	}
}
