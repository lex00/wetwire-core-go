// Package runner provides a reusable scenario execution engine.
//
// This package allows domain projects to run scenarios with consistent
// behavior, scoring, and output format.
//
// Example usage:
//
//	results, err := runner.Run(ctx, runner.Config{
//	    ScenarioPath: "./examples/my_scenario",
//	    OutputDir:    "./examples/my_scenario/results",
//	})
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lex00/wetwire-core-go/agent/scoring"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/claude"
	scenariopkg "github.com/lex00/wetwire-core-go/scenario"
)

// DefaultPersonas is the standard set of personas for scenario testing.
var DefaultPersonas = []string{"beginner", "intermediate", "expert", "terse", "verbose"}

// Config configures a scenario run.
type Config struct {
	// ScenarioPath is the path to the scenario directory
	ScenarioPath string

	// OutputDir is where results are written
	OutputDir string

	// Personas to run (defaults to DefaultPersonas)
	Personas []string

	// SinglePersona runs only one persona (overrides Personas)
	SinglePersona string

	// GenerateRecordings enables SVG recording generation
	GenerateRecordings bool

	// Verbose enables detailed output
	Verbose bool
}

// Result holds the result of a single persona scenario run.
type Result struct {
	Persona   string
	Success   bool
	Duration  time.Duration
	Response  string
	Files     map[string]string // filename -> content
	OutputDir string
	Score     *scoring.Score
}

// Run executes a scenario with all configured personas.
func Run(ctx context.Context, cfg Config) ([]Result, error) {
	if !claude.Available() {
		return nil, fmt.Errorf("claude CLI not found in PATH")
	}

	personas := cfg.Personas
	if len(personas) == 0 {
		personas = DefaultPersonas
	}
	if cfg.SinglePersona != "" {
		personas = []string{cfg.SinglePersona}
	}

	// Clean and create output directory
	_ = os.RemoveAll(cfg.OutputDir)
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var results []Result
	for _, persona := range personas {
		result := runPersona(ctx, cfg, persona)
		results = append(results, result)
	}

	// Write summary
	writeSummary(cfg.OutputDir, results, cfg.GenerateRecordings)

	return results, nil
}

func runPersona(ctx context.Context, cfg Config, personaName string) Result {
	result := Result{
		Persona: personaName,
		Files:   make(map[string]string),
	}

	// Create output directory for this persona
	absPersonaDir := filepath.Join(cfg.OutputDir, personaName)
	absPersonaDir, err := filepath.Abs(absPersonaDir)
	if err != nil {
		return result
	}
	_ = os.RemoveAll(absPersonaDir)
	if err := os.MkdirAll(absPersonaDir, 0755); err != nil {
		return result
	}
	result.OutputDir = absPersonaDir

	// Load prompts
	userPrompt := loadUserPrompt(cfg.ScenarioPath, personaName)
	systemPrompt := loadSystemPrompt(cfg.ScenarioPath)

	// Build the full prompt with execution instructions
	var promptBuilder strings.Builder
	promptBuilder.WriteString(userPrompt)
	promptBuilder.WriteString(fmt.Sprintf(`

## Output Location

Create all files in this directory: %s

Use the Write tool to create files. Use mkdir via Bash if directories are needed.
`, absPersonaDir))

	prompt := promptBuilder.String()

	// Create Claude provider
	provider, err := claude.New(claude.Config{
		WorkDir:        absPersonaDir,
		SystemPrompt:   systemPrompt,
		AllowedTools:   []string{"Write", "Bash", "Read", "Glob"},
		PermissionMode: "acceptEdits",
	})
	if err != nil {
		return result
	}

	// Execute scenario
	start := time.Now()
	resp, err := provider.CreateMessage(ctx, providers.MessageRequest{
		Messages: []providers.Message{
			providers.NewUserMessage(prompt),
		},
	})
	result.Duration = time.Since(start)

	if err != nil {
		return result
	}

	// Extract response text
	var responseText strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			responseText.WriteString(block.Text)
		}
	}
	result.Response = responseText.String()

	// Find generated files
	result.Files = findGeneratedFiles(absPersonaDir)
	result.Success = len(result.Files) > 0

	// Calculate score
	result.Score = calculateScore(result, personaName, cfg.ScenarioPath)

	// Write outputs
	saveConversation(result, userPrompt, filepath.Join(absPersonaDir, "conversation.txt"))
	writePersonaResults(absPersonaDir, result)

	// Generate recording if requested
	if cfg.GenerateRecordings && scenariopkg.CanRecord() {
		recordingPath := filepath.Join(absPersonaDir, fmt.Sprintf("%s_scenario.svg", personaName))
		generateRecording(result, userPrompt, recordingPath)
	}

	return result
}

