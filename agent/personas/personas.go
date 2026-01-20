// Package personas defines AI developer personas for testing scenarios.
//
// Personas simulate different types of users interacting with the Runner agent.
// Each persona has a distinct communication style that tests different aspects
// of the Runner's capabilities.
//
// Built-in personas: beginner, intermediate, expert
// Custom personas can be registered using Register().
package personas

import (
	"fmt"
	"strings"
	"sync"
)

// Persona represents a simulated developer with specific characteristics.
type Persona struct {
	// Name is the persona identifier (e.g., "beginner", "expert")
	Name string

	// Description explains the persona's characteristics
	Description string

	// SystemPrompt is injected into the Developer agent's system message
	SystemPrompt string

	// Traits are key characteristics that influence the persona's behavior
	Traits []string

	// ExpectedBehavior describes what the Runner should do for this persona
	ExpectedBehavior string
}

// Predefined personas for testing
var (
	// Beginner simulates a new user who is uncertain and needs guidance.
	Beginner = Persona{
		Name:        "beginner",
		Description: "New to infrastructure, uncertain about best practices, asks many questions",
		Traits:      []string{"uncertain", "questioning", "needs-guidance", "vague"},
		SystemPrompt: `You are a developer who is new to infrastructure-as-code.
You are uncertain about best practices and often ask questions like:
- "Should this be encrypted?"
- "What's the difference between these options?"
- "Is this secure enough?"

Be vague about requirements. Use phrases like "I think I need..." or "maybe something like...".
Ask for recommendations rather than specifying exact configurations.
Express uncertainty about security, naming, and configuration choices.`,
		ExpectedBehavior: "Runner should make safe defaults, explain choices, and guide the user",
	}

	// Intermediate simulates a user with some infrastructure knowledge.
	Intermediate = Persona{
		Name:        "intermediate",
		Description: "Has experience, knows what they want but may miss details",
		Traits:      []string{"experienced", "clear", "detail-oriented", "asks-follow-ups"},
		SystemPrompt: `You are a developer with moderate infrastructure experience.
You know the basics but might miss some details or best practices.
You can specify what you want but may not know the optimal configuration.

Provide clear requirements but leave some details unspecified.
You understand technical references and can make decisions when asked.
Occasionally ask for clarification on advanced features.`,
		ExpectedBehavior: "Runner should fill in details while respecting stated requirements",
	}

	// Expert simulates a senior engineer with precise requirements.
	Expert = Persona{
		Name:        "expert",
		Description: "Deep infrastructure knowledge, precise requirements, minimal hand-holding needed",
		Traits:      []string{"precise", "technical", "self-sufficient", "detailed"},
		SystemPrompt: `You are a senior infrastructure engineer with deep expertise.
You know exactly what you want and can specify precise configurations.
Use technical terminology correctly and be specific about:
- Security settings (encryption, access control)
- Resource policies (least privilege)
- Networking (addressing, security groups)
- Naming conventions

Provide complete, detailed requirements. Don't ask basic questions.
If the Runner asks something you already specified, point that out.`,
		ExpectedBehavior: "Runner should implement exactly as specified with minimal questions",
	}

)

// customPersonas holds user-registered personas.
var (
	customPersonas = make(map[string]Persona)
	customMu       sync.RWMutex
)

// BuiltIn returns the three built-in personas.
func BuiltIn() []Persona {
	return []Persona{Beginner, Intermediate, Expert}
}

// All returns all personas (built-in and custom).
func All() []Persona {
	result := BuiltIn()
	customMu.RLock()
	defer customMu.RUnlock()
	for _, p := range customPersonas {
		result = append(result, p)
	}
	return result
}

// Get returns a persona by name, or an error if not found.
// Checks built-in personas first, then custom personas.
func Get(name string) (Persona, error) {
	name = strings.ToLower(name)

	// Check built-in personas
	for _, p := range BuiltIn() {
		if p.Name == name {
			return p, nil
		}
	}

	// Check custom personas
	customMu.RLock()
	defer customMu.RUnlock()
	if p, ok := customPersonas[name]; ok {
		return p, nil
	}

	return Persona{}, fmt.Errorf("unknown persona: %s (built-in: beginner, intermediate, expert; custom: %s)", name, customNames())
}

// Register adds a custom persona. Returns error if name conflicts with built-in.
func Register(p Persona) error {
	name := strings.ToLower(p.Name)

	// Check for conflict with built-in
	for _, builtin := range BuiltIn() {
		if builtin.Name == name {
			return fmt.Errorf("cannot override built-in persona: %s", name)
		}
	}

	customMu.Lock()
	defer customMu.Unlock()
	customPersonas[name] = p
	return nil
}

// Unregister removes a custom persona by name.
func Unregister(name string) {
	customMu.Lock()
	defer customMu.Unlock()
	delete(customPersonas, strings.ToLower(name))
}

// ClearCustom removes all custom personas.
func ClearCustom() {
	customMu.Lock()
	defer customMu.Unlock()
	customPersonas = make(map[string]Persona)
}

// Names returns the names of all available personas (built-in and custom).
func Names() []string {
	personas := All()
	names := make([]string, len(personas))
	for i, p := range personas {
		names[i] = p.Name
	}
	return names
}

// BuiltInNames returns just the built-in persona names.
func BuiltInNames() []string {
	return []string{"beginner", "intermediate", "expert"}
}

// customNames returns comma-separated custom persona names (for error messages).
func customNames() string {
	customMu.RLock()
	defer customMu.RUnlock()
	if len(customPersonas) == 0 {
		return "none"
	}
	names := make([]string, 0, len(customPersonas))
	for name := range customPersonas {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}
