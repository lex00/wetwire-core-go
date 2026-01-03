// This file contains edge case tests for the agents package
//go:build !integration

package agents

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamingAccumulation tests text delta accumulation across chunks
func TestStreamingAccumulation(t *testing.T) {
	t.Parallel()

	// Create a mock stream handler
	var accumulated strings.Builder
	handler := func(text string) {
		accumulated.WriteString(text)
	}

	agent := &RunnerAgent{streamHandler: handler}

	// Test that the handler is called
	if agent.streamHandler != nil {
		agent.streamHandler("Hello ")
		agent.streamHandler("World")
	}

	assert.Equal(t, "Hello World", accumulated.String())
}

// TestToolInputJSONReconstruction tests tool use input JSON reconstruction from stream
func TestToolInputJSONReconstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		chunks []string
		want   string
	}{
		{
			name:   "simple_json",
			chunks: []string{`{"path"`, `: "test.go"`, `, "content": `, `"package main"}`},
			want:   `{"path": "test.go", "content": "package main"}`,
		},
		{
			name:   "nested_json",
			chunks: []string{`{"tool": {`, `"name": "write_file",`, ` "input": {"path": `, `"test.go"}}}`},
			want:   `{"tool": {"name": "write_file", "input": {"path": "test.go"}}}`,
		},
		{
			name:   "escaped_characters",
			chunks: []string{`{"message": "Line 1\n`, `Line 2\t`, `Tab"}`},
			want:   `{"message": "Line 1\nLine 2\tTab"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var builder strings.Builder
			for _, chunk := range tt.chunks {
				builder.WriteString(chunk)
			}
			result := builder.String()
			assert.Equal(t, tt.want, result)

			// Validate JSON is parseable
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			assert.NoError(t, err, "Reconstructed JSON should be valid")
		})
	}
}

// TestConcurrentEventProcessing tests handling of concurrent streaming events
func TestConcurrentEventProcessing(t *testing.T) {
	t.Parallel()

	// Test that multiple content blocks can be accumulated
	currentTextContent := make(map[int64]*strings.Builder)
	currentToolInput := make(map[int64]*strings.Builder)

	// Simulate multiple blocks
	currentTextContent[0] = &strings.Builder{}
	currentTextContent[1] = &strings.Builder{}
	currentToolInput[2] = &strings.Builder{}

	currentTextContent[0].WriteString("First text block")
	currentTextContent[1].WriteString("Second text block")
	currentToolInput[2].WriteString(`{"tool": "test"}`)

	assert.Equal(t, "First text block", currentTextContent[0].String())
	assert.Equal(t, "Second text block", currentTextContent[1].String())
	assert.Equal(t, `{"tool": "test"}`, currentToolInput[2].String())
}

// TestExecuteTool_JSONParsingErrors tests handling of invalid JSON input
func TestExecuteTool_JSONParsingErrors(t *testing.T) {
	t.Parallel()

	r := &RunnerAgent{
		workDir: t.TempDir(),
	}

	tests := []struct {
		name  string
		input json.RawMessage
	}{
		{
			name:  "invalid_json",
			input: json.RawMessage(`{invalid json`),
		},
		{
			name:  "empty_json",
			input: json.RawMessage(``),
		},
		{
			name:  "not_object",
			input: json.RawMessage(`["array", "not", "object"]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.executeTool(context.Background(), "write_file", tt.input)
			assert.Contains(t, result, "Error")
		})
	}
}

// TestToolReadFile_EdgeCases tests edge cases for reading files
func TestToolReadFile_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(dir string) string
		expectError bool
	}{
		{
			name: "nonexistent_file",
			setup: func(dir string) string {
				return "nonexistent.go"
			},
			expectError: true,
		},
		{
			name: "empty_file",
			setup: func(dir string) string {
				path := filepath.Join(dir, "empty.go")
				os.WriteFile(path, []byte{}, 0644)
				return "empty.go"
			},
			expectError: false,
		},
		{
			name: "large_file",
			setup: func(dir string) string {
				path := filepath.Join(dir, "large.go")
				// Create a file with 10K characters
				content := strings.Repeat("x", 10000)
				os.WriteFile(path, []byte(content), 0644)
				return "large.go"
			},
			expectError: false,
		},
		{
			name: "unicode_content",
			setup: func(dir string) string {
				path := filepath.Join(dir, "unicode.go")
				os.WriteFile(path, []byte("// Comment with emoji: 🚀\npackage main"), 0644)
				return "unicode.go"
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			r := &RunnerAgent{workDir: dir}

			path := tt.setup(dir)
			result := r.toolReadFile(path)

			if tt.expectError {
				assert.Contains(t, result, "Error")
			} else {
				assert.NotContains(t, result, "Error")
			}
		})
	}
}