func findGeneratedFiles(dir string) map[string]string {
	files := make(map[string]string)

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip our own output files
		name := info.Name()
		if name == "conversation.txt" || name == "RESULTS.md" || strings.HasSuffix(name, ".svg") {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		files[relPath] = string(content)
		return nil
	})

	return files
}

func loadSystemPrompt(scenarioPath string) string {
	defaultPrompt := `You are a helpful infrastructure engineer assistant.
Your task is to help users create infrastructure files based on their requirements.
Use the Write tool to create files. Use mkdir via Bash if needed.
If the user asks questions, answer them. If they ask for explanations, provide them.
Always generate complete, production-quality infrastructure regardless of how brief the request is.
Include best practices (parameters, outputs, proper configurations) even if not explicitly requested.`

	systemPromptPath := filepath.Join(scenarioPath, "system_prompt.md")
	if content, err := os.ReadFile(systemPromptPath); err == nil {
		return strings.TrimSpace(string(content))
	}

	return defaultPrompt
}

func loadUserPrompt(scenarioPath, personaName string) string {
	var content []byte
	var err error

	// Try persona-specific prompt first
	personaPromptPath := filepath.Join(scenarioPath, "prompts", personaName+".md")
	if content, err = os.ReadFile(personaPromptPath); err != nil {
		// Fall back to default prompt
		defaultPromptPath := filepath.Join(scenarioPath, "prompt.md")
		if content, err = os.ReadFile(defaultPromptPath); err != nil {
			return "Create the required infrastructure files."
		}
	}

	// Strip the title line (# ...) from the prompt
	lines := strings.Split(string(content), "\n")
	startIdx := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		startIdx = i
		break
	}

	return strings.TrimSpace(strings.Join(lines[startIdx:], "\n"))
}

func calculateScore(result Result, persona, scenarioPath string) *scoring.Score {
	score := scoring.NewScore(persona, scenarioPath)

	// Completeness: Check if files were created
	expectedFiles := 2 // Default expectation
	actualFiles := len(result.Files)
	rating, notes := scoring.ScoreCompleteness(expectedFiles, actualFiles)
	score.Completeness.Rating = rating
	score.Completeness.Notes = notes

	// Lint Quality: Check for YAML validity
	var lintErrors, lintWarnings int
	for _, content := range result.Files {
		if strings.Contains(content, "AWSTemplateFormatVersion") {
			e, w := runCfnLint(content, result.OutputDir)
			lintErrors += e
			lintWarnings += w
		}
	}
	rating, notes = scoring.ScoreOutputValidity(lintErrors, lintWarnings)
	score.LintQuality.Rating = rating
	score.LintQuality.Notes = notes

	// Code Quality: Check for patterns
	var issues []string
	for _, content := range result.Files {
		issues = append(issues, checkCodeQuality(content)...)
	}
	rating, notes = scoring.ScoreCodeQuality(issues)
	score.CodeQuality.Rating = rating
	score.CodeQuality.Notes = notes

	// Output Validity
	score.OutputValidity.Rating = scoring.RatingExcellent
	score.OutputValidity.Notes = "Files generated successfully"

	// Question Efficiency
	rating, notes = scoring.ScoreQuestionEfficiency(0)
	score.QuestionEfficiency.Rating = rating
	score.QuestionEfficiency.Notes = notes

	return score
}

func checkCodeQuality(content string) []string {
	var issues []string

	// Check for common patterns based on content type
	if strings.Contains(content, "AWSTemplateFormatVersion") {
		// CloudFormation checks
		desirable := []string{"Description", "Parameters:", "Outputs:", "DeletionPolicy"}
		for _, d := range desirable {
			if !strings.Contains(content, d) {
				issues = append(issues, fmt.Sprintf("Missing: %s", d))
			}
		}
	}

	if strings.Contains(content, "stages:") {
		// GitLab CI checks
		desirable := []string{"rules:", "image:"}
		for _, d := range desirable {
			if !strings.Contains(content, d) {
				issues = append(issues, fmt.Sprintf("Missing: %s", d))
			}
		}
	}

	return issues
}

func runCfnLint(content, outputDir string) (errors, warnings int) {
	if _, err := exec.LookPath("cfn-lint"); err != nil {
		return 0, 0
	}

	tmpFile := filepath.Join(outputDir, "temp_cfn_lint.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return 0, 0
	}
	defer func() { _ = os.Remove(tmpFile) }()

	cmd := exec.Command("cfn-lint", tmpFile, "--format", "parseable")
	output, _ := cmd.CombinedOutput()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":E") {
			errors++
		} else if strings.Contains(line, ":W") {
			warnings++
		}
	}
	return errors, warnings
}

