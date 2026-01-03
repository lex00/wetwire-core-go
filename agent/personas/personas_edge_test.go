// This file contains edge case tests for the personas package
package personas

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGet_EdgeCases tests edge cases for retrieving personas
func TestGet_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantName  string
		wantError bool
	}{
		{
			name:      "exact_match_lowercase",
			input:     "beginner",
			wantName:  "beginner",
			wantError: false,
		},
		{
			name:      "exact_match_uppercase",
			input:     "EXPERT",
			wantName:  "expert",
			wantError: false,
		},
		{
			name:      "mixed_case",
			input:     "InTeRmEdIaTe",
			wantName:  "intermediate",
			wantError: false,
		},
		{
			name:      "with_leading_space",
			input:     " terse",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "with_trailing_space",
			input:     "verbose ",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "empty_string",
			input:     "",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "unknown_persona",
			input:     "unknown",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "partial_match",
			input:     "beg",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "numeric_input",
			input:     "123",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "special_characters",
			input:     "expert!",
			wantName:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			persona, err := Get(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown persona")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, persona.Name)
			}
		})
	}
}

// TestAll_Completeness tests that All() returns all expected personas
func TestAll_Completeness(t *testing.T) {
	t.Parallel()

	personas := All()

	assert.Len(t, personas, 5)

	names := make(map[string]bool)
	for _, p := range personas {
		names[p.Name] = true
	}

	expectedNames := []string{"beginner", "intermediate", "expert", "terse", "verbose"}
	for _, name := range expectedNames {
		assert.True(t, names[name], "Missing persona: %s", name)
	}
}

// TestAll_Uniqueness tests that all personas have unique names
func TestAll_Uniqueness(t *testing.T) {
	t.Parallel()

	personas := All()
	names := make(map[string]bool)

	for _, p := range personas {
		assert.False(t, names[p.Name], "Duplicate persona name: %s", p.Name)
		names[p.Name] = true
	}
}

// TestAll_ValidStructure tests that all personas have valid structure
func TestAll_ValidStructure(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			assert.NotEmpty(t, p.Name, "Name should not be empty")
			assert.NotEmpty(t, p.Description, "Description should not be empty")
			assert.NotEmpty(t, p.SystemPrompt, "SystemPrompt should not be empty")
			assert.NotEmpty(t, p.ExpectedBehavior, "ExpectedBehavior should not be empty")
		})
	}
}

// TestNames_Completeness tests that Names() returns all persona names
func TestNames_Completeness(t *testing.T) {
	t.Parallel()

	names := Names()

	assert.Len(t, names, 5)

	expectedNames := []string{"beginner", "intermediate", "expert", "terse", "verbose"}
	for _, expected := range expectedNames {
		assert.Contains(t, names, expected)
	}
}

// TestNames_Order tests that Names() returns personas in expected order
func TestNames_Order(t *testing.T) {
	t.Parallel()

	names := Names()
	all := All()

	assert.Len(t, names, len(all))

	for i, persona := range all {
		assert.Equal(t, persona.Name, names[i])
	}
}

// TestBeginner_Characteristics tests Beginner persona characteristics
func TestBeginner_Characteristics(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "beginner", Beginner.Name)
	assert.Contains(t, strings.ToLower(Beginner.Description), "new")
	assert.Contains(t, strings.ToLower(Beginner.SystemPrompt), "new to aws")
	assert.Contains(t, strings.ToLower(Beginner.SystemPrompt), "uncertain")
	assert.Contains(t, Beginner.ExpectedBehavior, "safe defaults")
}

// TestIntermediate_Characteristics tests Intermediate persona characteristics
func TestIntermediate_Characteristics(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "intermediate", Intermediate.Name)
	assert.Contains(t, Intermediate.Description, "AWS experience")
	assert.Contains(t, strings.ToLower(Intermediate.SystemPrompt), "moderate")
	assert.Contains(t, Intermediate.ExpectedBehavior, "fill in details")
}

// TestExpert_Characteristics tests Expert persona characteristics
func TestExpert_Characteristics(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "expert", Expert.Name)
	assert.Contains(t, Expert.Description, "Deep AWS knowledge")
	assert.Contains(t, strings.ToLower(Expert.SystemPrompt), "senior")
	assert.Contains(t, strings.ToLower(Expert.SystemPrompt), "expertise")
	assert.Contains(t, Expert.ExpectedBehavior, "exactly as specified")
}

