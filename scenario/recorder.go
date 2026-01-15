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

// DefaultAgentGreeting is shown first - the agent asking how to help.
const DefaultAgentGreeting = `How can I help you today?

> `

// DefaultUserPrompt should be set to the actual prompt content.
const DefaultUserPrompt = ``

// DefaultAgentResponse is shown after user prompt, before output.
const DefaultAgentResponse = `

`

// RecorderConfig configures the scenario recorder.
type RecorderConfig struct {
	// OutputDir is the directory where recordings are saved
	OutputDir string

	// ScenarioName is the name of the scenario being recorded
	ScenarioName string

	// Format is the output format (default: svg)
	Format string

	// AgentGreeting is shown first, before the user types.
	// Set to " " to disable. If not set, DefaultAgentGreeting is used.
	AgentGreeting string

	// UserPrompt is what the user types (with typing simulation).
	// Use %s as placeholder for scenario name. Set to " " to disable.
	// If not set, DefaultUserPrompt is used.
	UserPrompt string

	// AgentResponse is shown after the user's prompt, before scenario output.
	// Use %s as placeholder for scenario name. Set to " " to disable.
	// If not set, DefaultAgentResponse is used.
	AgentResponse string

	// ResponseDelay is the pause after agent response before showing output (default: 500ms)
	ResponseDelay time.Duration

	// TermWidth is the terminal width in characters (default: 80)
	TermWidth int

	// TermHeight is the terminal height in characters (default: 24)
	TermHeight int

	// LineDelay is the minimum delay between output lines (default: 0.3s)
	LineDelay time.Duration

	// TypingSpeed is the delay between characters when simulating typing (default: 50ms)
	// Set to 0 to output greeting instantly (no typing effect)
	TypingSpeed time.Duration
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

// findTermsvg returns the path to termsvg, checking common locations.
func findTermsvg() string {
	// First check PATH
	if path, err := exec.LookPath("termsvg"); err == nil {
		return path
	}

	// Check ~/go/bin (common Go install location)
	home, err := os.UserHomeDir()
	if err == nil {
		goBinPath := filepath.Join(home, "go", "bin", "termsvg")
		if _, err := os.Stat(goBinPath); err == nil {
			return goBinPath
		}
	}

	return ""
}

// CanRecord returns true if termsvg is available on the system.
func CanRecord() bool {
	return findTermsvg() != ""
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
	termsvgPath := findTermsvg()
	exportCmd := exec.Command(termsvgPath, "export", castPath, "-o", svgPath)
	if output, err := exportCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to export SVG: %w: %s", err, output)
	}

	// Patch SVG to not loop (play once and stop at end)
	if err := patchSVGNoLoop(svgPath); err != nil {
		return fmt.Errorf("failed to patch SVG: %w", err)
	}

	// Clean up intermediate cast file
	r.Cleanup()

	return fnErr
}

// generateCastFile creates an asciinema v2 cast file from captured output.
func (r *Recorder) generateCastFile(path string, output string, _ time.Duration) error {
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

	// Apply defaults
	typingSpeed := r.config.TypingSpeed
	if typingSpeed == 0 {
		typingSpeed = 50 * time.Millisecond
	}
	responseDelay := r.config.ResponseDelay
	if responseDelay == 0 {
		responseDelay = 500 * time.Millisecond
	}
	lineDelay := r.config.LineDelay
	if lineDelay == 0 {
		lineDelay = 300 * time.Millisecond
	}

	// 1. Agent greeting (instant, shown all at once)
	agentGreeting := r.config.AgentGreeting
	if agentGreeting == "" {
		agentGreeting = DefaultAgentGreeting
	}
	if strings.TrimSpace(agentGreeting) != "" {
		escapedGreeting := escapeJSON(agentGreeting)
		event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, escapedGreeting)
		buf.WriteString(event)
		buf.WriteString("\n")
		currentTime += 0.1 // Small pause after header
	}

	// 2. User prompt (typing simulation - character by character)
	userPrompt := r.config.UserPrompt
	if userPrompt == "" {
		userPrompt = fmt.Sprintf(DefaultUserPrompt, r.config.ScenarioName)
	}
	if strings.TrimSpace(userPrompt) != "" {
		// Type each character
		for _, char := range userPrompt {
			var output string
			if char == '\n' {
				// Newline needs carriage return to go back to left margin
				output = "\\r\\n"
				currentTime += 0.05 // Small pause at end of line
			} else if char == '\r' {
				// Skip carriage returns (we handle newlines above)
				continue
			} else {
				output = escapeJSON(string(char))
			}
			event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, output)
			buf.WriteString(event)
			buf.WriteString("\n")
			currentTime += typingSpeed.Seconds()
		}
		// Add final newline after user finishes typing
		event := fmt.Sprintf("[%.6f, \"o\", \"\\r\\n\"]", currentTime)
		buf.WriteString(event)
		buf.WriteString("\n")
		currentTime += 0.1
	}

	// 3. Agent response (instant)
	agentResponse := r.config.AgentResponse
	if agentResponse == "" {
		agentResponse = fmt.Sprintf(DefaultAgentResponse, r.config.ScenarioName)
	}
	if strings.TrimSpace(agentResponse) != "" {
		escapedResponse := escapeJSON(agentResponse)
		event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, escapedResponse)
		buf.WriteString(event)
		buf.WriteString("\n")
		currentTime += responseDelay.Seconds()
	}

	// 4. Scenario output (line by line with delay)
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

	for _, line := range nonEmptyLines {
		escapedLine := escapeJSON(line + "\r\n")
		event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, escapedLine)
		buf.WriteString(event)
		buf.WriteString("\n")
		currentTime += lineDelay.Seconds()
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

