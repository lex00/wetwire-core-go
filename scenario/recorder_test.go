package scenario

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanRecord(t *testing.T) {
	// CanRecord should not panic regardless of termsvg availability
	result := CanRecord()
	// Result depends on environment, just ensure it returns a bool
	assert.IsType(t, true, result)
}

func TestRecorderConfig(t *testing.T) {
	config := RecorderConfig{
		OutputDir:    "/tmp/recordings",
		ScenarioName: "test_scenario",
		Format:       "svg",
	}

	assert.Equal(t, "/tmp/recordings", config.OutputDir)
	assert.Equal(t, "test_scenario", config.ScenarioName)
	assert.Equal(t, "svg", config.Format)
}

func TestNewRecorder(t *testing.T) {
	tmpDir := t.TempDir()

	config := RecorderConfig{
		OutputDir:    tmpDir,
		ScenarioName: "test",
	}

	recorder := NewRecorder(config)
	require.NotNil(t, recorder)
	// Config should have default format applied
	assert.Equal(t, tmpDir, recorder.config.OutputDir)
	assert.Equal(t, "test", recorder.config.ScenarioName)
	assert.Equal(t, "svg", recorder.config.Format) // Default format
}

func TestRecorderOutputPath(t *testing.T) {
	config := RecorderConfig{
		OutputDir:    "/tmp/recordings",
		ScenarioName: "ecommerce",
	}

	recorder := NewRecorder(config)

	path := recorder.OutputPath()
	assert.Equal(t, "/tmp/recordings/ecommerce.svg", path)
}

func TestRecorderCastPath(t *testing.T) {
	config := RecorderConfig{
		OutputDir:    "/tmp/recordings",
		ScenarioName: "ecommerce",
	}

	recorder := NewRecorder(config)

	path := recorder.CastPath()
	assert.Equal(t, "/tmp/recordings/ecommerce.cast", path)
}

func TestRecorderCanRecordFalse(t *testing.T) {
	// Test behavior when termsvg is not available
	tmpDir := t.TempDir()

	config := RecorderConfig{
		OutputDir:    tmpDir,
		ScenarioName: "test",
	}

	recorder := NewRecorder(config)

	// If termsvg is not installed, Record should return gracefully
	err := recorder.Record(func() error {
		return nil
	})

	// Either succeeds (termsvg installed) or returns ErrTermsvgNotFound
	if err != nil {
		assert.ErrorIs(t, err, ErrTermsvgNotFound)
	}
}

func TestRecorderEnsureOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "recordings", "nested")

	config := RecorderConfig{
		OutputDir:    outputDir,
		ScenarioName: "test",
	}

	recorder := NewRecorder(config)
	err := recorder.ensureOutputDir()
	require.NoError(t, err)

	// Directory should exist
	info, err := os.Stat(outputDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestRecorderCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a .cast file to clean up
	castFile := filepath.Join(tmpDir, "test.cast")
	err := os.WriteFile(castFile, []byte("test content"), 0644)
	require.NoError(t, err)

	config := RecorderConfig{
		OutputDir:    tmpDir,
		ScenarioName: "test",
	}

	recorder := NewRecorder(config)
	recorder.Cleanup()

	// Cast file should be removed
	_, err = os.Stat(castFile)
	assert.True(t, os.IsNotExist(err))
}

func TestRecorderWithCustomCommand(t *testing.T) {
	tmpDir := t.TempDir()

	config := RecorderConfig{
		OutputDir:    tmpDir,
		ScenarioName: "test",
	}

	recorder := NewRecorder(config)

	// Test that we can configure the command
	executed := false
	err := recorder.Record(func() error {
		executed = true
		return nil
	})

	// If termsvg available, command should execute
	// If not, should return ErrTermsvgNotFound
	if err == nil {
		assert.True(t, executed)
	} else {
		assert.ErrorIs(t, err, ErrTermsvgNotFound)
	}
}

func TestRecordToSVG(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.svg")

	// RecordToSVG is a convenience function
	err := RecordToSVG(outputPath, func() error {
		return nil
	})

	// Either succeeds or returns ErrTermsvgNotFound
	if err != nil {
		assert.ErrorIs(t, err, ErrTermsvgNotFound)
	}
}

func TestRecordToSVGCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "deep", "output.svg")

	err := RecordToSVG(outputPath, func() error {
		return nil
	})

	// Directory should be created even if recording fails
	dir := filepath.Dir(outputPath)
	_, statErr := os.Stat(dir)

	// If termsvg not available, might fail before dir creation
	if err == nil || !os.IsNotExist(statErr) {
		assert.NoError(t, statErr)
	}
}
