package personas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	all := All()
	assert.Len(t, all, 5)

	names := make(map[string]bool)
	for _, p := range all {
		names[p.Name] = true
		assert.NotEmpty(t, p.Description)
		assert.NotEmpty(t, p.SystemPrompt)
		assert.NotEmpty(t, p.ExpectedBehavior)
		assert.NotEmpty(t, p.Traits, "persona %s should have traits", p.Name)
	}

	assert.True(t, names["beginner"])
	assert.True(t, names["intermediate"])
	assert.True(t, names["expert"])
	assert.True(t, names["terse"])
	assert.True(t, names["verbose"])
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"lowercase", "beginner", false},
		{"uppercase", "EXPERT", false},
		{"mixed case", "Terse", false},
		{"unknown", "unknown", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Get(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, p.Name)
			}
		})
	}
}

func TestNames(t *testing.T) {
	names := Names()
	assert.Len(t, names, 5)
	assert.Contains(t, names, "beginner")
	assert.Contains(t, names, "expert")
}

func TestPersonaContent(t *testing.T) {
	// Verify beginner persona encourages safe defaults
	assert.Contains(t, Beginner.SystemPrompt, "uncertain")
	assert.Contains(t, Beginner.ExpectedBehavior, "safe defaults")
	assert.Contains(t, Beginner.Traits, "uncertain")

	// Verify expert persona is precise
	assert.Contains(t, Expert.SystemPrompt, "precise")
	assert.Contains(t, Expert.ExpectedBehavior, "minimal questions")
	assert.Contains(t, Expert.Traits, "precise")

	// Verify terse persona is concise
	assert.Contains(t, Terse.SystemPrompt, "few words")
	assert.Contains(t, Terse.Traits, "concise")

	// Verify verbose persona over-explains
	assert.Contains(t, Verbose.SystemPrompt, "verbose")
	assert.Contains(t, Verbose.Traits, "wordy")
}

func TestPersonaTraits(t *testing.T) {
	// Each persona should have at least 3 traits
	for _, p := range All() {
		assert.GreaterOrEqual(t, len(p.Traits), 3, "persona %s should have at least 3 traits", p.Name)
	}
}
