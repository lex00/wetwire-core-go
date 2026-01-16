// This file contains edge case tests for the results package
package results

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lex00/wetwire-core-go/agent/scoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatMarkdown_SpecialCharacters tests markdown generation with special characters
func TestFormatMarkdown_SpecialCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "backticks",
			content:  "Use `kubectl get pods` to list pods",
			expected: []string{"`kubectl get pods`"},
		},
		{
			name:     "asterisks",
			content:  "This is *important* and **very important**",
			expected: []string{"*important*", "**very important**"},
		},
		{
			name:     "underscores",
			content:  "The variable_name_with_underscores is valid",
			expected: []string{"variable_name_with_underscores"},
		},
		{
			name:     "brackets",
			content:  "Check [this link](http://example.com) for more",
			expected: []string{"[this link]"},
		},
		{
			name:     "angle_brackets",
			content:  "Use <Context> for state management",
			expected: []string{"<Context>"},
		},
		{
			name:     "hash_symbols",
			content:  "# This looks like a header",
			expected: []string{"# This looks like a header"},
		},
		{
			name:     "code_blocks",
			content:  "```go\npackage main\n```",
			expected: []string{"```go", "package main", "```"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("test", "test")
			session.InitialPrompt = tt.content
			session.AddMessage("developer", tt.content)
			session.Complete()

			writer := NewWriter(t.TempDir())
			md := writer.formatMarkdown(session)

			for _, exp := range tt.expected {
				assert.Contains(t, md, exp)
			}
		})
	}
}

// TestFormatMarkdown_VeryLongMessages tests handling of very long messages
func TestFormatMarkdown_VeryLongMessages(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")

	// Create a very long message (10K+ chars)
	longMessage := strings.Repeat("This is a very long message. ", 500)
	session.InitialPrompt = longMessage
	session.AddMessage("developer", longMessage)
	session.Complete()

	writer := NewWriter(t.TempDir())
	md := writer.formatMarkdown(session)

	// Should still generate valid markdown
	assert.Contains(t, md, "# Session Results")
	assert.Contains(t, md, longMessage)
	assert.Greater(t, len(md), 10000)
}

// TestFormatMarkdown_EmptySections tests markdown with empty sections
func TestFormatMarkdown_EmptySections(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")
	session.InitialPrompt = "Test prompt"
	// No questions, no lint cycles, no messages added
	session.Complete()

	writer := NewWriter(t.TempDir())
	md := writer.formatMarkdown(session)

	// Should have headers but no content for empty sections
	assert.Contains(t, md, "# Session Results")
	assert.Contains(t, md, "## Initial Prompt")
	assert.NotContains(t, md, "## Clarifying Questions")
	assert.NotContains(t, md, "## Lint Cycles")
}

// TestFormatMarkdown_UnicodeEmoji tests unicode and emoji in content
func TestFormatMarkdown_UnicodeEmoji(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")
	session.InitialPrompt = "🚀 Deploy to production 生产环境"
	session.AddMessage("developer", "Status: ✅ Passed")
	session.AddMessage("runner", "Error: ❌ Failed")
	session.AddQuestion("Language? 语言?", "中文 🇨🇳")
	session.Complete()

	writer := NewWriter(t.TempDir())
	md := writer.formatMarkdown(session)

	// Unicode and emoji should be preserved
	assert.Contains(t, md, "🚀")
	assert.Contains(t, md, "✅")
	assert.Contains(t, md, "❌")
	assert.Contains(t, md, "生产环境")
	assert.Contains(t, md, "中文")
	assert.Contains(t, md, "🇨🇳")
}

// TestFormatMarkdown_MarkdownInjection tests potential markdown injection
func TestFormatMarkdown_MarkdownInjection(t *testing.T) {
	t.Parallel()

	maliciousInputs := []string{
		"](http://evil.com)",
		"<!-- HTML comment -->",
		"<script>alert('xss')</script>",
		"![image](http://evil.com/image.png)",
		"[link](javascript:alert('xss'))",
	}

	for _, input := range maliciousInputs {
		session := NewSession("test", "test")
		session.InitialPrompt = input
		session.AddMessage("developer", input)
		session.AddQuestion(input, input)
		session.Complete()

		writer := NewWriter(t.TempDir())
		md := writer.formatMarkdown(session)

		// Markdown should contain the raw text (not executed)
		assert.Contains(t, md, input)
		// Verify the output is still valid markdown structure
		assert.Contains(t, md, "# Session Results")
	}
}

// TestFormatMarkdown_LargeLintCycles tests many lint cycles
func TestFormatMarkdown_LargeLintCycles(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")
	session.InitialPrompt = "Test"

	// Add many lint cycles
	for i := 0; i < 10; i++ {
		issues := []string{
			"Error 1 in cycle " + string(rune(i)),
			"Error 2 in cycle " + string(rune(i)),
		}
		session.AddLintCycle(issues, i, i == 9) // Only last one passes
	}
	session.Complete()

	writer := NewWriter(t.TempDir())
	md := writer.formatMarkdown(session)

	assert.Contains(t, md, "## Lint Cycles")
	assert.Contains(t, md, "Cycle 1")
	assert.Contains(t, md, "Cycle 10")
}

