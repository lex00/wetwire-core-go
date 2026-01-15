package scenario

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrTermsvgNotFound is returned when termsvg is not installed.
var ErrTermsvgNotFound = errors.New("termsvg not found: install with 'go install github.com/mrmarble/termsvg/cmd/termsvg@latest'")

// DefaultGreeting is the default greeting shown at the start of recordings.
const DefaultGreeting = `$ wetwire scenario run %s

Loading scenario...
`

// RecorderConfig configures the scenario recorder.
type RecorderConfig struct {
	// OutputDir is the directory where recordings are saved
	OutputDir string

	// ScenarioName is the name of the scenario being recorded
	ScenarioName string

	// Format is the output format (default: svg)
	Format string

	// Greeting is shown at the start of the recording to provide context.
	// Use %s as placeholder for scenario name. Empty string disables greeting.
	// If not set, DefaultGreeting is used.
	Greeting string

	// GreetingDelay is the pause after greeting before showing output (default: 1s)
	GreetingDelay time.Duration

	// TermWidth is the terminal width in characters (default: 80)
	TermWidth int

	// TermHeight is the terminal height in characters (default: 24)
	TermHeight int

	// LineDelay is the minimum delay between output lines (default: 0.3s)
	LineDelay time.Duration
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

// Record executes the given function while recording output to SVG.
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

	// Capture the function's output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	os.Stdout = writePipe

	// Copy output in background
	done := make(chan error)
	go func() {
		_, err := io.Copy(&buf, readPipe)
		done <- err
	}()

	// Execute the function
	startTime := time.Now()
	fnErr := fn()
	duration := time.Since(startTime)

	// Restore stdout and close pipe
	os.Stdout = oldStdout
	_ = writePipe.Close()
	<-done
	_ = readPipe.Close()

	// Generate asciinema cast file
	if err := r.generateCastFile(castPath, buf.String(), duration); err != nil {
		return fmt.Errorf("failed to generate cast file: %w", err)
	}

	// Export to SVG using termsvg
	exportCmd := exec.Command("termsvg", "export", castPath, "-o", svgPath)
	if output, err := exportCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to export SVG: %w: %s", err, output)
	}

	// Clean up intermediate cast file
	r.Cleanup()

	return fnErr
}

// generateCastFile creates an asciinema v2 cast file from captured output.
func (r *Recorder) generateCastFile(path string, output string, duration time.Duration) error {
	var buf bytes.Buffer

	// Apply defaults for terminal dimensions
	termWidth := r.config.TermWidth
	if termWidth == 0 {
		termWidth = 80
	}
	termHeight := r.config.TermHeight
	if termHeight == 0 {
		termHeight = 24
	}

	// Write header (asciinema v2 format)
	header := fmt.Sprintf(`{"version": 2, "width": %d, "height": %d, "timestamp": %d, "title": "%s"}`,
		termWidth, termHeight, time.Now().Unix(), r.config.ScenarioName)
	buf.WriteString(header)
	buf.WriteString("\n")

	currentTime := 0.0

	// Add greeting if configured
	greeting := r.config.Greeting
	if greeting == "" {
		greeting = fmt.Sprintf(DefaultGreeting, r.config.ScenarioName)
	}

	greetingDelay := r.config.GreetingDelay
	if greetingDelay == 0 {
		greetingDelay = time.Second
	}

	// Write greeting lines with typing effect
	greetingLines := strings.Split(greeting, "\n")
	for _, line := range greetingLines {
		escapedLine := escapeJSON(line + "\r\n")
		event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, escapedLine)
		buf.WriteString(event)
		buf.WriteString("\n")
		currentTime += 0.05 // Fast typing for greeting
	}

	// Pause after greeting
	currentTime += greetingDelay.Seconds()

	// Write output lines with timing
	lines := strings.Split(output, "\n")

	// Filter out empty lines
	var nonEmptyLines []string
	for _, line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	// If no output, add a placeholder
	if len(nonEmptyLines) == 0 {
		nonEmptyLines = []string{"(no output)"}
	}

	// Apply line delay (default 0.3s)
	lineDelay := r.config.LineDelay
	if lineDelay == 0 {
		lineDelay = 300 * time.Millisecond
	}
	timePerLine := lineDelay.Seconds()

	for _, line := range nonEmptyLines {
		// Escape the line for JSON
		escapedLine := escapeJSON(line + "\r\n")
		event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, escapedLine)
		buf.WriteString(event)
		buf.WriteString("\n")
		currentTime += timePerLine
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// escapeJSON escapes a string for use in JSON.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
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

	// Greeting is shown at the start of the recording.
	// Use %s as placeholder for scenario name.
	// Leave empty to use DefaultGreeting, set to " " to disable.
	Greeting string

	// GreetingDelay is the pause after greeting (default: 1s)
	GreetingDelay time.Duration

	// TermWidth is the terminal width in characters (default: 80)
	TermWidth int

	// TermHeight is the terminal height in characters (default: 24)
	TermHeight int

	// LineDelay is the minimum delay between output lines (default: 0.3s)
	LineDelay time.Duration
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
		OutputDir:     opts.OutputDir,
		ScenarioName:  name,
		Greeting:      opts.Greeting,
		GreetingDelay: opts.GreetingDelay,
		TermWidth:     opts.TermWidth,
		TermHeight:    opts.TermHeight,
		LineDelay:     opts.LineDelay,
	}

	recorder := NewRecorder(config)
	return recorder.Record(fn)
}
