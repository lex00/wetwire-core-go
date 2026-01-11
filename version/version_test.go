package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	v := Version()
	assert.NotEmpty(t, v, "Version should not be empty")
}

func TestVersionReturnsDevForLocalBuild(t *testing.T) {
	// When running locally without build info, should return "dev"
	v := Version()
	// In test context without proper build info, expect "dev"
	assert.Equal(t, "dev", v)
}

func TestModulePath(t *testing.T) {
	mp := ModulePath()
	assert.Equal(t, "github.com/lex00/wetwire-core-go", mp)
}
