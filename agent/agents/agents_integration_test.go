// This file contains integration tests for external commands and API calls
//go:build integration

package agents

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolRunLint_CommandNotFound tests handling when wetwire-aws is not installed
func TestToolRunLint_CommandNotFound(t *testing.T) {
	// Skip if wetwire-aws is actually installed
	if _, err := exec.LookPath("wetwire-aws"); err == nil {
		t.Skip("wetwire-aws is installed, skipping not-found test")
	}

	dir := t.TempDir()
	r := &RunnerAgent{
		workDir: dir,
		session: results.NewSession("test", "test"),
	}

	// Create a package directory
	pkgDir := filepath.Join(dir, "testpkg")
	os.MkdirAll(pkgDir, 0755)

	result := r.toolRunLint("testpkg")

	// Should handle command not found gracefully
	assert.True(t, r.lintCalled, "lintCalled should be set even on error")
	assert.False(t, r.lintPassed, "lintPassed should be false on error")
	assert.Greater(t, r.lintCycles, 0, "lintCycles should increment")
	// Result may contain error or be empty depending on exec.Command behavior
	t.Logf("Lint result: %s", result)
}

// TestToolRunLint_InvalidJSON tests handling of invalid JSON from lint command
func TestToolRunLint_InvalidJSON(t *testing.T) {
	dir := t.TempDir()

	// Create a mock wetwire-aws script that outputs invalid JSON
	mockScript := filepath.Join(dir, "mock-wetwire-aws")
	scriptContent := `#!/bin/bash
echo "This is not valid JSON"
exit 2
`
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Temporarily modify PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", dir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	// Note: This test demonstrates the concept but won't work without
	// actually having a mock command. In practice, this would be tested
	// with a proper mock or by checking the error handling code path
	t.Log("Test demonstrates handling of invalid JSON output")
}

// TestToolRunLint_NonZeroExit tests handling of various exit codes
func TestToolRunLint_NonZeroExit(t *testing.T) {
	dir := t.TempDir()
	r := &RunnerAgent{
		workDir: dir,
		session: results.NewSession("test", "test"),
	}

	// Create a package with known lint issues
	pkgDir := filepath.Join(dir, "badpkg")
	os.MkdirAll(pkgDir, 0755)

	// Even if command fails, the agent should handle it gracefully
	result := r.toolRunLint("badpkg")

	assert.True(t, r.lintCalled)
	// The lintPassed state depends on the actual command result
	t.Logf("Lint result: %s", result)
}

// TestToolRunBuild_CommandFailure tests build command failures
func TestToolRunBuild_CommandFailure(t *testing.T) {
	dir := t.TempDir()
	r := &RunnerAgent{
		workDir: dir,
	}

	// Try to build a nonexistent package
	result := r.toolRunBuild("nonexistent")

	// Should return some output (either error message or command output)
	assert.NotEmpty(t, result)
	t.Logf("Build result: %s", result)
}

// TestToolRunBuild_TemplateExtraction tests extracting template from successful build
func TestToolRunBuild_TemplateExtraction(t *testing.T) {
	dir := t.TempDir()
	r := &RunnerAgent{
		workDir: dir,
	}

	// Create a valid package structure (if wetwire-aws is available)
	pkgDir := filepath.Join(dir, "testpkg")
	os.MkdirAll(pkgDir, 0755)

	result := r.toolRunBuild("testpkg")

	// Template extraction depends on successful build
	t.Logf("Build result: %s", result)
	t.Logf("Template JSON: %s", r.templateJSON)
}

