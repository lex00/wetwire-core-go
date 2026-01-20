package personas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltIn(t *testing.T) {
	builtIn := BuiltIn()
	assert.Len(t, builtIn, 3)

	names := make(map[string]bool)
	for _, p := range builtIn {
		names[p.Name] = true
		assert.NotEmpty(t, p.Description)
		assert.NotEmpty(t, p.SystemPrompt)
		assert.NotEmpty(t, p.ExpectedBehavior)
		assert.NotEmpty(t, p.Traits, "persona %s should have traits", p.Name)
	}

	assert.True(t, names["beginner"])
	assert.True(t, names["intermediate"])
	assert.True(t, names["expert"])
}

func TestAll(t *testing.T) {
	// Clear any custom personas from other tests
	ClearCustom()

	all := All()
	assert.Len(t, all, 3, "All() should return 3 built-in personas when no custom registered")

	// Register a custom persona
	err := Register(Persona{
		Name:             "custom",
		Description:      "Custom test persona",
		SystemPrompt:     "You are a custom persona.",
		Traits:           []string{"custom", "test"},
		ExpectedBehavior: "Custom behavior",
	})
	require.NoError(t, err)

	all = All()
	assert.Len(t, all, 4, "All() should include custom persona")

	// Clean up
	ClearCustom()
}

func TestGet(t *testing.T) {
	ClearCustom()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"lowercase", "beginner", false},
		{"uppercase", "EXPERT", false},
		{"mixed case", "Intermediate", false},
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

func TestGetCustom(t *testing.T) {
	ClearCustom()
	defer ClearCustom()

	// Register custom persona
	err := Register(Persona{
		Name:             "mycustom",
		Description:      "My custom persona",
		SystemPrompt:     "Custom system prompt",
		Traits:           []string{"custom"},
		ExpectedBehavior: "Custom behavior",
	})
	require.NoError(t, err)

	// Should be able to get it
	p, err := Get("mycustom")
	require.NoError(t, err)
	assert.Equal(t, "mycustom", p.Name)

	// Case insensitive
	p, err = Get("MYCUSTOM")
	require.NoError(t, err)
	assert.Equal(t, "mycustom", p.Name)
}

func TestNames(t *testing.T) {
	ClearCustom()

	names := Names()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "beginner")
	assert.Contains(t, names, "intermediate")
	assert.Contains(t, names, "expert")
}

func TestBuiltInNames(t *testing.T) {
	names := BuiltInNames()
	assert.Equal(t, []string{"beginner", "intermediate", "expert"}, names)
}

func TestPersonaContent(t *testing.T) {
	// Verify beginner persona encourages safe defaults
	assert.Contains(t, Beginner.SystemPrompt, "uncertain")
	assert.Contains(t, Beginner.ExpectedBehavior, "safe defaults")
	assert.Contains(t, Beginner.Traits, "uncertain")

	// Verify intermediate persona has experience
	assert.Contains(t, Intermediate.SystemPrompt, "moderate")
	assert.Contains(t, Intermediate.Traits, "experienced")

	// Verify expert persona is precise
	assert.Contains(t, Expert.SystemPrompt, "precise")
	assert.Contains(t, Expert.ExpectedBehavior, "minimal questions")
	assert.Contains(t, Expert.Traits, "precise")
}

func TestPersonaTraits(t *testing.T) {
	// Each built-in persona should have at least 3 traits
	for _, p := range BuiltIn() {
		assert.GreaterOrEqual(t, len(p.Traits), 3, "persona %s should have at least 3 traits", p.Name)
	}
}

func TestRegister(t *testing.T) {
	ClearCustom()
	defer ClearCustom()

	// Register custom persona
	err := Register(Persona{
		Name:             "testpersona",
		Description:      "Test persona",
		SystemPrompt:     "Test prompt",
		Traits:           []string{"test"},
		ExpectedBehavior: "Test behavior",
	})
	require.NoError(t, err)

	// Cannot override built-in
	err = Register(Persona{Name: "beginner"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot override built-in")

	// Can register another custom
	err = Register(Persona{Name: "another"})
	require.NoError(t, err)

	// Can override custom
	err = Register(Persona{Name: "testpersona", Description: "Updated"})
	require.NoError(t, err)

	p, _ := Get("testpersona")
	assert.Equal(t, "Updated", p.Description)
}

func TestUnregister(t *testing.T) {
	ClearCustom()
	defer ClearCustom()

	_ = Register(Persona{Name: "toremove"})
	_, err := Get("toremove")
	require.NoError(t, err)

	Unregister("toremove")
	_, err = Get("toremove")
	require.Error(t, err)
}

func TestClearCustom(t *testing.T) {
	ClearCustom()

	_ = Register(Persona{Name: "a"})
	_ = Register(Persona{Name: "b"})
	assert.Len(t, All(), 5) // 3 built-in + 2 custom

	ClearCustom()
	assert.Len(t, All(), 3) // just built-in
}