// patchSVGNoLoop modifies the SVG to play once and stop at the final frame.
func patchSVGNoLoop(svgPath string) error {
	content, err := os.ReadFile(svgPath)
	if err != nil {
		return err
	}

	svg := string(content)

	// Change from infinite loop to play once
	svg = strings.Replace(svg, "animation-iteration-count:infinite", "animation-iteration-count:1", 1)

	// Add fill-mode to keep final frame visible (insert after iteration-count)
	svg = strings.Replace(svg,
		"animation-iteration-count:1;",
		"animation-iteration-count:1;animation-fill-mode:forwards;",
		1)

	return os.WriteFile(svgPath, []byte(svg), 0644)
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

	// AgentGreeting is shown first, before the user types.
	// Set to " " to disable. If not set, DefaultAgentGreeting is used.
	AgentGreeting string

	// UserPrompt is what the user types (with typing simulation).
	// Use %s as placeholder for scenario name. Set to " " to disable.
	// If not set, DefaultUserPrompt is used.
	UserPrompt string

	// AgentResponse is shown after the user's prompt, before scenario output.
	// Use %s as placeholder for scenario name. Set to " " to disable.
	// If not set, DefaultAgentResponse is used.
	AgentResponse string

	// ResponseDelay is the pause after agent response before showing output (default: 500ms)
	ResponseDelay time.Duration

	// TermWidth is the terminal width in characters (default: 80)
	TermWidth int

	// TermHeight is the terminal height in characters (default: 24)
	TermHeight int

	// LineDelay is the minimum delay between output lines (default: 0.3s)
	LineDelay time.Duration

	// TypingSpeed is the delay between characters when simulating typing (default: 50ms)
	// Set to 0 to output greeting instantly (no typing effect)
	TypingSpeed time.Duration
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
		AgentGreeting: opts.AgentGreeting,
		UserPrompt:    opts.UserPrompt,
		AgentResponse: opts.AgentResponse,
		ResponseDelay: opts.ResponseDelay,
		TermWidth:     opts.TermWidth,
		TermHeight:    opts.TermHeight,
		LineDelay:     opts.LineDelay,
		TypingSpeed:   opts.TypingSpeed,
	}

	recorder := NewRecorder(config)
	return recorder.Record(fn)
}

// SessionRecordOptions configures recording of a Session conversation.
type SessionRecordOptions struct {
	// OutputDir is where recordings are saved (default: ./recordings)
	OutputDir string

	// TermWidth is the terminal width in characters (default: 80)
	TermWidth int

	// TermHeight is the terminal height in characters (default: 30)
	TermHeight int

	// TypingSpeed is delay between characters for user messages (default: 25ms)
	TypingSpeed time.Duration

	// LineDelay is delay between lines for agent messages (default: 100ms)
	LineDelay time.Duration

	// MessageDelay is pause between conversation turns (default: 500ms)
	MessageDelay time.Duration
}