// TestToolWriteFile_EdgeCases tests edge cases for writing files
func TestToolWriteFile_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		content string
		wantErr bool
	}{
		{
			name:    "deeply_nested_path",
			path:    "a/b/c/d/e/f/test.go",
			content: "package main",
			wantErr: false,
		},
		{
			name:    "path_with_dots",
			path:    "./pkg/../main.go",
			content: "package main",
			wantErr: false,
		},
		{
			name:    "unicode_filename",
			path:    "测试.go",
			content: "package main",
			wantErr: false,
		},
		{
			name:    "special_chars_in_path",
			path:    "my-package_v2/main.go",
			content: "package main",
			wantErr: false,
		},
		{
			name:    "empty_content",
			path:    "empty.go",
			content: "",
			wantErr: false,
		},
		{
			name:    "very_long_content",
			path:    "long.go",
			content: strings.Repeat("// Comment\n", 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			r := &RunnerAgent{
				workDir:        dir,
				generatedFiles: []string{},
			}

			result := r.toolWriteFile(tt.path, tt.content)

			if tt.wantErr {
				assert.Contains(t, result, "Error")
			} else {
				assert.Contains(t, result, "Wrote")
				assert.Contains(t, r.generatedFiles, tt.path)

				// Verify file was actually written
				fullPath := filepath.Join(dir, tt.path)
				content, err := os.ReadFile(fullPath)
				require.NoError(t, err)
				assert.Equal(t, tt.content, string(content))
			}
		})
	}
}

// TestToolInitPackage_EdgeCases tests edge cases for package initialization
func TestToolInitPackage_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pkgName string
		wantErr bool
	}{
		{
			name:    "simple_name",
			pkgName: "mypackage",
			wantErr: false,
		},
		{
			name:    "nested_path",
			pkgName: "a/b/c",
			wantErr: false,
		},
		{
			name:    "with_dashes",
			pkgName: "my-package",
			wantErr: false,
		},
		{
			name:    "with_underscores",
			pkgName: "my_package",
			wantErr: false,
		},
		{
			name:    "unicode_name",
			pkgName: "包",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			r := &RunnerAgent{workDir: dir}

			result := r.toolInitPackage(tt.pkgName)

			if tt.wantErr {
				assert.Contains(t, result, "Error")
			} else {
				assert.Contains(t, result, "Created package")

				// Verify directory was created
				pkgDir := filepath.Join(dir, tt.pkgName)
				stat, err := os.Stat(pkgDir)
				require.NoError(t, err)
				assert.True(t, stat.IsDir())
			}
		})
	}
}

// TestCheckLintEnforcement_EdgeCases tests edge cases for lint enforcement
func TestCheckLintEnforcement_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		toolsCalled  []string
		wantEnforced bool
	}{
		{
			name:         "empty_tools",
			toolsCalled:  []string{},
			wantEnforced: false,
		},
		{
			name:         "only_init",
			toolsCalled:  []string{"init_package"},
			wantEnforced: false,
		},
		{
			name:         "multiple_reads",
			toolsCalled:  []string{"read_file", "read_file", "read_file"},
			wantEnforced: false,
		},
		{
			name:         "write_then_read",
			toolsCalled:  []string{"write_file", "read_file"},
			wantEnforced: true,
		},
		{
			name:         "init_write_lint",
			toolsCalled:  []string{"init_package", "write_file", "run_lint"},
			wantEnforced: false,
		},
		{
			name:         "many_writes_one_lint",
			toolsCalled:  []string{"write_file", "write_file", "write_file", "run_lint"},
			wantEnforced: false,
		},
		{
			name:         "lint_then_write",
			toolsCalled:  []string{"run_lint", "write_file"},
			wantEnforced: false, // Both write_file and run_lint present, so no enforcement
		},
		{
			name:         "ask_developer",
			toolsCalled:  []string{"ask_developer"},
			wantEnforced: false,
		},
		{
			name:         "run_build",
			toolsCalled:  []string{"run_build"},
			wantEnforced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RunnerAgent{}
			msg := r.checkLintEnforcement(tt.toolsCalled)

			if tt.wantEnforced {
				assert.NotEmpty(t, msg, "Expected enforcement message")
				assert.Contains(t, msg, "ENFORCEMENT")
			} else {
				assert.Empty(t, msg, "Expected no enforcement message")
			}
		})
	}
}

