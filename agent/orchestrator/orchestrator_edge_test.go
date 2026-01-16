// This file contains edge case tests for the orchestrator package
package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lex00/wetwire-core-go/agent/personas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRunner implements the Runner interface for testing
type mockRunner struct {
	runFunc          func(ctx context.Context, prompt string) error
	askDeveloperFunc func(ctx context.Context, question string) (string, error)
	generatedFiles   []string
	template         string
}

func (m *mockRunner) Run(ctx context.Context, prompt string) error {
	if m.runFunc != nil {
		return m.runFunc(ctx, prompt)
	}
	return nil
}

func (m *mockRunner) AskDeveloper(ctx context.Context, question string) (string, error) {
	if m.askDeveloperFunc != nil {
		return m.askDeveloperFunc(ctx, question)
	}
	return "default answer", nil
}

func (m *mockRunner) GetGeneratedFiles() []string {
	return m.generatedFiles
}

func (m *mockRunner) GetTemplate() string {
	return m.template
}

// mockDeveloper implements the Developer interface for testing
type mockDeveloper struct {
	respondFunc func(ctx context.Context, message string) (string, error)
	responses   []string
	callCount   int
}

func (m *mockDeveloper) Respond(ctx context.Context, message string) (string, error) {
	if m.respondFunc != nil {
		return m.respondFunc(ctx, message)
	}
	if m.callCount < len(m.responses) {
		response := m.responses[m.callCount]
		m.callCount++
		return response, nil
	}
	return "no more responses", nil
}

// TestOrchestrator_Run_EdgeCases tests edge cases for orchestrator execution
func TestOrchestrator_Run_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    Config
		runner    *mockRunner
		developer *mockDeveloper
		wantErr   bool
	}{
		{
			name: "successful_run",
			config: Config{
				Persona:       personas.Beginner,
				Scenario:      "test",
				InitialPrompt: "Create a bucket",
			},
			runner: &mockRunner{
				runFunc: func(ctx context.Context, prompt string) error {
					return nil
				},
				generatedFiles: []string{"main.go"},
				template:       `{"Resources": {}}`,
			},
			developer: &mockDeveloper{},
			wantErr:   false,
		},
		{
			name: "runner_fails",
			config: Config{
				Persona:       personas.Expert,
				Scenario:      "test",
				InitialPrompt: "Complex task",
			},
			runner: &mockRunner{
				runFunc: func(ctx context.Context, prompt string) error {
					return errors.New("runner error")
				},
			},
			developer: &mockDeveloper{},
			wantErr:   true,
		},
		{
			name: "empty_prompt",
			config: Config{
				Persona:       personas.Terse,
				Scenario:      "test",
				InitialPrompt: "",
			},
			runner:    &mockRunner{},
			developer: &mockDeveloper{},
			wantErr:   false,
		},
		{
			name: "very_long_prompt",
			config: Config{
				Persona:       personas.Verbose,
				Scenario:      "test",
				InitialPrompt: strings.Repeat("Create resources. ", 1000),
			},
			runner:    &mockRunner{},
			developer: &mockDeveloper{},
			wantErr:   false,
		},
		{
			name: "unicode_prompt",
			config: Config{
				Persona:       personas.Beginner,
				Scenario:      "test",
				InitialPrompt: "创建一个 S3 bucket 🚀",
			},
			runner:    &mockRunner{},
			developer: &mockDeveloper{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orch := New(tt.config, tt.developer, tt.runner)
			session, err := orch.Run(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.Equal(t, tt.config.Persona.Name, session.Persona)
				assert.Equal(t, tt.config.Scenario, session.Scenario)
				assert.Equal(t, tt.config.InitialPrompt, session.InitialPrompt)
			}
		})
	}
}

