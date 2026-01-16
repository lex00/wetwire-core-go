package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	t.Run("with field", func(t *testing.T) {
		err := ValidationError{Field: "name", Message: "is required"}
		assert.Equal(t, "name: is required", err.Error())
	})

	t.Run("without field", func(t *testing.T) {
		err := ValidationError{Message: "something went wrong"}
		assert.Equal(t, "something went wrong", err.Error())
	})
}

func TestValidationResult(t *testing.T) {
	t.Run("empty result is valid", func(t *testing.T) {
		result := &ValidationResult{}
		assert.True(t, result.IsValid())
		assert.Equal(t, "", result.Error())
	})

	t.Run("result with errors is invalid", func(t *testing.T) {
		result := &ValidationResult{}
		result.AddError("name", "is required")
		result.AddError("domains", "must have at least one")

		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "name: is required")
		assert.Contains(t, result.Error(), "domains: must have at least one")
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
			},
		}

		result := Validate(config)
		assert.True(t, result.IsValid())
	})

	t.Run("missing name", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "name")
	})

	t.Run("no domains", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "domains")
	})

	t.Run("domain missing name", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{CLI: "mock-cli-a"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "domain name is required")
	})

	t.Run("domain missing CLI", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "domain CLI is required")
	})

	t.Run("duplicate domain names", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
				{Name: "domain-a", CLI: "mock-cli-a-2"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "duplicate domain name")
	})

	t.Run("unknown dependency", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a", DependsOn: []string{"unknown"}},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "unknown dependency")
	})

	t.Run("self dependency", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a", DependsOn: []string{"domain-a"}},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "cannot depend on itself")
	})

	t.Run("cross-domain missing from", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
				{Name: "domain-b", CLI: "mock-cli-b"},
			},
			CrossDomain: []CrossDomainSpec{
				{To: "domain-b", Type: "artifact"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "source domain is required")
	})

	t.Run("cross-domain unknown domain", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
			},
			CrossDomain: []CrossDomainSpec{
				{From: "domain-a", To: "domain-b", Type: "artifact"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "unknown domain: domain-b")
	})

	t.Run("cross-domain missing type", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a", CLI: "mock-cli-a"},
				{Name: "domain-b", CLI: "mock-cli-b"},
			},
			CrossDomain: []CrossDomainSpec{
				{From: "domain-a", To: "domain-b"},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "relationship type is required")
	})

	t.Run("circular dependencies", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "a", CLI: "cli-a", DependsOn: []string{"b"}},
				{Name: "b", CLI: "cli-b", DependsOn: []string{"a"}},
			},
		}

		result := Validate(config)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error(), "circular")
	})
}

func TestValidateRequired(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
			Domains: []DomainSpec{
				{Name: "domain-a"},
			},
		}

		err := ValidateRequired(config)
		assert.NoError(t, err)
	})

	t.Run("missing name", func(t *testing.T) {
		config := &ScenarioConfig{
			Domains: []DomainSpec{{Name: "domain-a"}},
		}

		err := ValidateRequired(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("no domains", func(t *testing.T) {
		config := &ScenarioConfig{
			Name: "test",
		}

		err := ValidateRequired(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "domains")
	})
}