// TestTerse_Characteristics tests Terse persona characteristics
func TestTerse_Characteristics(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "terse", Terse.Name)
	assert.Contains(t, Terse.Description, "Minimal")
	assert.Contains(t, strings.ToLower(Terse.SystemPrompt), "concise")
	assert.Contains(t, Terse.ExpectedBehavior, "infer")
}

// TestVerbose_Characteristics tests Verbose persona characteristics
func TestVerbose_Characteristics(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "verbose", Verbose.Name)
	assert.Contains(t, strings.ToLower(Verbose.Description), "over-explain")
	assert.Contains(t, strings.ToLower(Verbose.SystemPrompt), "verbose")
	assert.Contains(t, Verbose.ExpectedBehavior, "filter")
}

// TestSystemPrompt_Length tests that system prompts are substantial
func TestSystemPrompt_Length(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			assert.Greater(t, len(p.SystemPrompt), 50,
				"SystemPrompt should be substantial for %s", p.Name)
		})
	}
}

// TestSystemPrompt_NoTemplateErrors tests that system prompts don't have obvious errors
func TestSystemPrompt_NoTemplateErrors(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			// Check for common template errors
			assert.NotContains(t, p.SystemPrompt, "{{",
				"SystemPrompt should not contain template placeholders")
			assert.NotContains(t, p.SystemPrompt, "}}",
				"SystemPrompt should not contain template placeholders")
			assert.NotContains(t, p.SystemPrompt, "TODO",
				"SystemPrompt should not contain TODOs")
			assert.NotContains(t, p.SystemPrompt, "FIXME",
				"SystemPrompt should not contain FIXMEs")
		})
	}
}

// TestDescription_Clarity tests that descriptions are clear and concise
func TestDescription_Clarity(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			assert.Greater(t, len(p.Description), 10,
				"Description should be meaningful for %s", p.Name)
			assert.Less(t, len(p.Description), 200,
				"Description should be concise for %s", p.Name)
		})
	}
}

// TestExpectedBehavior_Clarity tests that expected behaviors are clear
func TestExpectedBehavior_Clarity(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			assert.Greater(t, len(p.ExpectedBehavior), 10,
				"ExpectedBehavior should be meaningful for %s", p.Name)
			// Should contain "Runner" or "agent" or "should"
			lower := strings.ToLower(p.ExpectedBehavior)
			assert.True(t,
				strings.Contains(lower, "runner") ||
					strings.Contains(lower, "agent") ||
					strings.Contains(lower, "should"),
				"ExpectedBehavior should describe what the Runner should do")
		})
	}
}

// TestPersona_Immutability tests that predefined personas are not modified
func TestPersona_Immutability(t *testing.T) {
	t.Parallel()

	// Get original values
	originalName := Beginner.Name
	originalDesc := Beginner.Description

	// Attempt to get the persona
	p, err := Get("beginner")
	require.NoError(t, err)

	// Modify the returned persona
	p.Name = "modified"
	p.Description = "modified"

	// Original should be unchanged
	assert.Equal(t, originalName, Beginner.Name)
	assert.Equal(t, originalDesc, Beginner.Description)
}

// TestGet_CaseInsensitivity tests case insensitive retrieval
func TestGet_CaseInsensitivity(t *testing.T) {
	t.Parallel()

	cases := []string{
		"beginner",
		"BEGINNER",
		"Beginner",
		"bEgInNeR",
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			persona, err := Get(c)
			require.NoError(t, err)
			assert.Equal(t, "beginner", persona.Name)
		})
	}
}

// TestAll_ReturnsCopies tests that All() returns a new slice each time
func TestAll_ReturnsCopies(t *testing.T) {
	t.Parallel()

	all1 := All()
	all2 := All()

	// Should be equal but not the same slice
	assert.Equal(t, all1, all2)
	assert.NotSame(t, &all1, &all2)

	// Modifying one shouldn't affect the other
	all1[0].Name = "modified"
	assert.NotEqual(t, all1[0].Name, all2[0].Name)
}

// TestNames_ReturnsCopies tests that Names() returns a new slice each time
func TestNames_ReturnsCopies(t *testing.T) {
	t.Parallel()

	names1 := Names()
	names2 := Names()

	// Should be equal but not the same slice
	assert.Equal(t, names1, names2)
	assert.NotSame(t, &names1, &names2)

	// Modifying one shouldn't affect the other
	names1[0] = "modified"
	assert.NotEqual(t, names1[0], names2[0])
}

