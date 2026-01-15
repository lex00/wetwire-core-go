package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScenarioConfigGetDomain(t *testing.T) {
	config := &ScenarioConfig{
		Domains: []DomainSpec{
			{Name: "aws", CLI: "wetwire-aws"},
			{Name: "gitlab", CLI: "wetwire-gitlab"},
		},
	}

	t.Run("existing domain", func(t *testing.T) {
		d := config.GetDomain("aws")
		assert.NotNil(t, d)
		assert.Equal(t, "aws", d.Name)
		assert.Equal(t, "wetwire-aws", d.CLI)
	})

	t.Run("non-existing domain", func(t *testing.T) {
		d := config.GetDomain("unknown")
		assert.Nil(t, d)
	})
}

func TestScenarioConfigDomainNames(t *testing.T) {
	config := &ScenarioConfig{
		Domains: []DomainSpec{
			{Name: "aws"},
			{Name: "gitlab"},
			{Name: "k8s"},
		},
	}

	names := config.DomainNames()
	assert.Equal(t, []string{"aws", "gitlab", "k8s"}, names)
}

func TestDomainSpec(t *testing.T) {
	domain := DomainSpec{
		Name: "aws",
		CLI:  "wetwire-aws",
		MCPTools: map[string]string{
			"lint":  "wetwire_lint",
			"build": "wetwire_build",
		},
		DependsOn: []string{"vpc"},
		Outputs:   []string{"cfn-templates/*.json"},
	}

	assert.Equal(t, "aws", domain.Name)
	assert.Equal(t, "wetwire-aws", domain.CLI)
	assert.Equal(t, "wetwire_lint", domain.MCPTools["lint"])
	assert.Equal(t, []string{"vpc"}, domain.DependsOn)
	assert.Equal(t, []string{"cfn-templates/*.json"}, domain.Outputs)
}

func TestCrossDomainSpec(t *testing.T) {
	cd := CrossDomainSpec{
		From: "aws",
		To:   "gitlab",
		Type: "artifact_reference",
		Validation: CrossDomainValidation{
			RequiredRefs: []string{"${aws.vpc.outputs.vpc_id}"},
		},
	}

	assert.Equal(t, "aws", cd.From)
	assert.Equal(t, "gitlab", cd.To)
	assert.Equal(t, "artifact_reference", cd.Type)
	assert.Contains(t, cd.Validation.RequiredRefs, "${aws.vpc.outputs.vpc_id}")
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
