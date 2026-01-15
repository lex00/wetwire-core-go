package scenario

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrTermsvgNotFound is returned when termsvg is not installed.
var ErrTermsvgNotFound = errors.New("termsvg not found: install with 'go install github.com/mrmarble/termsvg/cmd/termsvg@latest'")

// RecorderConfig configures the scenario recorder.
type RecorderConfig struct {
	// OutputDir is the directory where recordings are saved
	OutputDir string

	// ScenarioName is the name of the scenario being recorded
	ScenarioName string

	// Format is the output format (default: svg)
	Format string
}

// Recorder records scenario execution to SVG using termsvg.
type Recorder struct {
	config RecorderConfig
}

// NewRecorder creates a new Recorder with the given config.
func NewRecorder(config RecorderConfig) *Recorder {
	if config.Format == "" {
		config.Format = "svg"
	}
	return &Recorder{config: config}
}

// CanRecord returns true if termsvg is available on the system.
func CanRecord() bool {
	_, err := exec.LookPath("termsvg")
	return err == nil
}

// OutputPath returns the path to the output SVG file.
func (r *Recorder) OutputPath() string {
	return filepath.Join(r.config.OutputDir, r.config.ScenarioName+".svg")
}

// CastPath returns the path to the intermediate .cast file.
func (r *Recorder) CastPath() string {
	return filepath.Join(r.config.OutputDir, r.config.ScenarioName+".cast")
}

// Record executes the given function while recording terminal output to SVG.
// If termsvg is not installed, returns ErrTermsvgNotFound.
func (r *Recorder) Record(fn func() error) error {
	if !CanRecord() {
		return ErrTermsvgNotFound
	}

	if err := r.ensureOutputDir(); err != nil {
		return err
	}

	castPath := r.CastPath()
	svgPath := r.OutputPath()

	// Start recording with termsvg
	// termsvg rec <output.cast> -- <command>
	// For our use case, we'll use a wrapper approach

	// Create a temporary script to run the function
	// This is a simplified approach - in production you might use PTY
	cmd := exec.Command("termsvg", "rec", castPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the recording
	if err := cmd.Start(); err != nil {
		return err
	}

	// Execute the scenario function
	fnErr := fn()

	// Wait for recording to finish
	_ = cmd.Wait()

	// Export to SVG
	exportCmd := exec.Command("termsvg", "export", castPath, "-o", svgPath)
	if err := exportCmd.Run(); err != nil {
		return err
	}

	// Clean up intermediate cast file
	r.Cleanup()

	return fnErr
}

// ensureOutputDir creates the output directory if it doesn't exist.
func (r *Recorder) ensureOutputDir() error {
	return os.MkdirAll(r.config.OutputDir, 0755)
}

// Cleanup removes intermediate files (cast files).
func (r *Recorder) Cleanup() {
	castPath := r.CastPath()
	_ = os.Remove(castPath)
}

// RecordToSVG is a convenience function to record scenario execution to SVG.
// It handles all setup and cleanup automatically.
func RecordToSVG(outputPath string, fn func() error) error {
	dir := filepath.Dir(outputPath)
	name := filepath.Base(outputPath)

	// Remove .svg extension from name
	if len(name) > 4 && name[len(name)-4:] == ".svg" {
		name = name[:len(name)-4]
	}

	config := RecorderConfig{
		OutputDir:    dir,
		ScenarioName: name,
	}

	recorder := NewRecorder(config)
	return recorder.Record(fn)
}

// RecordOptions contains options for recording a scenario.
type RecordOptions struct {
	// Enabled determines if recording should be attempted
	Enabled bool

	// OutputDir is where recordings are saved (default: ./recordings)
	OutputDir string

	// GracefulFallback if true, continues without recording if termsvg unavailable
	GracefulFallback bool
}

// RunWithRecording runs a scenario with optional recording.
// If opts.Enabled is false or termsvg is unavailable (and GracefulFallback is true),
// it simply runs the function without recording.
func RunWithRecording(name string, opts RecordOptions, fn func() error) error {
	if !opts.Enabled {
		return fn()
	}

	if !CanRecord() {
		if opts.GracefulFallback {
			return fn()
		}
		return ErrTermsvgNotFound
	}

	if opts.OutputDir == "" {
		opts.OutputDir = "./recordings"
	}

	config := RecorderConfig{
		OutputDir:    opts.OutputDir,
		ScenarioName: name,
	}

	recorder := NewRecorder(config)
	return recorder.Record(fn)
}
