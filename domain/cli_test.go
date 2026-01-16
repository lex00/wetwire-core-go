package domain_test

import (
	"testing"

	"github.com/lex00/wetwire-core-go/domain"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing CLI generation
type testDomain struct{}

func (t *testDomain) Name() string                    { return "test" }
func (t *testDomain) Version() string                 { return "1.0.0" }
func (t *testDomain) Builder() domain.Builder         { return &mockBuilder{} }
func (t *testDomain) Linter() domain.Linter           { return &mockLinter{} }
func (t *testDomain) Initializer() domain.Initializer { return &mockInitializer{} }
func (t *testDomain) Validator() domain.Validator     { return &mockValidator{} }

type testImporterDomain struct {
	testDomain
}

func (t *testImporterDomain) Importer() domain.Importer { return &mockImporter{} }

type testListerDomain struct {
	testDomain
}

func (t *testListerDomain) Lister() domain.Lister { return &mockLister{} }

type testGrapherDomain struct {
	testDomain
}

func (t *testGrapherDomain) Grapher() domain.Grapher { return &mockGrapher{} }

type testFullDomain struct {
	testDomain
}

func (t *testFullDomain) Importer() domain.Importer { return &mockImporter{} }
func (t *testFullDomain) Lister() domain.Lister     { return &mockLister{} }
func (t *testFullDomain) Grapher() domain.Grapher   { return &mockGrapher{} }

// TestRunReturnsValidCommand tests that Run() returns a valid cobra.Command
func TestRunReturnsValidCommand(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	require.NotNil(t, cmd, "Run() should return a non-nil command")
	assert.IsType(t, &cobra.Command{}, cmd, "Run() should return a *cobra.Command")
	assert.Equal(t, "wetwire-test", cmd.Use, "Command should use domain name")
	assert.Equal(t, "1.0.0", cmd.Version, "Command should use domain version")
}

// TestRequiredCommandsGenerated tests that all 4 required commands are generated
func TestRequiredCommandsGenerated(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	// Check that all required subcommands exist
	requiredCommands := []string{"build", "lint", "init", "validate"}

	for _, cmdName := range requiredCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == cmdName {
				found = true
				break
			}
		}
		assert.True(t, found, "Command '%s' should be generated", cmdName)
	}

	// Should have exactly 4 commands for basic domain
	assert.Len(t, cmd.Commands(), 4, "Basic domain should have exactly 4 commands")
}

// TestOptionalCommandsOnlyIfImplemented tests that optional commands are only added
// if the domain implements the corresponding interfaces
func TestOptionalCommandsOnlyIfImplemented(t *testing.T) {
	tests := []struct {
		name            string
		domain          domain.Domain
		expectedCmds    []string
		notExpectedCmds []string
	}{
		{
			name:            "Basic domain without optional interfaces",
			domain:          &testDomain{},
			expectedCmds:    []string{"build", "lint", "init", "validate"},
			notExpectedCmds: []string{"import", "list", "graph"},
		},
		{
			name:            "Domain with Importer",
			domain:          &testImporterDomain{},
			expectedCmds:    []string{"build", "lint", "init", "validate", "import"},
			notExpectedCmds: []string{"list", "graph"},
		},
		{
			name:            "Domain with Lister",
			domain:          &testListerDomain{},
			expectedCmds:    []string{"build", "lint", "init", "validate", "list"},
			notExpectedCmds: []string{"import", "graph"},
		},
		{
			name:            "Domain with Grapher",
			domain:          &testGrapherDomain{},
			expectedCmds:    []string{"build", "lint", "init", "validate", "graph"},
			notExpectedCmds: []string{"import", "list"},
		},
		{
			name:            "Domain with all optional interfaces",
			domain:          &testFullDomain{},
			expectedCmds:    []string{"build", "lint", "init", "validate", "import", "list", "graph"},
			notExpectedCmds: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := domain.Run(tt.domain)

			// Check expected commands exist
			for _, expectedCmd := range tt.expectedCmds {
				found := false
				for _, subCmd := range cmd.Commands() {
					if subCmd.Name() == expectedCmd {
						found = true
						break
					}
				}
				assert.True(t, found, "Command '%s' should be generated", expectedCmd)
			}

			// Check not-expected commands don't exist
			for _, notExpectedCmd := range tt.notExpectedCmds {
				found := false
				for _, subCmd := range cmd.Commands() {
					if subCmd.Name() == notExpectedCmd {
						found = true
						break
					}
				}
				assert.False(t, found, "Command '%s' should not be generated", notExpectedCmd)
			}
		})
	}
}

