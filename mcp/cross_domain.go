package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lex00/wetwire-core-go/scenario"
)

// CrossDomainValidateSchema is the JSON schema for wetwire_validate_cross_domain tool.
var CrossDomainValidateSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"scenario": map[string]any{
			"type":        "string",
			"description": "Scenario name or path to scenario.yaml",
		},
		"output_dir": map[string]any{
			"type":        "string",
			"description": "Directory where generated code lives",
		},
	},
	"required": []string{"scenario", "output_dir"},
}

// CrossDomainValidationResult is the result of cross-domain validation.
type CrossDomainValidationResult struct {
	Valid            bool                   `json:"valid"`
	DomainsValidated []string               `json:"domains_validated"`
	CrossReferences  []CrossReferenceResult `json:"cross_references"`
	Errors           []string               `json:"errors"`
	Score            int                    `json:"score"`
}

// CrossReferenceResult represents a single cross-reference validation.
type CrossReferenceResult struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Type  string `json:"type"`
	Valid bool   `json:"valid"`
}

// ValidateCrossDomain validates cross-domain references in a scenario.
func ValidateCrossDomain(ctx context.Context, args map[string]any) (string, error) {
	scenarioPath, ok := args["scenario"].(string)
	if !ok || scenarioPath == "" {
		return "", fmt.Errorf("scenario path is required")
	}

	outputDir, ok := args["output_dir"].(string)
	if !ok || outputDir == "" {
		return "", fmt.Errorf("output_dir is required")
	}

	// Load scenario
	config, err := scenario.Load(scenarioPath)
	if err != nil {
		return "", fmt.Errorf("failed to load scenario: %w", err)
	}

	// Initialize result
	result := &CrossDomainValidationResult{
		Valid:            true,
		DomainsValidated: config.DomainNames(),
		CrossReferences:  []CrossReferenceResult{},
		Errors:           []string{},
		Score:            0,
	}

	// If no cross-domain relationships, return early
	if len(config.CrossDomain) == 0 {
		return toJSON(result)
	}

	// Validate each cross-domain relationship
	for _, cd := range config.CrossDomain {
		refs, errs := validateCrossDomainRelationship(config, cd, outputDir)
		result.CrossReferences = append(result.CrossReferences, refs...)
		result.Errors = append(result.Errors, errs...)
	}

	// Calculate validity and score
	result.Valid = len(result.Errors) == 0
	result.Score = calculateCrossDomainScore(result)

	return toJSON(result)
}

// validateCrossDomainRelationship validates a single cross-domain relationship.
func validateCrossDomainRelationship(
	config *scenario.ScenarioConfig,
	cd scenario.CrossDomainSpec,
	outputDir string,
) ([]CrossReferenceResult, []string) {
	var refs []CrossReferenceResult
	var errs []string

	fromDomain := config.GetDomain(cd.From)
	if fromDomain == nil {
		errs = append(errs, fmt.Sprintf("source domain not found: %s", cd.From))
		return refs, errs
	}

	toDomain := config.GetDomain(cd.To)
	if toDomain == nil {
		errs = append(errs, fmt.Sprintf("target domain not found: %s", cd.To))
		return refs, errs
	}

	switch cd.Type {
	case "artifact_reference":
		r, e := validateArtifactReferences(fromDomain, toDomain, cd, outputDir)
		refs = append(refs, r...)
		errs = append(errs, e...)
	case "output_mapping":
		r, e := validateOutputMapping(fromDomain, toDomain, cd, outputDir)
		refs = append(refs, r...)
		errs = append(errs, e...)
	default:
		// Generic validation - just check directories exist
		fromDir := filepath.Join(outputDir, cd.From)
		toDir := filepath.Join(outputDir, cd.To)

		if _, err := os.Stat(fromDir); os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("source domain directory not found: %s", fromDir))
		}
		if _, err := os.Stat(toDir); os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("target domain directory not found: %s", toDir))
		}
	}

	return refs, errs
}