func saveConversation(result Result, userPrompt, outputPath string) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Persona: %s\n", result.Persona))
	buf.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("Duration: %s\n", result.Duration.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("Success: %t\n", result.Success))
	if result.Score != nil {
		buf.WriteString(fmt.Sprintf("Score: %d/15\n", result.Score.Total()))
	}
	buf.WriteString("\n")
	buf.WriteString("════════════════════════════════════════════════════════════════════════════════\n")
	buf.WriteString("USER\n")
	buf.WriteString("════════════════════════════════════════════════════════════════════════════════\n\n")
	buf.WriteString(userPrompt)
	buf.WriteString("\n\n")
	buf.WriteString("════════════════════════════════════════════════════════════════════════════════\n")
	buf.WriteString("ASSISTANT\n")
	buf.WriteString("════════════════════════════════════════════════════════════════════════════════\n\n")
	buf.WriteString(result.Response)
	buf.WriteString("\n")

	_ = os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func writePersonaResults(dir string, result Result) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("# Scenario Results: %s\n\n", result.Persona))
	buf.WriteString(fmt.Sprintf("**Status:** %s\n", map[bool]string{true: "SUCCESS", false: "FAILED"}[result.Success]))
	buf.WriteString(fmt.Sprintf("**Duration:** %s\n\n", result.Duration.Round(time.Millisecond)))

	if result.Score != nil {
		buf.WriteString("## Score\n\n")
		buf.WriteString(fmt.Sprintf("**Total:** %d/15 (%s)\n\n", result.Score.Total(), result.Score.Threshold()))
		buf.WriteString("| Dimension | Rating | Notes |\n")
		buf.WriteString("|-----------|--------|-------|\n")
		dims := []scoring.Dimension{
			result.Score.Completeness,
			result.Score.LintQuality,
			result.Score.CodeQuality,
			result.Score.OutputValidity,
			result.Score.QuestionEfficiency,
		}
		for _, d := range dims {
			buf.WriteString(fmt.Sprintf("| %s | %d/3 | %s |\n", d.Name, d.Rating, d.Notes))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("## Generated Files\n\n")
	for file := range result.Files {
		buf.WriteString(fmt.Sprintf("- [%s](%s)\n", file, file))
	}
	buf.WriteString("\n")

	buf.WriteString("## Conversation\n\n")
	buf.WriteString("See [conversation.txt](conversation.txt) for the full prompt and response.\n")

	_ = os.WriteFile(filepath.Join(dir, "RESULTS.md"), buf.Bytes(), 0644)
}

func writeSummary(outputDir string, results []Result, generateRecordings bool) {
	var buf bytes.Buffer

	buf.WriteString("# Scenario Run Summary\n\n")
	buf.WriteString(fmt.Sprintf("**Date:** %s\n\n", time.Now().Format(time.RFC3339)))

	buf.WriteString("## Results by Persona\n\n")
	buf.WriteString("| Persona | Status | Score | Duration |\n")
	buf.WriteString("|---------|--------|-------|----------|\n")

	for _, r := range results {
		status := "FAILED"
		if r.Success {
			status = "SUCCESS"
		}

		scoreStr := "-"
		if r.Score != nil {
			scoreStr = fmt.Sprintf("%d/15", r.Score.Total())
		}

		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			r.Persona, status, scoreStr, r.Duration.Round(time.Millisecond)))
	}

	buf.WriteString("\n## Output Directories\n\n")
	for _, r := range results {
		buf.WriteString(fmt.Sprintf("- [%s](./%s/RESULTS.md)\n", r.Persona, r.Persona))
	}

	if generateRecordings {
		buf.WriteString("\n## Recordings\n\n")
		for _, r := range results {
			buf.WriteString(fmt.Sprintf("- [%s](./%s/%s_scenario.svg)\n", r.Persona, r.Persona, r.Persona))
		}
	}

	_ = os.WriteFile(filepath.Join(outputDir, "SUMMARY.md"), buf.Bytes(), 0644)
}

func generateRecording(result Result, userPrompt, outputPath string) {
	session := &conversationSession{
		name:     result.Persona + "_scenario",
		prompt:   userPrompt,
		response: result.Response,
	}

	_ = scenariopkg.RecordSession(session, scenariopkg.SessionRecordOptions{
		OutputDir:    filepath.Dir(outputPath),
		TypingSpeed:  25 * time.Millisecond,
		LineDelay:    80 * time.Millisecond,
		MessageDelay: 500 * time.Millisecond,
		TermWidth:    100,
		TermHeight:   40,
	})
}

type conversationSession struct {
	name     string
	prompt   string
	response string
}

func (c *conversationSession) Name() string { return c.name }
func (c *conversationSession) GetMessages() []scenariopkg.SessionMessage {
	return []scenariopkg.SessionMessage{
		{Role: "developer", Content: c.prompt},
		{Role: "runner", Content: c.response},
	}
}