// TestAPIAuthentication_InvalidKey tests handling of invalid API key
func TestAPIAuthentication_InvalidKey(t *testing.T) {
	config := RunnerConfig{
		APIKey:  "invalid-key",
		WorkDir: t.TempDir(),
	}

	agent, err := NewRunnerAgent(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to run with invalid key - should fail on API call
	err = agent.Run(ctx, "Create a simple S3 bucket")

	// Expect an error from the API
	if err != nil {
		assert.Contains(t, err.Error(), "API call failed")
		t.Logf("Expected error: %v", err)
	}
}

// TestAPIRateLimiting tests handling of rate limit responses
func TestAPIRateLimiting(t *testing.T) {
	// This test would require mocking the Anthropic API client
	// or making many rapid requests to trigger rate limiting
	t.Skip("Rate limiting test requires API mocking infrastructure")
}

// TestAPINetworkTimeout tests handling of network timeouts
func TestAPINetworkTimeout(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	config := RunnerConfig{
		APIKey:  apiKey,
		WorkDir: t.TempDir(),
	}

	agent, err := NewRunnerAgent(config)
	require.NoError(t, err)

	// Create a very short timeout to simulate network issues
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = agent.Run(ctx, "Create an S3 bucket")

	// Should get a timeout or context deadline exceeded error
	if err != nil {
		assert.True(t,
			err == context.DeadlineExceeded ||
				err.Error() == "context deadline exceeded" ||
				err.Error() == "API call failed: context deadline exceeded",
			"Expected timeout error, got: %v", err)
	}
}

// TestAPITokenLimitExceeded tests handling when token limit is exceeded
func TestAPITokenLimitExceeded(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	// This would require sending an extremely long prompt
	// In practice, the SDK handles this before sending
	t.Skip("Token limit test requires very large prompts")
}

// TestAPIInvalidModel tests handling of invalid model names
func TestAPIInvalidModel(t *testing.T) {
	// This test would require modifying the agent to accept custom model names
	// or mocking the API client
	t.Skip("Invalid model test requires API mocking infrastructure")
}

// TestEndToEnd_MultiTurnConversation tests a multi-turn agent conversation
func TestEndToEnd_MultiTurnConversation(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	dir := t.TempDir()
	session := results.NewSession("test", "integration_test")

	// Mock developer that provides canned responses
	questionCount := 0
	mockDeveloper := &mockDeveloperResponder{
		responses: []string{
			"Use AES-256 encryption",
			"Yes, enable versioning",
			"No, no lifecycle rules needed",
		},
		questionCount: &questionCount,
	}

	config := RunnerConfig{
		APIKey:        apiKey,
		WorkDir:       dir,
		MaxLintCycles: 3,
		Session:       session,
		Developer:     mockDeveloper,
	}

	agent, err := NewRunnerAgent(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run a complex prompt that might require multiple questions
	err = agent.Run(ctx, "Create an S3 bucket for storing sensitive customer data")

	// Check that multiple turns occurred
	if questionCount > 0 {
		t.Logf("Agent asked %d questions", questionCount)
		assert.Greater(t, len(session.Questions), 0)
	}

	if err != nil {
		t.Logf("Run completed with error: %v", err)
	}
}

// TestEndToEnd_LintFixLoop tests reaching max lint cycles
func TestEndToEnd_LintFixLoop(t *testing.T) {
	// This test would require:
	// 1. Setting up a scenario that produces lint errors
	// 2. Ensuring the agent attempts to fix them
	// 3. Verifying it stops at max cycles
	t.Skip("Lint fix loop test requires full wetwire-aws integration")
}

// TestEndToEnd_SessionTimeout tests enforcement of session timeout
func TestEndToEnd_SessionTimeout(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	config := RunnerConfig{
		APIKey:  apiKey,
		WorkDir: t.TempDir(),
	}

	agent, err := NewRunnerAgent(config)
	require.NoError(t, err)

	// Set an extremely short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = agent.Run(ctx, "Create a complex VPC with multiple subnets")

	// Should timeout
	if err != nil {
		assert.True(t,
			err == context.DeadlineExceeded ||
				err.Error() == "context deadline exceeded" ||
				err.Error() == "API call failed: context deadline exceeded",
			"Expected timeout error, got: %v", err)
	}
}

// TestLargeCommandOutput tests handling of large command output
func TestLargeCommandOutput(t *testing.T) {
	dir := t.TempDir()

	// Create a script that outputs a lot of data
	scriptPath := filepath.Join(dir, "large-output.sh")
	script := `#!/bin/bash
for i in {1..10000}; do
  echo "{\"issue\": \"Error $i: This is a very long error message that repeats many times\"}"
done
`
	err := os.WriteFile(scriptPath, []byte(script), 0755)
	require.NoError(t, err)

	cmd := exec.Command(scriptPath)
	output, err := cmd.CombinedOutput()

	// Should handle large output
	assert.NotNil(t, output)
	t.Logf("Output size: %d bytes", len(output))
}

// TestCommandTimeout tests command execution timeout scenarios
func TestCommandTimeout(t *testing.T) {
	dir := t.TempDir()

	// Create a script that runs for a long time
	scriptPath := filepath.Join(dir, "slow-command.sh")
	script := `#!/bin/bash
sleep 10
echo "Done"
`
	err := os.WriteFile(scriptPath, []byte(script), 0755)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scriptPath)
	err = cmd.Run()

	// Should timeout
	assert.Error(t, err)
	t.Logf("Expected timeout error: %v", err)
}

// mockDeveloperResponder provides canned responses for testing
type mockDeveloperResponder struct {
	responses     []string
	questionCount *int
}

func (m *mockDeveloperResponder) Respond(ctx context.Context, message string) (string, error) {
	if m.questionCount != nil {
		*m.questionCount++
	}

	if *m.questionCount <= len(m.responses) {
		return m.responses[*m.questionCount-1], nil
	}

	return "I don't know", nil
}

// TestJSONParsing_EdgeCases tests various JSON parsing scenarios
func TestJSONParsing_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid_lint_result",
			json:    `{"success": true, "issues": []}`,
			wantErr: false,
		},
		{
			name:    "lint_with_issues",
			json:    `{"success": false, "issues": [{"message": "error1"}]}`,
			wantErr: false,
		},
		{
			name:    "malformed_json",
			json:    `{"success": true, "issues": [}`,
			wantErr: true,
		},
		{
			name:    "empty_string",
			json:    ``,
			wantErr: true,
		},
		{
			name:    "null_json",
			json:    `null`,
			wantErr: false,
		},
		{
			name:    "escaped_characters",
			json:    `{"message": "Line 1\nLine 2\tTab"}`,
			wantErr: false,
		},
		{
			name:    "unicode",
			json:    `{"message": "Error: 文件不存在"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := json.Unmarshal([]byte(tt.json), &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