// TestWriter_EdgeCases tests edge cases for writing results
func TestWriter_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() (*Session, string)
		wantErr bool
	}{
		{
			name: "read_only_directory",
			setup: func() (*Session, string) {
				dir := t.TempDir()
				// Make directory read-only
				os.Chmod(dir, 0444)
				return NewSession("test", "test"), dir
			},
			wantErr: true,
		},
		{
			name: "very_long_persona_name",
			setup: func() (*Session, string) {
				dir := t.TempDir()
				longName := strings.Repeat("persona", 100)
				session := NewSession(longName, "test")
				session.InitialPrompt = "Test"
				session.Complete()
				return session, dir
			},
			wantErr: true, // File name too long error expected
		},
		{
			name: "special_chars_in_persona",
			setup: func() (*Session, string) {
				dir := t.TempDir()
				session := NewSession("persona/../../../etc", "test")
				session.InitialPrompt = "Test"
				session.Complete()
				return session, dir
			},
			wantErr: false,
		},
		{
			name: "empty_session",
			setup: func() (*Session, string) {
				dir := t.TempDir()
				session := NewSession("", "")
				session.Complete()
				return session, dir
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, dir := tt.setup()
			writer := NewWriter(dir)
			err := writer.Write(session)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSession_AddMessage_EdgeCases tests edge cases for adding messages
func TestSession_AddMessage_EdgeCases(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")

	// Empty messages
	session.AddMessage("", "")
	assert.Len(t, session.Messages, 1)
	assert.Equal(t, "", session.Messages[0].Role)
	assert.Equal(t, "", session.Messages[0].Content)

	// Very long role and content
	longString := strings.Repeat("x", 10000)
	session.AddMessage(longString, longString)
	assert.Len(t, session.Messages, 2)
	assert.Equal(t, longString, session.Messages[1].Role)
	assert.Equal(t, longString, session.Messages[1].Content)

	// Unicode in role and content
	session.AddMessage("开发者", "内容")
	assert.Len(t, session.Messages, 3)
	assert.Equal(t, "开发者", session.Messages[2].Role)
}

// TestSession_AddQuestion_EdgeCases tests edge cases for adding questions
func TestSession_AddQuestion_EdgeCases(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")

	// Empty question and answer
	session.AddQuestion("", "")
	assert.Len(t, session.Questions, 1)

	// Very long question and answer
	longString := strings.Repeat("x", 10000)
	session.AddQuestion(longString, longString)
	assert.Len(t, session.Questions, 2)

	// Special characters
	session.AddQuestion("What is `kubectl`?", "It's a *CLI* tool")
	assert.Len(t, session.Questions, 3)

	// Many questions
	for i := 0; i < 100; i++ {
		session.AddQuestion("Q", "A")
	}
	assert.Len(t, session.Questions, 103)
}

// TestSession_AddLintCycle_EdgeCases tests edge cases for lint cycles
func TestSession_AddLintCycle_EdgeCases(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")

	// Empty issues
	session.AddLintCycle([]string{}, 0, true)
	assert.Len(t, session.LintCycles, 1)
	assert.Equal(t, 0, session.LintCycles[0].IssueCount)
	assert.True(t, session.LintCycles[0].Passed)

	// Nil issues
	session.AddLintCycle(nil, 0, false)
	assert.Len(t, session.LintCycles, 2)
	assert.Equal(t, 0, session.LintCycles[1].IssueCount)

	// Many issues
	manyIssues := make([]string, 1000)
	for i := range manyIssues {
		manyIssues[i] = "Issue"
	}
	session.AddLintCycle(manyIssues, 500, false)
	assert.Len(t, session.LintCycles, 3)
	assert.Equal(t, 1000, session.LintCycles[2].IssueCount)
	assert.Equal(t, 500, session.LintCycles[2].FixedCount)

	// Negative fixed count
	session.AddLintCycle([]string{"err"}, -10, false)
	assert.Len(t, session.LintCycles, 4)
	assert.Equal(t, -10, session.LintCycles[3].FixedCount)
}

// TestSession_Duration_EdgeCases tests edge cases for duration calculation
func TestSession_Duration_EdgeCases(t *testing.T) {
	t.Parallel()

	// Session with zero start time
	session := &Session{
		StartTime: time.Time{},
		EndTime:   time.Time{},
	}
	duration := session.Duration()
	assert.Greater(t, duration, time.Duration(0))

	// Session with future end time
	session = NewSession("test", "test")
	session.EndTime = time.Now().Add(1 * time.Hour)
	duration = session.Duration()
	assert.Greater(t, duration, time.Duration(0))

	// Session with end time before start time
	session = NewSession("test", "test")
	session.EndTime = session.StartTime.Add(-1 * time.Hour)
	duration = session.Duration()
	assert.Less(t, duration, time.Duration(0))
}

// TestSession_JSON_EdgeCases tests JSON marshaling edge cases
func TestSession_JSON_EdgeCases(t *testing.T) {
	t.Parallel()

	session := NewSession("test", "test")
	session.InitialPrompt = strings.Repeat("x", 10000)

	// Add messages with tool calls
	msg := Message{
		Role:      "runner",
		Content:   "Creating resources",
		Timestamp: time.Now(),
		ToolCalls: []ToolCall{
			{
				Name:   "write_file",
				Input:  strings.Repeat("y", 5000),
				Output: strings.Repeat("z", 5000),
			},
		},
	}
	session.Messages = append(session.Messages, msg)

	// Add many questions
	for i := 0; i < 100; i++ {
		session.AddQuestion("Q", "A")
	}

	// Add score with all dimensions
	score := scoring.NewScore("test", "test")
	score.Completeness.Rating = 3
	score.LintQuality.Rating = 2
	score.OutputValidity.Rating = 0
	score.QuestionEfficiency.Rating = 3
	session.Score = score

	session.Complete()

	// Marshal and unmarshal
	data, err := json.Marshal(session)
	require.NoError(t, err)
	assert.Greater(t, len(data), 10000)

	var parsed Session
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, session.Persona, parsed.Persona)
	assert.Len(t, parsed.Questions, 100)
	assert.NotNil(t, parsed.Score)
}

// TestGenerateID_Uniqueness tests that generated IDs are unique
func TestGenerateID_Uniqueness(t *testing.T) {
	// Note: Cannot run in parallel as it relies on sequential time passage
	// Only test a few IDs with sufficient spacing due to second-level precision
	ids := make(map[string]bool)
	for i := 0; i < 3; i++ {
		// Sleep between each generation to ensure different timestamps
		// The ID format uses second precision, so we need at least 1 second between calls
		if i > 0 {
			time.Sleep(1100 * time.Millisecond)
		}
		id := generateID()
		assert.False(t, ids[id], "ID should be unique: %s", id)
		ids[id] = true
	}
}

// TestGenerateID_Format tests the format of generated IDs
func TestGenerateID_Format(t *testing.T) {
	t.Parallel()

	id := generateID()

	// Should start with "session_"
	assert.True(t, strings.HasPrefix(id, "session_"))

	// Should contain a timestamp
	parts := strings.Split(id, "_")
	assert.Len(t, parts, 3)
	assert.NotEmpty(t, parts[1]) // Date part
	assert.NotEmpty(t, parts[2]) // Time part
}

// TestWriter_Write_FilePermissions tests file permissions after writing
func TestWriter_Write_FilePermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	session := NewSession("test", "test")
	session.InitialPrompt = "Test"
	session.Complete()

	writer := NewWriter(dir)
	err := writer.Write(session)
	require.NoError(t, err)

	// Check file permissions
	files := []string{
		filepath.Join(dir, "test", "RESULTS.md"),
		filepath.Join(dir, "test", "session.json"),
	}

	for _, file := range files {
		info, err := os.Stat(file)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
		// File should be readable
		content, err := os.ReadFile(file)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
	}
}

// TestWriter_Write_ConcurrentWrites tests concurrent writes to the same directory
func TestWriter_Write_ConcurrentWrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writer := NewWriter(dir)

	done := make(chan bool, 10)

	// Spawn multiple goroutines writing simultaneously
	for i := 0; i < 10; i++ {
		go func(idx int) {
			session := NewSession("concurrent", "test")
			session.InitialPrompt = "Test"
			session.Complete()
			err := writer.Write(session)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all writes to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify files were created (last write wins)
	resultsPath := filepath.Join(dir, "concurrent", "RESULTS.md")
	assert.FileExists(t, resultsPath)
}

// TestMessage_ToolCalls_EdgeCases tests edge cases for tool calls in messages
func TestMessage_ToolCalls_EdgeCases(t *testing.T) {
	t.Parallel()

	// Empty tool calls
	msg := Message{
		Role:      "runner",
		Content:   "Test",
		Timestamp: time.Now(),
		ToolCalls: []ToolCall{},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var parsed Message
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Empty(t, parsed.ToolCalls)

	// Many tool calls
	manyToolCalls := make([]ToolCall, 100)
	for i := range manyToolCalls {
		manyToolCalls[i] = ToolCall{
			Name:   "tool",
			Input:  "input",
			Output: "output",
		}
	}

	msg.ToolCalls = manyToolCalls
	data, err = json.Marshal(msg)
	require.NoError(t, err)
	assert.Greater(t, len(data), 1000)
}

// TestLintCycle_EdgeCases tests edge cases for lint cycle structure
func TestLintCycle_EdgeCases(t *testing.T) {
	t.Parallel()

	cycle := LintCycle{
		Cycle:      0,
		IssueCount: 0,
		Issues:     nil,
		FixedCount: 0,
		Passed:     false,
	}

	data, err := json.Marshal(cycle)
	require.NoError(t, err)

	var parsed LintCycle
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, cycle.Cycle, parsed.Cycle)
	assert.Equal(t, cycle.IssueCount, parsed.IssueCount)
	assert.Nil(t, parsed.Issues)
	assert.Equal(t, cycle.Passed, parsed.Passed)
}