// RecordSession records a conversation from session messages to SVG.
// Developer messages are shown with typing simulation (user input).
// Runner messages are shown line-by-line (agent output).
func RecordSession(session SessionMessages, opts SessionRecordOptions) error {
	if !CanRecord() {
		return ErrTermsvgNotFound
	}

	// Apply defaults
	if opts.OutputDir == "" {
		opts.OutputDir = "./recordings"
	}
	if opts.TermWidth == 0 {
		opts.TermWidth = 80
	}
	if opts.TermHeight == 0 {
		opts.TermHeight = 30
	}
	if opts.TypingSpeed == 0 {
		opts.TypingSpeed = 25 * time.Millisecond
	}
	if opts.LineDelay == 0 {
		opts.LineDelay = 100 * time.Millisecond
	}
	if opts.MessageDelay == 0 {
		opts.MessageDelay = 500 * time.Millisecond
	}

	// Ensure output dir exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	castPath := filepath.Join(opts.OutputDir, session.Name()+".cast")
	svgPath := filepath.Join(opts.OutputDir, session.Name()+".svg")

	// Generate cast file from session messages
	if err := generateSessionCast(castPath, session, opts); err != nil {
		return fmt.Errorf("generating cast file: %w", err)
	}

	// Export to SVG with black background
	termsvgPath := findTermsvg()
	exportCmd := exec.Command(termsvgPath, "export", castPath, "-o", svgPath, "-b", "#000000")
	if output, err := exportCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exporting SVG: %w: %s", err, output)
	}

	// Patch SVG to not loop
	if err := patchSVGNoLoop(svgPath); err != nil {
		return fmt.Errorf("patching SVG: %w", err)
	}

	// Clean up cast file
	_ = os.Remove(castPath)

	return nil
}

// SessionMessages interface for accessing session conversation data.
type SessionMessages interface {
	Name() string
	GetMessages() []SessionMessage
}

// SessionMessage represents a single message in a conversation.
type SessionMessage struct {
	Role    string // "developer" (user) or "runner" (agent)
	Content string
}

// generateSessionCast creates an asciinema cast file from session messages.
func generateSessionCast(path string, session SessionMessages, opts SessionRecordOptions) error {
	var buf bytes.Buffer

	// Write header
	header := fmt.Sprintf(`{"version": 2, "width": %d, "height": %d, "timestamp": %d, "title": "%s"}`,
		opts.TermWidth, opts.TermHeight, time.Now().Unix(), session.Name())
	buf.WriteString(header)
	buf.WriteString("\n")

	currentTime := 0.0

	// ANSI color codes
	greenOn := "\\u001b[32m"  // Green text
	colorOff := "\\u001b[0m"  // Reset color

	for _, msg := range session.GetMessages() {
		if msg.Role == "developer" {
			// User message - show green prompt and type in green
			prompt := greenOn + "> "
			event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, prompt)
			buf.WriteString(event)
			buf.WriteString("\n")
			currentTime += 0.05

			// Type each character in green
			for _, char := range msg.Content {
				var output string
				if char == '\n' {
					output = "\\r\\n" + greenOn + "> " // New line with green prompt
					currentTime += 0.05
				} else if char == '\r' {
					continue
				} else {
					output = escapeJSON(string(char))
				}
				event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, output)
				buf.WriteString(event)
				buf.WriteString("\n")
				currentTime += opts.TypingSpeed.Seconds()
			}

			// End of user input - reset color and add blank lines
			event = fmt.Sprintf("[%.6f, \"o\", \"%s\\r\\n\\r\\n\\r\\n\"]", currentTime, colorOff)
			buf.WriteString(event)
			buf.WriteString("\n")
			currentTime += opts.MessageDelay.Seconds()

		} else if msg.Role == "runner" {
			// Agent message - output line by line
			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				if line == "" {
					// Empty line
					event := fmt.Sprintf("[%.6f, \"o\", \"\\r\\n\"]", currentTime)
					buf.WriteString(event)
					buf.WriteString("\n")
				} else {
					escapedLine := escapeJSON(line + "\r\n")
					event := fmt.Sprintf("[%.6f, \"o\", \"%s\"]", currentTime, escapedLine)
					buf.WriteString(event)
					buf.WriteString("\n")
				}
				currentTime += opts.LineDelay.Seconds()
			}
			// Add blank lines after agent response
			event := fmt.Sprintf("[%.6f, \"o\", \"\\r\\n\\r\\n\"]", currentTime)
			buf.WriteString(event)
			buf.WriteString("\n")
			currentTime += opts.MessageDelay.Seconds()
		}
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
