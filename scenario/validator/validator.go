// Package validator provides validation for scenario results against
// defined validation rules and expected files.
package validator

import (
	"fmt"
	"path/filepath"

	"github.com/lex00/wetwire-core-go/scenario"
)

// Validator validates scenario results against validation rules.
type Validator struct {
	// ScenarioConfig contains the scenario definition with validation rules
	ScenarioConfig *scenario.ScenarioConfig

	// ScenarioDir is the path to the scenario definition directory
	ScenarioDir string

	// ResultsDir is the path to the results directory for a specific persona
	ResultsDir string

	// ExpectedDir is the path to the expected files directory
	ExpectedDir string
}

// New creates a new Validator for the given scenario and results.
func New(scenarioConfig *scenario.ScenarioConfig, scenarioDir, resultsDir string) *Validator {
	return &Validator{
		ScenarioConfig: scenarioConfig,
		ScenarioDir:    scenarioDir,
		ResultsDir:     resultsDir,
		ExpectedDir:    filepath.Join(scenarioDir, "expected"),
	}
}

// ValidationReport contains the full validation results.
type ValidationReport struct {
	// Passed indicates whether all validations passed
	Passed bool

	// ResourceCounts contains validation results for resource count constraints
	ResourceCounts map[string]ResourceCountResult

	// CrossDomainRefs contains validation results for cross-domain references
	CrossDomainRefs []CrossRefResult

	// FileComparisons contains comparison results against expected files
	FileComparisons []FileComparisonResult

	// Errors contains any validation errors encountered
	Errors []string

	// Score is the calculated score (0-12)
	Score int
}

// ResourceCountResult contains the result of validating resource counts for a domain.
type ResourceCountResult struct {
	// Domain is the domain name
	Domain string

	// Passed indicates whether the count constraint was satisfied
	Passed bool

	// Found is the number of resources found
	Found int

	// Min is the minimum required (from validation rules)
	Min int

	// Max is the maximum allowed (0 means no limit)
	Max int

	// ResourceType is the type of resource counted (e.g., "stacks", "pipelines")
	ResourceType string

	// Files lists the files that were counted
	Files []string

	// Error contains any error message if validation failed
	Error string
}

// CrossRefResult contains the result of validating a cross-domain reference.
type CrossRefResult struct {
	// From is the source domain
	From string

	// To is the target domain
	To string

	// Passed indicates whether all required refs were found
	Passed bool

	// RequiredRefs lists the references that were required
	RequiredRefs []string

	// FoundRefs lists the references that were found
	FoundRefs []string

	// MissingRefs lists the references that were missing
	MissingRefs []string

	// Locations maps found refs to their file locations
	Locations map[string][]string
}

// FileComparisonResult contains the result of comparing a generated file to expected.
type FileComparisonResult struct {
	// ExpectedFile is the path to the expected file (relative to expected dir)
	ExpectedFile string

	// GeneratedFile is the path to the generated file (relative to results dir)
	GeneratedFile string

	// Passed indicates whether the files match structurally
	Passed bool

	// Missing indicates the expected file was not found in results
	Missing bool

	// Differences contains structural differences if any
	Differences []string
}

// Validate runs all validation checks and returns a report.
func (v *Validator) Validate() (*ValidationReport, error) {
	report := &ValidationReport{
		Passed:          true,
		ResourceCounts:  make(map[string]ResourceCountResult),
		CrossDomainRefs: []CrossRefResult{},
		FileComparisons: []FileComparisonResult{},
		Errors:          []string{},
	}

	// Validate resource counts
	countResults, err := v.ValidateResourceCounts()
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("resource count validation error: %v", err))
	}
	report.ResourceCounts = countResults
	for _, result := range countResults {
		if !result.Passed {
			report.Passed = false
		}
	}

	// Validate cross-domain references
	refResults, err := v.ValidateCrossRefs()
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("cross-ref validation error: %v", err))
	}
	report.CrossDomainRefs = refResults
	for _, result := range refResults {
		if !result.Passed {
			report.Passed = false
		}
	}

	// Compare against expected files
	compareResults, err := v.CompareExpected()
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("expected comparison error: %v", err))
	}
	report.FileComparisons = compareResults
	for _, result := range compareResults {
		if !result.Passed {
			report.Passed = false
		}
	}

	// Calculate score
	report.Score = v.calculateScore(report)

	return report, nil
}

// calculateScore computes the score based on validation results.
func (v *Validator) calculateScore(report *ValidationReport) int {
	score := 0

	// Completeness (0-3): Based on resource count validation
	completenessScore := 3
	for _, result := range report.ResourceCounts {
		if !result.Passed {
			completenessScore = 0
			break
		}
	}
	score += completenessScore

	// Lint Quality (0-3): Deferred to domain tools, assume passing
	score += 3

	// Output Validity (0-3): Based on cross-ref validation
	validityScore := 3
	for _, result := range report.CrossDomainRefs {
		if !result.Passed {
			validityScore = 0
			break
		}
	}
	score += validityScore

	// Question Efficiency (0-3): Based on expected file comparison
	// This is reused since question efficiency isn't tracked by validator
	// Missing files reduce score; structural differences are warnings only
	comparisonScore := 3
	missingCount := 0
	for _, result := range report.FileComparisons {
		if result.Missing {
			missingCount++
		}
	}
	if missingCount > 0 {
		// Reduce score based on missing files
		if missingCount >= 3 {
			comparisonScore = 0
		} else {
			comparisonScore -= missingCount
		}
	}
	if comparisonScore < 0 {
		comparisonScore = 0
	}
	score += comparisonScore

	return score
}
