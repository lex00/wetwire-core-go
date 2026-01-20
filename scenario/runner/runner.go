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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lex00/wetwire-core-go/agent/scoring"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/claude"
	scenariopkg "github.com/lex00/wetwire-core-go/scenario"
	"github.com/lex00/wetwire-core-go/scenario/validator"
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

	// Validate enables validation against scenario rules and expected files
	Validate bool
}

// Result holds the result of a single persona scenario run.
type Result struct {
	Persona          string
	Success          bool
	Duration         time.Duration
	Response         string
	Files            map[string]string // filename -> content
	OutputDir        string
	Score            *scoring.Score
	ValidationReport *validator.ValidationReport
}

// Run executes a scenario with all configured personas.
func Run(ctx context.Context, cfg Config) ([]Result, error) {
	if !claude.Available() {
		return nil, fmt.Errorf("claude CLI not found in PATH")
	}

	// Load scenario config to get model setting
	scenarioConfig, err := scenariopkg.Load(cfg.ScenarioPath)
	if err != nil {
		// Not fatal - use defaults if no scenario.yaml
		scenarioConfig = &scenariopkg.ScenarioConfig{}
	}

	personas := cfg.Personas
	if len(personas) == 0 {
		// Check if scenario config specifies personas
		if scenarioConfig.Prompts != nil && len(scenarioConfig.Prompts.Personas) > 0 {
			personas = scenarioConfig.Prompts.Personas
		} else {
			personas = DefaultPersonas
		}
	}
	if cfg.SinglePersona != "" {
		personas = []string{cfg.SinglePersona}
	}

	// Clean and create output directory
	_ = os.RemoveAll(cfg.OutputDir)
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	model := scenarioConfig.Model
	if model != "" {
		fmt.Printf("Model: %s\n", model)
	}

	var results []Result

	if len(personas) == 1 {
		// Single persona: run directly with streaming
		result := runPersona(ctx, cfg, personas[0], model, scenarioConfig, cfg.Verbose)
		results = []Result{result}
	} else {
		// Multiple personas: run in parallel without streaming (would be interleaved)
		var wg sync.WaitGroup
		var mu sync.Mutex
		results = make([]Result, len(personas))

		fmt.Printf("Running %d personas in parallel...\n", len(personas))

		for i, persona := range personas {
			wg.Add(1)
			go func(idx int, p string) {
				defer wg.Done()
				mu.Lock()
				fmt.Printf("  [%s] Starting...\n", p)
				mu.Unlock()

				result := runPersona(ctx, cfg, p, model, scenarioConfig, false) // no streaming for parallel

				mu.Lock()
				results[idx] = result
				status := "FAILED"
				if result.Success {
					status = "SUCCESS"
				}
				fmt.Printf("  [%s] Done: %s (%s)\n", p, status, result.Duration.Round(time.Millisecond))
				mu.Unlock()
			}(i, persona)
		}

		wg.Wait()
		fmt.Println()
	}

	// Write summary after all personas complete
	writeSummary(cfg.OutputDir, results, cfg.GenerateRecordings)

	return results, nil
}

func runPersona(ctx context.Context, cfg Config, personaName, model string, scenarioConfig *scenariopkg.ScenarioConfig, verbose bool) Result {
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
		Model:          model,
		AllowedTools:   []string{"Write", "Bash", "Read", "Glob"},
		PermissionMode: "acceptEdits",
	})
	if err != nil {
		return result
	}

	// Execute scenario with streaming for progress visibility
	start := time.Now()
	var responseText strings.Builder

	// Stream handler to show progress and capture response
	streamHandler := func(text string) {
		if verbose {
			fmt.Print(text)
		}
		responseText.WriteString(text)
	}

	resp, err := provider.StreamMessage(ctx, providers.MessageRequest{
		Messages: []providers.Message{
			providers.NewUserMessage(prompt),
		},
	}, streamHandler)
	result.Duration = time.Since(start)

	if verbose {
		fmt.Println() // newline after streaming
	}

	if err != nil {
		return result
	}

	// Use streamed text, or extract from response if empty
	if responseText.Len() > 0 {
		result.Response = responseText.String()
	} else {
		for _, block := range resp.Content {
			if block.Type == "text" {
				responseText.WriteString(block.Text)
			}
		}
		result.Response = responseText.String()
	}

	// Find generated files
	result.Files = findGeneratedFiles(absPersonaDir)
	result.Success = len(result.Files) > 0

	// Calculate score
	result.Score = calculateScore(result, personaName, cfg.ScenarioPath)

	// Run validation if enabled
	if cfg.Validate && scenarioConfig != nil {
		absScenarioPath, _ := filepath.Abs(cfg.ScenarioPath)
		v := validator.New(scenarioConfig, absScenarioPath, absPersonaDir)
		report, err := v.Validate()
		if err == nil {
			result.ValidationReport = report
			// Update score based on validation
			if report.Score > 0 {
				result.Score = updateScoreFromValidation(result.Score, report)
			}
		}
	}

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

	// Lint Quality: Deferred to domain tools
	if len(result.Files) > 0 {
		score.LintQuality.Rating = scoring.RatingExcellent
		score.LintQuality.Notes = "Deferred to domain tools"
	} else {
		score.LintQuality.Rating = scoring.RatingNone
		score.LintQuality.Notes = "No files to lint"
	}

	// Output Validity: Files were generated
	if len(result.Files) > 0 {
		score.OutputValidity.Rating = scoring.RatingExcellent
		score.OutputValidity.Notes = fmt.Sprintf("%d files generated", len(result.Files))
	} else {
		score.OutputValidity.Rating = scoring.RatingNone
		score.OutputValidity.Notes = "No files generated"
	}

	// Question Efficiency
	rating, notes = scoring.ScoreQuestionEfficiency(0)
	score.QuestionEfficiency.Rating = rating
	score.QuestionEfficiency.Notes = notes

	return score
}