// validateArtifactReferences validates artifact_reference type cross-domain relationships.
func validateArtifactReferences(
	fromDomain *scenario.DomainSpec,
	toDomain *scenario.DomainSpec,
	cd scenario.CrossDomainSpec,
	outputDir string,
) ([]CrossReferenceResult, []string) {
	var refs []CrossReferenceResult
	var errs []string

	fromDir := filepath.Join(outputDir, fromDomain.Name)
	toDir := filepath.Join(outputDir, toDomain.Name)

	// Check for required refs in the target domain's files
	for _, reqRef := range cd.Validation.RequiredRefs {
		ref := CrossReferenceResult{
			From:  fromDomain.Name,
			To:    toDomain.Name,
			Type:  extractRefType(reqRef),
			Valid: false,
		}

		// Check if the reference exists in target domain files
		found, err := findReferenceInDirectory(toDir, reqRef)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error scanning %s: %v", toDir, err))
		} else if found {
			ref.Valid = true
		} else {
			errs = append(errs, fmt.Sprintf("required reference not found in %s: %s", toDomain.Name, reqRef))
		}

		// Check if the source domain has outputs that could satisfy the reference
		if len(fromDomain.Outputs) > 0 {
			hasOutput, _ := checkOutputPatterns(fromDir, fromDomain.Outputs)
			if !hasOutput {
				errs = append(errs, fmt.Sprintf("source domain %s missing output files", fromDomain.Name))
			}
		}

		refs = append(refs, ref)
	}

	return refs, errs
}

// validateOutputMapping validates output_mapping type cross-domain relationships.
func validateOutputMapping(
	fromDomain *scenario.DomainSpec,
	toDomain *scenario.DomainSpec,
	cd scenario.CrossDomainSpec,
	outputDir string,
) ([]CrossReferenceResult, []string) {
	var refs []CrossReferenceResult
	var errs []string

	fromDir := filepath.Join(outputDir, fromDomain.Name)

	// Check that source domain has output files
	if len(fromDomain.Outputs) > 0 {
		hasOutput, _ := checkOutputPatterns(fromDir, fromDomain.Outputs)
		if !hasOutput {
			errs = append(errs, fmt.Sprintf("source domain %s has no output files", fromDomain.Name))
		}
	}

	// For output_mapping, we just verify the relationship structure is valid
	ref := CrossReferenceResult{
		From:  fromDomain.Name,
		To:    toDomain.Name,
		Type:  "output_mapping",
		Valid: len(errs) == 0,
	}
	refs = append(refs, ref)

	return refs, errs
}

// findReferenceInDirectory searches for a reference pattern in files within a directory.
func findReferenceInDirectory(dir string, ref string) (bool, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false, nil
	}

	found := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		// Read file and search for reference
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		if strings.Contains(string(data), ref) {
			found = true
		}

		return nil
	})

	return found, err
}

// checkOutputPatterns checks if any files matching the output patterns exist.
func checkOutputPatterns(dir string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		fullPattern := filepath.Join(dir, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// extractRefType extracts a short type name from a reference string.
func extractRefType(ref string) string {
	// Extract the last part of a reference like "${aws.vpc.outputs.vpc_id}"
	re := regexp.MustCompile(`\.([^.}]+)\}?$`)
	matches := re.FindStringSubmatch(ref)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

// calculateCrossDomainScore calculates a score based on validation results.
func calculateCrossDomainScore(result *CrossDomainValidationResult) int {
	score := 0

	// Base score for having cross-domain validation
	if len(result.CrossReferences) > 0 {
		score += 5
	}

	// Points for each valid reference
	for _, ref := range result.CrossReferences {
		if ref.Valid {
			score += 5
		}
	}

	// Points for each domain validated
	score += len(result.DomainsValidated) * 2

	// Penalty for errors
	score -= len(result.Errors) * 3

	// Minimum score is 0
	if score < 0 {
		score = 0
	}

	return score
}

// toJSON converts a result to JSON string.
func toJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
