package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScenarioConfigGetDomain(t *testing.T) {
	config := &ScenarioConfig{
		Domains: []DomainSpec{
			{Name: "domain-a", CLI: "mock-cli-a"},
			{Name: "domain-b", CLI: "mock-cli-b"},
		},
	}

	t.Run("existing domain", func(t *testing.T) {
		d := config.GetDomain("domain-a")
		assert.NotNil(t, d)
		assert.Equal(t, "domain-a", d.Name)
		assert.Equal(t, "mock-cli-a", d.CLI)
	})

	t.Run("non-existing domain", func(t *testing.T) {
		d := config.GetDomain("unknown")
		assert.Nil(t, d)
	})
}

func TestScenarioConfigDomainNames(t *testing.T) {
	config := &ScenarioConfig{
		Domains: []DomainSpec{
			{Name: "domain-a"},
			{Name: "domain-b"},
			{Name: "domain-c"},
		},
	}

	names := config.DomainNames()
	assert.Equal(t, []string{"domain-a", "domain-b", "domain-c"}, names)
}

func TestDomainSpec(t *testing.T) {
	domain := DomainSpec{
		Name: "domain-a",
		CLI:  "mock-cli-a",
		MCPTools: map[string]string{
			"lint":  "wetwire_lint",
			"build": "wetwire_build",
		},
		DependsOn: []string{"resource-1"},
		Outputs:   []string{"templates/*.json"},
	}

	assert.Equal(t, "domain-a", domain.Name)
	assert.Equal(t, "mock-cli-a", domain.CLI)
	assert.Equal(t, "wetwire_lint", domain.MCPTools["lint"])
	assert.Equal(t, []string{"resource-1"}, domain.DependsOn)
	assert.Equal(t, []string{"templates/*.json"}, domain.Outputs)
}

func TestCrossDomainSpec(t *testing.T) {
	cd := CrossDomainSpec{
		From: "domain-a",
		To:   "domain-b",
		Type: "artifact_reference",
		Validation: CrossDomainValidation{
			RequiredRefs: []string{"${domain-a.resource-1.outputs.resource_id}"},
		},
	}

	assert.Equal(t, "domain-a", cd.From)
	assert.Equal(t, "domain-b", cd.To)
	assert.Equal(t, "artifact_reference", cd.Type)
	assert.Contains(t, cd.Validation.RequiredRefs, "${domain-a.resource-1.outputs.resource_id}")
}

func TestValidationRules(t *testing.T) {
	rules := ValidationRules{
		Stacks:    &CountConstraint{Min: 3, Max: 10},
		Pipelines: &CountConstraint{Min: 1},
	}

	assert.Equal(t, 3, rules.Stacks.Min)
	assert.Equal(t, 10, rules.Stacks.Max)
	assert.Equal(t, 1, rules.Pipelines.Min)
	assert.Equal(t, 0, rules.Pipelines.Max)
}