// TestPersistentFormatFlag tests that persistent --format flag is on root command
func TestPersistentFormatFlag(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	flag := cmd.PersistentFlags().Lookup("format")
	require.NotNil(t, flag, "Root command should have persistent --format flag")
	assert.Equal(t, "f", flag.Shorthand, "Format flag should have -f shorthand")
	assert.Equal(t, "text", flag.DefValue, "Format flag should default to 'text'")
}

// TestPersistentVerboseFlag tests that persistent --verbose flag is inherited
func TestPersistentVerboseFlag(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	// Check root has verbose flag
	flag := cmd.PersistentFlags().Lookup("verbose")
	require.NotNil(t, flag, "Root command should have persistent --verbose flag")
	assert.Equal(t, "v", flag.Shorthand, "Verbose flag should have -v shorthand")
	assert.Equal(t, "false", flag.DefValue, "Verbose flag should default to false")

	// Check that subcommands inherit the flag
	for _, subCmd := range cmd.Commands() {
		inheritedFlag := subCmd.InheritedFlags().Lookup("verbose")
		assert.NotNil(t, inheritedFlag, "Subcommand '%s' should inherit --verbose flag", subCmd.Name())
	}
}

// TestCommandShortDescriptions tests that commands have short descriptions
func TestCommandShortDescriptions(t *testing.T) {
	d := &testFullDomain{}
	cmd := domain.Run(d)

	for _, subCmd := range cmd.Commands() {
		assert.NotEmpty(t, subCmd.Short, "Command '%s' should have a short description", subCmd.Name())
	}
}

// TestBuildCommandFlags tests build command has expected flags
func TestBuildCommandFlags(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	buildCmd := findCommand(cmd, "build")
	require.NotNil(t, buildCmd, "Build command should exist")

	// Check for --type flag
	typeFlag := buildCmd.Flags().Lookup("type")
	assert.NotNil(t, typeFlag, "Build command should have --type flag")
}

// TestLintCommandFlags tests lint command has expected flags
func TestLintCommandFlags(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	lintCmd := findCommand(cmd, "lint")
	require.NotNil(t, lintCmd, "Lint command should exist")

	// Lint inherits format from root, so check inherited flags
	formatFlag := lintCmd.InheritedFlags().Lookup("format")
	assert.NotNil(t, formatFlag, "Lint command should inherit --format flag")
}

// TestInitCommandFlags tests init command has expected flags
func TestInitCommandFlags(t *testing.T) {
	d := &testDomain{}
	cmd := domain.Run(d)

	initCmd := findCommand(cmd, "init")
	require.NotNil(t, initCmd, "Init command should exist")

	// Check for --name flag
	nameFlag := initCmd.Flags().Lookup("name")
	assert.NotNil(t, nameFlag, "Init command should have --name flag")

	// Check for --path flag
	pathFlag := initCmd.Flags().Lookup("path")
	assert.NotNil(t, pathFlag, "Init command should have --path flag")
}

// TestListCommandFlags tests list command has expected flags (if domain supports it)
func TestListCommandFlags(t *testing.T) {
	d := &testListerDomain{}
	cmd := domain.Run(d)

	listCmd := findCommand(cmd, "list")
	require.NotNil(t, listCmd, "List command should exist for ListerDomain")

	// Check for --type flag
	typeFlag := listCmd.Flags().Lookup("type")
	assert.NotNil(t, typeFlag, "List command should have --type flag")
}

// TestGraphCommandFlags tests graph command has expected flags (if domain supports it)
func TestGraphCommandFlags(t *testing.T) {
	d := &testGrapherDomain{}
	cmd := domain.Run(d)

	graphCmd := findCommand(cmd, "graph")
	require.NotNil(t, graphCmd, "Graph command should exist for GrapherDomain")

	// Graph inherits format from root
	formatFlag := graphCmd.InheritedFlags().Lookup("format")
	assert.NotNil(t, formatFlag, "Graph command should inherit --format flag")
}

// TestImportCommandFlags tests import command has expected flags (if domain supports it)
func TestImportCommandFlags(t *testing.T) {
	d := &testImporterDomain{}
	cmd := domain.Run(d)

	importCmd := findCommand(cmd, "import")
	require.NotNil(t, importCmd, "Import command should exist for ImporterDomain")

	// Check for --target flag
	targetFlag := importCmd.Flags().Lookup("target")
	assert.NotNil(t, targetFlag, "Import command should have --target flag")
}

// Helper function to find a command by name
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