// TestOrchestrator_Run_Timeout tests timeout enforcement
func TestOrchestrator_Run_Timeout(t *testing.T) {
	t.Parallel()

	config := Config{
		Persona:       personas.Beginner,
		Scenario:      "test",
		InitialPrompt: "Test",
		Timeout:       100 * time.Millisecond,
	}

	runner := &mockRunner{
		runFunc: func(ctx context.Context, prompt string) error {
			// Simulate long-running operation
			select {
			case <-time.After(1 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	developer := &mockDeveloper{}
	orch := New(config, developer, runner)

	_, err := orch.Run(context.Background())

	// Should get a timeout error
	assert.Error(t, err)
	assert.True(t,
		err == context.DeadlineExceeded ||
			err.Error() == "context deadline exceeded" ||
			strings.Contains(err.Error(), "deadline exceeded"),
	)
}

// TestOrchestrator_Run_ContextCancellation tests context cancellation
func TestOrchestrator_Run_ContextCancellation(t *testing.T) {
	t.Parallel()

	config := Config{
		Persona:       personas.Expert,
		Scenario:      "test",
		InitialPrompt: "Test",
	}

	runner := &mockRunner{
		runFunc: func(ctx context.Context, prompt string) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	developer := &mockDeveloper{}
	orch := New(config, developer, runner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := orch.Run(ctx)

	assert.Error(t, err)
	assert.True(t,
		err == context.Canceled ||
			err.Error() == "context canceled" ||
			strings.Contains(err.Error(), "canceled"),
	)
}

// TestOrchestrator_CalculateScore_EdgeCases tests edge cases for score calculation
func TestOrchestrator_CalculateScore_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		expectedResources  int
		actualResources    int
		lintPassed         bool
		validationErrors   int
		validationWarnings int
		minTotal           int
		maxTotal           int
	}{
		{
			name:               "perfect_score",
			expectedResources:  10,
			actualResources:    10,
			lintPassed:         true,
			validationErrors:   0,
			validationWarnings: 0,
			minTotal:           10, // Should be excellent (0-12 scale)
			maxTotal:           12,
		},
		{
			name:               "zero_resources",
			expectedResources:  0,
			actualResources:    0,
			lintPassed:         true,
			validationErrors:   0,
			validationWarnings: 0,
			minTotal:           8,
			maxTotal:           12,
		},
		{
			name:               "complete_failure",
			expectedResources:  10,
			actualResources:    0,
			lintPassed:         false,
			validationErrors:   5,
			validationWarnings: 10,
			minTotal:           0,
			maxTotal:           4,
		},
		{
			name:               "partial_success",
			expectedResources:  10,
			actualResources:    5,
			lintPassed:         true,
			validationErrors:   0,
			validationWarnings: 3,
			minTotal:           5,
			maxTotal:           10,
		},
		{
			name:               "very_large_numbers",
			expectedResources:  1000,
			actualResources:    1000,
			lintPassed:         true,
			validationErrors:   0,
			validationWarnings: 0,
			minTotal:           10,
			maxTotal:           12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Persona = personas.Beginner
			config.Scenario = tt.name

			runner := &mockRunner{}
			developer := &mockDeveloper{}
			orch := New(config, developer, runner)

			// Add lint cycles to session
			if tt.lintPassed {
				orch.session.AddLintCycle([]string{}, 0, true)
			} else {
				orch.session.AddLintCycle([]string{"error"}, 0, false)
			}

			score := orch.CalculateScore(
				tt.expectedResources,
				tt.actualResources,
				tt.lintPassed,
				tt.validationErrors,
				tt.validationWarnings,
			)

			assert.NotNil(t, score)
			total := score.Total()
			assert.GreaterOrEqual(t, total, tt.minTotal, "Score too low")
			assert.LessOrEqual(t, total, tt.maxTotal, "Score too high")
		})
	}
}

// TestOrchestrator_Session_EdgeCases tests session tracking edge cases
func TestOrchestrator_Session_EdgeCases(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.Persona = personas.Intermediate
	config.Scenario = "test"

	runner := &mockRunner{}
	developer := &mockDeveloper{}
	orch := New(config, developer, runner)

	session := orch.Session()
	assert.NotNil(t, session)
	assert.Equal(t, "intermediate", session.Persona)
	assert.Equal(t, "test", session.Scenario)
}

// TestHumanDeveloper_Respond tests human developer interaction
func TestHumanDeveloper_Respond(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		reader   func() (string, error)
		message  string
		expected string
		wantErr  bool
	}{
		{
			name: "successful_read",
			reader: func() (string, error) {
				return "user answer", nil
			},
			message:  "What do you want?",
			expected: "user answer",
			wantErr:  false,
		},
		{
			name: "read_error",
			reader: func() (string, error) {
				return "", errors.New("read error")
			},
			message:  "Question?",
			expected: "",
			wantErr:  true,
		},
		{
			name: "empty_answer",
			reader: func() (string, error) {
				return "", nil
			},
			message:  "Question?",
			expected: "",
			wantErr:  false,
		},
		{
			name: "very_long_answer",
			reader: func() (string, error) {
				return strings.Repeat("answer ", 1000), nil
			},
			message:  "Question?",
			expected: strings.Repeat("answer ", 1000),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			human := NewHumanDeveloper(tt.reader)
			response, err := human.Respond(context.Background(), tt.message)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, response)
			}
		})
	}
}