// updateScoreFromValidation updates the score based on validation results.
func updateScoreFromValidation(score *scoring.Score, report *validator.ValidationReport) *scoring.Score {
	// Completeness: Based on resource count validation
	allCountsPassed := true
	for _, result := range report.ResourceCounts {
		if !result.Passed {
			allCountsPassed = false
			break
		}
	}
	if allCountsPassed && len(report.ResourceCounts) > 0 {
		score.Completeness.Rating = scoring.RatingExcellent
		score.Completeness.Notes = "All resource counts met"
	} else if !allCountsPassed {
		score.Completeness.Rating = scoring.RatingNone
		score.Completeness.Notes = "Resource count validation failed"
	}

	// Output Validity: Based on cross-ref validation
	allRefsPassed := true
	for _, result := range report.CrossDomainRefs {
		if !result.Passed {
			allRefsPassed = false
			break
		}
	}
	if allRefsPassed && len(report.CrossDomainRefs) > 0 {
		score.OutputValidity.Rating = scoring.RatingExcellent
		score.OutputValidity.Notes = "All cross-domain refs found"
	} else if !allRefsPassed {
		score.OutputValidity.Rating = scoring.RatingNone
		score.OutputValidity.Notes = "Cross-domain ref validation failed"
	}

	// Question Efficiency: Based on expected file comparison
	missingCount := 0
	for _, result := range report.FileComparisons {
		if result.Missing {
			missingCount++
		}
	}
	if missingCount == 0 && len(report.FileComparisons) > 0 {
		score.QuestionEfficiency.Rating = scoring.RatingExcellent
		score.QuestionEfficiency.Notes = "All expected files found"
	} else if missingCount > 0 {
		if missingCount >= 3 {
			score.QuestionEfficiency.Rating = scoring.RatingNone
		} else {
			score.QuestionEfficiency.Rating = scoring.Rating(int(scoring.RatingExcellent) - missingCount)
		}
		score.QuestionEfficiency.Notes = fmt.Sprintf("%d expected files missing", missingCount)
	}

	return score
}

func saveConversation(result Result, userPrompt, outputPath string) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Persona: %s\n", result.Persona))
	buf.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("Duration: %s\n", result.Duration.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("Success: %t\n", result.Success))
	if result.Score != nil {
		buf.WriteString(fmt.Sprintf("Score: %d/12\n", result.Score.Total()))
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
		buf.WriteString(fmt.Sprintf("**Total:** %d/12 (%s)\n\n", result.Score.Total(), result.Score.Threshold()))
		buf.WriteString("| Dimension | Rating | Notes |\n")
		buf.WriteString("|-----------|--------|-------|\n")
		dims := []scoring.Dimension{
			result.Score.Completeness,
			result.Score.LintQuality,
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

	// Include validation report if available
	if result.ValidationReport != nil {
		buf.WriteString("## Validation\n\n")
		status := "✅ PASSED"
		if !result.ValidationReport.Passed {
			status = "❌ FAILED"
		}
		buf.WriteString(fmt.Sprintf("**Status:** %s\n\n", status))

		// Resource counts
		if len(result.ValidationReport.ResourceCounts) > 0 {
			buf.WriteString("### Resource Counts\n\n")
			buf.WriteString("| Domain | Type | Found | Constraint | Status |\n")
			buf.WriteString("|--------|------|-------|------------|--------|\n")
			for domain, rc := range result.ValidationReport.ResourceCounts {
				constraint := fmt.Sprintf("min: %d", rc.Min)
				if rc.Max > 0 {
					constraint += fmt.Sprintf(", max: %d", rc.Max)
				}
				status := "✅"
				if !rc.Passed {
					status = "❌"
				}
				buf.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n",
					domain, rc.ResourceType, rc.Found, constraint, status))
			}
			buf.WriteString("\n")
		}

		// Cross-domain refs
		if len(result.ValidationReport.CrossDomainRefs) > 0 {
			buf.WriteString("### Cross-Domain References\n\n")
			for _, ref := range result.ValidationReport.CrossDomainRefs {
				buf.WriteString(fmt.Sprintf("**%s → %s:**\n", ref.From, ref.To))
				for _, found := range ref.FoundRefs {
					buf.WriteString(fmt.Sprintf("- ✅ `%s`\n", found))
				}
				for _, missing := range ref.MissingRefs {
					buf.WriteString(fmt.Sprintf("- ❌ `%s` (missing)\n", missing))
				}
				buf.WriteString("\n")
			}
		}
	}

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
			scoreStr = fmt.Sprintf("%d/12", r.Score.Total())
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
