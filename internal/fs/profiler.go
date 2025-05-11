package fs

import (
	"fmt"
	"github.com/auriora/onemount/pkg/logging"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

// ProfileType represents the type of profiling to perform
type ProfileType int

const (
	// ProfileCPU profiles CPU usage
	ProfileCPU ProfileType = iota
	// ProfileMemory profiles memory usage
	ProfileMemory
	// ProfileGoroutine profiles goroutine usage
	ProfileGoroutine
	// ProfileBlock profiles blocking operations
	ProfileBlock
	// ProfileMutex profiles mutex contention
	ProfileMutex
)

// Profiler manages profiling of the application
type Profiler struct {
	enabled      bool
	outputDir    string
	cpuProfile   *os.File
	memProfile   *os.File
	blockProfile *os.File
	mutexProfile *os.File
}

// NewProfiler creates a new profiler
func NewProfiler(outputDir string) *Profiler {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logging.Error().Err(err).Str("dir", outputDir).Msg("Failed to create profiling directory")
		return nil
	}

	return &Profiler{
		outputDir: outputDir,
		enabled:   false,
	}
}

// Start begins profiling of the specified type
func (p *Profiler) Start(profileType ProfileType) error {
	if p == nil {
		return fmt.Errorf("profiler is nil")
	}

	if p.enabled {
		return fmt.Errorf("profiling already in progress")
	}

	timestamp := time.Now().Format("20060102-150405")
	var err error

	switch profileType {
	case ProfileCPU:
		cpuFile := filepath.Join(p.outputDir, fmt.Sprintf("cpu-%s.pprof", timestamp))
		p.cpuProfile, err = os.Create(cpuFile)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %w", err)
		}

		if err := pprof.StartCPUProfile(p.cpuProfile); err != nil {
			p.cpuProfile.Close()
			return fmt.Errorf("could not start CPU profile: %w", err)
		}

		logging.Info().Str("file", cpuFile).Msg("Started CPU profiling")

	case ProfileBlock:
		// Enable block profiling
		runtime.SetBlockProfileRate(1)
		logging.Info().Msg("Started block profiling")

	case ProfileMutex:
		// Enable mutex profiling
		runtime.SetMutexProfileFraction(1)
		logging.Info().Msg("Started mutex profiling")

	default:
		return fmt.Errorf("unsupported profile type: %v", profileType)
	}

	p.enabled = true
	return nil
}

// Stop ends profiling and writes the results to disk
func (p *Profiler) Stop(profileType ProfileType) error {
	if p == nil {
		return fmt.Errorf("profiler is nil")
	}

	if !p.enabled {
		return fmt.Errorf("no profiling in progress")
	}

	timestamp := time.Now().Format("20060102-150405")
	var err error

	switch profileType {
	case ProfileCPU:
		if p.cpuProfile != nil {
			pprof.StopCPUProfile()
			p.cpuProfile.Close()
			p.cpuProfile = nil
			logging.Info().Msg("Stopped CPU profiling")
		}

	case ProfileMemory:
		memFile := filepath.Join(p.outputDir, fmt.Sprintf("memory-%s.pprof", timestamp))
		p.memProfile, err = os.Create(memFile)
		if err != nil {
			return fmt.Errorf("could not create memory profile: %w", err)
		}
		defer p.memProfile.Close()

		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(p.memProfile); err != nil {
			return fmt.Errorf("could not write memory profile: %w", err)
		}
		logging.Info().Str("file", memFile).Msg("Wrote memory profile")

	case ProfileGoroutine:
		goroutineFile := filepath.Join(p.outputDir, fmt.Sprintf("goroutine-%s.pprof", timestamp))
		f, err := os.Create(goroutineFile)
		if err != nil {
			return fmt.Errorf("could not create goroutine profile: %w", err)
		}
		defer f.Close()

		if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
			return fmt.Errorf("could not write goroutine profile: %w", err)
		}
		logging.Info().Str("file", goroutineFile).Msg("Wrote goroutine profile")

	case ProfileBlock:
		blockFile := filepath.Join(p.outputDir, fmt.Sprintf("block-%s.pprof", timestamp))
		f, err := os.Create(blockFile)
		if err != nil {
			return fmt.Errorf("could not create block profile: %w", err)
		}
		defer f.Close()

		if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
			return fmt.Errorf("could not write block profile: %w", err)
		}
		runtime.SetBlockProfileRate(0) // Disable block profiling
		logging.Info().Str("file", blockFile).Msg("Wrote block profile")

	case ProfileMutex:
		mutexFile := filepath.Join(p.outputDir, fmt.Sprintf("mutex-%s.pprof", timestamp))
		f, err := os.Create(mutexFile)
		if err != nil {
			return fmt.Errorf("could not create mutex profile: %w", err)
		}
		defer f.Close()

		if err := pprof.Lookup("mutex").WriteTo(f, 0); err != nil {
			return fmt.Errorf("could not write mutex profile: %w", err)
		}
		runtime.SetMutexProfileFraction(0) // Disable mutex profiling
		logging.Info().Str("file", mutexFile).Msg("Wrote mutex profile")

	default:
		return fmt.Errorf("unsupported profile type: %v", profileType)
	}

	p.enabled = false
	return nil
}

// CaptureProfile captures a profile of the specified type
func (p *Profiler) CaptureProfile(profileType ProfileType, duration time.Duration) error {
	if err := p.Start(profileType); err != nil {
		return err
	}

	time.Sleep(duration)

	return p.Stop(profileType)
}

// CaptureAllProfiles captures all available profiles
func (p *Profiler) CaptureAllProfiles(duration time.Duration) error {
	// Start CPU profiling
	if err := p.Start(ProfileCPU); err != nil {
		return err
	}

	// Start block profiling
	runtime.SetBlockProfileRate(1)

	// Start mutex profiling
	runtime.SetMutexProfileFraction(1)

	// Wait for the specified duration
	time.Sleep(duration)

	// Stop CPU profiling
	pprof.StopCPUProfile()
	if p.cpuProfile != nil {
		p.cpuProfile.Close()
		p.cpuProfile = nil
	}

	// Capture memory profile
	if err := p.Stop(ProfileMemory); err != nil {
		logging.Error().Err(err).Msg("Failed to capture memory profile")
	}

	// Capture goroutine profile
	if err := p.Stop(ProfileGoroutine); err != nil {
		logging.Error().Err(err).Msg("Failed to capture goroutine profile")
	}

	// Capture block profile
	if err := p.Stop(ProfileBlock); err != nil {
		logging.Error().Err(err).Msg("Failed to capture block profile")
	}

	// Capture mutex profile
	if err := p.Stop(ProfileMutex); err != nil {
		logging.Error().Err(err).Msg("Failed to capture mutex profile")
	}

	return nil
}
