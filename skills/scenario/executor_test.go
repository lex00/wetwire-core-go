package scenario

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lex00/wetwire-core-go/scenario"
)

func TestGenerateInstructions(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name:        "test_infra",
		Description: "Test infrastructure",
		Domains: []scenario.DomainSpec{
			{
				Name: "domain-a",
				CLI:  "mock-cli-a",
				MCPTools: map[string]string{
					"lint":  "wetwire_lint",
					"build": "wetwire_build",
				},
				Outputs: []string{"templates/*.json"},
			},
			{
				Name:      "domain-b",
				CLI:       "mock-cli-b",
				DependsOn: []string{"domain-a"},
				MCPTools: map[string]string{
					"lint":  "wetwire_lint",
					"build": "wetwire_build",
				},
			},
		},
		CrossDomain: []scenario.CrossDomainSpec{
			{
				From: "domain-a",
				To:   "domain-b",
				Type: "artifact_reference",
			},
		},
		Validation: map[string]scenario.ValidationRules{
			"domain-a": {
				Stacks: &scenario.CountConstraint{Min: 2},
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateInstructions(&buf, config, "Generate infrastructure code.", nil)
	require.NoError(t, err)

	output := buf.String()

	// Should contain scenario header
	assert.Contains(t, output, "test_infra")
	assert.Contains(t, output, "Test infrastructure")

	// Should contain prompt
	assert.Contains(t, output, "Generate infrastructure code")

	// Should contain domain steps in correct order
	assert.Contains(t, output, "domain-a")
	assert.Contains(t, output, "domain-b")

	// domain-a should come before domain-b
	domainAIdx := bytes.Index([]byte(output), []byte("Step 1"))
	domainBIdx := bytes.Index([]byte(output), []byte("Step 2"))
	assert.Less(t, domainAIdx, domainBIdx)

	// Should contain MCP tool instructions
	assert.Contains(t, output, "wetwire_lint")
	assert.Contains(t, output, "wetwire_build")

	// Should contain validation criteria
	assert.Contains(t, output, "Stacks")
	assert.Contains(t, output, "2")

	// Should contain cross-domain step
	assert.Contains(t, output, "cross-domain")
}

func TestGenerateInstructionsSingleDomain(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name: "simple_scenario",
		Domains: []scenario.DomainSpec{
			{
				Name: "domain-a",
				CLI:  "mock-cli-a",
				MCPTools: map[string]string{
					"lint": "wetwire_lint",
				},
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateInstructions(&buf, config, "", nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "simple_scenario")
	assert.Contains(t, output, "domain-a")
	assert.Contains(t, output, "wetwire_lint")

	// Should not contain cross-domain step
	assert.NotContains(t, output, "cross-domain")
}

func TestGenerateInstructionsWithPromptVariants(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name: "multi_variant",
		Prompts: &scenario.PromptConfig{
			Default:  "default.md",
			Variants: map[string]string{"minimal": "minimal.md"},
		},
		Domains: []scenario.DomainSpec{
			{Name: "domain-a", CLI: "mock-cli-a"},
		},
	}

	variants := map[string]string{
		"default": "Full infrastructure prompt",
		"minimal": "Minimal infrastructure prompt",
	}

	var buf bytes.Buffer
	err := GenerateInstructions(&buf, config, variants["minimal"], nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Minimal infrastructure prompt")
}

func TestGenerateInstructionsMultipleDependencies(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name: "complex",
		Domains: []scenario.DomainSpec{
			{Name: "resource-1", CLI: "mock-cli-a"},
			{Name: "resource-2", CLI: "mock-cli-a", DependsOn: []string{"resource-1"}},
			{Name: "app", CLI: "mock-cli-c", DependsOn: []string{"resource-2"}},
		},
	}

	var buf bytes.Buffer
	err := GenerateInstructions(&buf, config, "", nil)
	require.NoError(t, err)

	output := buf.String()

	// Should have correct order: resource-1 -> resource-2 -> app
	res1Idx := bytes.Index([]byte(output), []byte("resource-1"))
	res2Idx := bytes.Index([]byte(output), []byte("resource-2"))
	appIdx := bytes.Index([]byte(output), []byte("app"))

	assert.Less(t, res1Idx, res2Idx, "resource-1 should come before resource-2")
	assert.Less(t, res2Idx, appIdx, "resource-2 should come before app")
}

func TestGenerateInstructionsValidationOutput(t *testing.T) {
	config := &scenario.ScenarioConfig{
		Name: "validated",
		Domains: []scenario.DomainSpec{
			{Name: "domain-a", CLI: "mock-cli-a"},
		},
		Validation: map[string]scenario.ValidationRules{
			"domain-a": {
				Stacks:    &scenario.CountConstraint{Min: 3, Max: 10},
				Resources: &scenario.CountConstraint{Min: 5},
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateInstructions(&buf, config, "", nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Validation")
	assert.Contains(t, output, "Stacks")
	assert.Contains(t, output, "3")
}

func TestGenerateInstructionsEmptyConfig(t *testing.T) {
	config := &scenario.ScenarioConfig{}

	var buf bytes.Buffer
	err := GenerateInstructions(&buf, config, "", nil)

	// Should error on empty/invalid config
	require.Error(t, err)
}

func TestFormatDomainStep(t *testing.T) {
	domain := &scenario.DomainSpec{
		Name: "domain-a",
		CLI:  "mock-cli-a",
		MCPTools: map[string]string{
			"lint":  "wetwire_lint",
			"build": "wetwire_build",
		},
		Outputs: []string{"templates/*.json"},
	}

	var buf bytes.Buffer
	FormatDomainStep(&buf, 1, domain)

	output := buf.String()
	assert.Contains(t, output, "Step 1")
	assert.Contains(t, output, "domain-a")
	assert.Contains(t, output, "mock-cli-a")
	assert.Contains(t, output, "wetwire_lint")
	assert.Contains(t, output, "wetwire_build")
	assert.Contains(t, output, "templates/*.json")
}

func TestFormatValidationCriteria(t *testing.T) {
	validation := map[string]scenario.ValidationRules{
		"domain-a": {
			Stacks: &scenario.CountConstraint{Min: 2, Max: 5},
		},
		"domain-b": {
			Pipelines: &scenario.CountConstraint{Min: 1},
		},
	}

	var buf bytes.Buffer
	FormatValidationCriteria(&buf, validation)

	output := buf.String()
	assert.Contains(t, output, "domain-a")
	assert.Contains(t, output, "Stacks")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "5")
	assert.Contains(t, output, "domain-b")
	assert.Contains(t, output, "Pipelines")
}