// TestAIDeveloper_Respond tests AI developer with personas
func TestAIDeveloper_Respond(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		persona   personas.Persona
		message   string
		responder func(ctx context.Context, systemPrompt, message string) (string, error)
		expected  string
		wantErr   bool
	}{
		{
			name:    "beginner_persona",
			persona: personas.Beginner,
			message: "What encryption?",
			responder: func(ctx context.Context, systemPrompt, message string) (string, error) {
				assert.Contains(t, systemPrompt, "new to infrastructure")
				return "I'm not sure, what do you recommend?", nil
			},
			expected: "I'm not sure, what do you recommend?",
			wantErr:  false,
		},
		{
			name:    "expert_persona",
			persona: personas.Expert,
			message: "What encryption?",
			responder: func(ctx context.Context, systemPrompt, message string) (string, error) {
				assert.Contains(t, systemPrompt, "senior infrastructure engineer")
				return "Use AES-256 with KMS key", nil
			},
			expected: "Use AES-256 with KMS key",
			wantErr:  false,
		},
		{
			name:    "responder_error",
			persona: personas.Terse,
			message: "Question?",
			responder: func(ctx context.Context, systemPrompt, message string) (string, error) {
				return "", errors.New("API error")
			},
			expected: "",
			wantErr:  true,
		},
		{
			name:    "empty_response",
			persona: personas.Verbose,
			message: "Question?",
			responder: func(ctx context.Context, systemPrompt, message string) (string, error) {
				return "", nil
			},
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := NewAIDeveloper(tt.persona, tt.responder)
			response, err := ai.Respond(context.Background(), tt.message)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, response)
			}
		})
	}
}

// TestDefaultConfig_EdgeCases tests default configuration edge cases
func TestDefaultConfig_EdgeCases(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.Equal(t, 3, config.MaxLintCycles)
	assert.Equal(t, 10*time.Minute, config.Timeout)
	assert.Empty(t, config.Scenario)
	assert.Empty(t, config.InitialPrompt)
	assert.Empty(t, config.OutputDir)
}

// TestOrchestrator_Run_CollectsResults tests that results are collected properly
func TestOrchestrator_Run_CollectsResults(t *testing.T) {
	t.Parallel()

	config := Config{
		Persona:       personas.Expert,
		Scenario:      "s3_bucket",
		InitialPrompt: "Create an S3 bucket",
	}

	expectedFiles := []string{"storage.go", "main.go"}
	expectedTemplate := `{"Resources": {"Bucket": {}}}`

	runner := &mockRunner{
		generatedFiles: expectedFiles,
		template:       expectedTemplate,
	}

	developer := &mockDeveloper{}
	orch := New(config, developer, runner)

	session, err := orch.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedFiles, session.GeneratedFiles)
	assert.Equal(t, expectedTemplate, session.TemplateJSON)
	assert.False(t, session.EndTime.IsZero())
	assert.True(t, session.EndTime.After(session.StartTime) || session.EndTime.Equal(session.StartTime))
}

// TestOrchestrator_Run_NoTimeout tests running without timeout
func TestOrchestrator_Run_NoTimeout(t *testing.T) {
	t.Parallel()

	config := Config{
		Persona:       personas.Beginner,
		Scenario:      "test",
		InitialPrompt: "Test",
		Timeout:       0, // No timeout
	}

	callCount := 0
	runner := &mockRunner{
		runFunc: func(ctx context.Context, prompt string) error {
			callCount++
			// Should not have a timeout
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}

	developer := &mockDeveloper{}
	orch := New(config, developer, runner)

	_, err := orch.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

// TestOrchestrator_MultipleConcurrentRuns tests concurrent orchestrator runs
func TestOrchestrator_MultipleConcurrentRuns(t *testing.T) {
	t.Parallel()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			config := Config{
				Persona:       personas.All()[idx%5],
				Scenario:      "concurrent_test",
				InitialPrompt: "Test",
			}

			runner := &mockRunner{}
			developer := &mockDeveloper{}
			orch := New(config, developer, runner)

			_, err := orch.Run(context.Background())
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all runs to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