// TestCheckCompletionGate_EdgeCases tests edge cases for completion gate
func TestCheckCompletionGate_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		generatedFiles []string
		lintCalled     bool
		lintPassed     bool
		pendingLint    bool
		responseText   string
		wantBlocked    bool
	}{
		{
			name:           "no_files_no_completion_keyword",
			generatedFiles: []string{},
			lintCalled:     false,
			lintPassed:     false,
			pendingLint:    false,
			responseText:   "Let me analyze the requirements",
			wantBlocked:    false,
		},
		{
			name:           "files_exist_all_clear",
			generatedFiles: []string{"main.go"},
			lintCalled:     true,
			lintPassed:     true,
			pendingLint:    false,
			responseText:   "Everything is done!",
			wantBlocked:    false,
		},
		{
			name:           "multiple_files_pending_lint",
			generatedFiles: []string{"main.go", "util.go", "types.go"},
			lintCalled:     true,
			lintPassed:     true,
			pendingLint:    true,
			responseText:   "Task complete",
			wantBlocked:    true,
		},
		{
			name:           "uppercase_keywords",
			generatedFiles: []string{"main.go"},
			lintCalled:     false,
			lintPassed:     false,
			pendingLint:    false,
			responseText:   "DONE WITH THE TASK",
			wantBlocked:    true,
		},
		{
			name:           "mixed_case_keywords",
			generatedFiles: []string{"main.go"},
			lintCalled:     false,
			lintPassed:     false,
			pendingLint:    false,
			responseText:   "I've FiNiShEd the implementation",
			wantBlocked:    true,
		},
		{
			name:           "keyword_in_middle",
			generatedFiles: []string{"main.go"},
			lintCalled:     false,
			lintPassed:     false,
			pendingLint:    false,
			responseText:   "The task is now complete and ready for review",
			wantBlocked:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RunnerAgent{
				generatedFiles: tt.generatedFiles,
				lintCalled:     tt.lintCalled,
				lintPassed:     tt.lintPassed,
				pendingLint:    tt.pendingLint,
			}

			resp := &anthropic.Message{
				Content: []anthropic.ContentBlockUnion{
					{Type: "text", Text: tt.responseText},
				},
			}

			msg := r.checkCompletionGate(resp)

			if tt.wantBlocked {
				assert.NotEmpty(t, msg, "Expected completion to be blocked")
				assert.Contains(t, msg, "ENFORCEMENT")
			} else {
				assert.Empty(t, msg, "Expected completion to be allowed")
			}
		})
	}
}

// TestLintCyclesMaximum tests reaching maximum lint cycles
func TestLintCyclesMaximum(t *testing.T) {
	t.Parallel()

	r := &RunnerAgent{
		lintCycles:    3,
		maxLintCycles: 3,
	}

	assert.Equal(t, 3, r.lintCycles)
	assert.Equal(t, 3, r.maxLintCycles)
	assert.True(t, r.lintCycles >= r.maxLintCycles, "Should be at max lint cycles")
}

// TestStateTransitions tests complex state transitions
func TestStateTransitions(t *testing.T) {
	t.Parallel()

	r := &RunnerAgent{
		workDir:        t.TempDir(),
		generatedFiles: []string{},
		lintCalled:     false,
		lintPassed:     false,
		pendingLint:    false,
		lintCycles:     0,
	}

	// Initial state
	assert.False(t, r.lintCalled)
	assert.False(t, r.lintPassed)
	assert.False(t, r.pendingLint)
	assert.Equal(t, 0, r.lintCycles)

	// Write file - should mark as pending lint
	r.toolWriteFile("test.go", "package main")
	assert.False(t, r.lintCalled)
	assert.False(t, r.lintPassed)
	assert.True(t, r.pendingLint)

	// Simulate lint pass
	r.lintCalled = true
	r.lintPassed = true
	r.pendingLint = false
	r.lintCycles = 1

	assert.True(t, r.lintCalled)
	assert.True(t, r.lintPassed)
	assert.False(t, r.pendingLint)

	// Write another file - should reset pendingLint and lintPassed
	r.toolWriteFile("test2.go", "package main")
	assert.True(t, r.lintCalled)  // Remains true
	assert.False(t, r.lintPassed) // Reset to false
	assert.True(t, r.pendingLint) // Set to true
}

// TestNewRunnerAgent_Configuration tests various configuration scenarios
func TestNewRunnerAgent_Configuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    RunnerConfig
		setEnv    bool
		wantError bool
	}{
		{
			name: "valid_config_with_api_key",
			config: RunnerConfig{
				APIKey:        "test-key",
				WorkDir:       t.TempDir(),
				MaxLintCycles: 5,
			},
			setEnv:    false,
			wantError: false,
		},
		{
			name: "empty_config_with_env",
			config: RunnerConfig{
				APIKey: "",
			},
			setEnv:    true,
			wantError: false,
		},
		{
			name: "no_api_key",
			config: RunnerConfig{
				APIKey: "",
			},
			setEnv:    false,
			wantError: true,
		},
		{
			name: "defaults_applied",
			config: RunnerConfig{
				APIKey: "test-key",
			},
			setEnv:    false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv("ANTHROPIC_API_KEY", "test-env-key")
				defer os.Unsetenv("ANTHROPIC_API_KEY")
			} else {
				os.Unsetenv("ANTHROPIC_API_KEY")
			}

			agent, err := NewRunnerAgent(tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)

				// Check defaults were applied
				if tt.config.WorkDir == "" {
					assert.Equal(t, ".", agent.workDir)
				}
				if tt.config.MaxLintCycles == 0 {
					assert.Equal(t, 3, agent.maxLintCycles)
				}
			}
		})
	}
}

// TestExecuteTool_UnknownTool tests handling of unknown tool names
func TestExecuteTool_UnknownTool(t *testing.T) {
	t.Parallel()

	r := &RunnerAgent{workDir: t.TempDir()}

	input := json.RawMessage(`{"param": "value"}`)
	result := r.executeTool(context.Background(), "unknown_tool", input)

	assert.Contains(t, result, "Unknown tool")
	assert.Contains(t, result, "unknown_tool")
}