// TestPersona_SystemPromptExamples tests that system prompts contain examples
func TestPersona_SystemPromptExamples(t *testing.T) {
	t.Parallel()

	// Terse and Verbose should have communication examples
	terseLower := strings.ToLower(Terse.SystemPrompt)
	assert.True(t,
		strings.Contains(terseLower, "example") ||
			strings.Contains(terseLower, "log bucket") ||
			strings.Contains(terseLower, "lambda"),
		"Terse persona should have communication examples")

	verboseLower := strings.ToLower(Verbose.SystemPrompt)
	assert.True(t,
		strings.Contains(verboseLower, "example") ||
			strings.Contains(verboseLower, "instead of"),
		"Verbose persona should have communication examples")
}

// TestGet_ErrorMessage tests error message quality
func TestGet_ErrorMessage(t *testing.T) {
	t.Parallel()

	_, err := Get("nonexistent")
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "unknown persona")
	assert.Contains(t, errMsg, "nonexistent")
	assert.Contains(t, errMsg, "available")

	// Should list available personas
	for _, name := range []string{"beginner", "intermediate", "expert", "terse", "verbose"} {
		assert.Contains(t, errMsg, name,
			"Error message should list available persona: %s", name)
	}
}

// TestPersona_FieldsNotEmpty tests that no persona has empty fields
func TestPersona_FieldsNotEmpty(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			assert.NotEmpty(t, p.Name)
			assert.NotEmpty(t, p.Description)
			assert.NotEmpty(t, p.SystemPrompt)
			assert.NotEmpty(t, p.ExpectedBehavior)
		})
	}
}

// TestPersona_SystemPromptStructure tests system prompt structure
func TestPersona_SystemPromptStructure(t *testing.T) {
	t.Parallel()

	personas := All()

	for _, p := range personas {
		t.Run(p.Name, func(t *testing.T) {
			// Should be proper sentences (end with punctuation)
			trimmed := strings.TrimSpace(p.SystemPrompt)
			lastChar := trimmed[len(trimmed)-1]
			assert.True(t,
				lastChar == '.' || lastChar == '"' || lastChar == '`',
				"SystemPrompt should end with proper punctuation for %s", p.Name)
		})
	}
}

// TestBeginner_KeyPhrases tests that Beginner has uncertainty phrases
func TestBeginner_KeyPhrases(t *testing.T) {
	t.Parallel()

	lower := strings.ToLower(Beginner.SystemPrompt)

	uncertaintyPhrases := []string{"uncertain", "not sure", "should", "maybe", "think"}
	found := false
	for _, phrase := range uncertaintyPhrases {
		if strings.Contains(lower, phrase) {
			found = true
			break
		}
	}
	assert.True(t, found, "Beginner should express uncertainty")
}

// TestExpert_KeyPhrases tests that Expert has confidence phrases
func TestExpert_KeyPhrases(t *testing.T) {
	t.Parallel()

	lower := strings.ToLower(Expert.SystemPrompt)

	expertisePhrases := []string{"know", "precise", "specific", "exactly", "deep"}
	found := false
	for _, phrase := range expertisePhrases {
		if strings.Contains(lower, phrase) {
			found = true
			break
		}
	}
	assert.True(t, found, "Expert should demonstrate expertise")
}

// TestTerse_KeyPhrases tests that Terse emphasizes brevity
func TestTerse_KeyPhrases(t *testing.T) {
	t.Parallel()

	lower := strings.ToLower(Terse.SystemPrompt)

	brevityPhrases := []string{"concise", "minimal", "few words", "short"}
	found := false
	for _, phrase := range brevityPhrases {
		if strings.Contains(lower, phrase) {
			found = true
			break
		}
	}
	assert.True(t, found, "Terse should emphasize brevity")
}

// TestVerbose_KeyPhrases tests that Verbose emphasizes detail
func TestVerbose_KeyPhrases(t *testing.T) {
	t.Parallel()

	lower := strings.ToLower(Verbose.SystemPrompt)

	verbosePhrases := []string{"verbose", "explain", "context", "detail"}
	found := false
	for _, phrase := range verbosePhrases {
		if strings.Contains(lower, phrase) {
			found = true
			break
		}
	}
	assert.True(t, found, "Verbose should emphasize verbosity")
}
