package validator

import (
	"fmt"
	"strings"
)

// FormatReport formats a validation report as a human-readable string.
func FormatReport(report *ValidationReport) string {
	var sb strings.Builder

	sb.WriteString("Scenario Validation Report\n")
	sb.WriteString("==========================\n\n")

	// Resource Counts
	if len(report.ResourceCounts) > 0 {
		sb.WriteString("Resource Counts:\n")
		for domain, result := range report.ResourceCounts {
			sb.WriteString(fmt.Sprintf("  %s:\n", domain))
			icon := "✓"
			if !result.Passed {
				icon = "✗"
			}
			constraint := fmt.Sprintf("min: %d", result.Min)
			if result.Max > 0 {
				constraint += fmt.Sprintf(", max: %d", result.Max)
			} else {
				constraint += ", max: -"
			}
			sb.WriteString(fmt.Sprintf("    %s Found %d %s (%s)\n", icon, result.Found, result.ResourceType, constraint))
			if result.Error != "" {
				sb.WriteString(fmt.Sprintf("    ⚠ %s\n", result.Error))
			}
			if len(result.Files) > 0 {
				sb.WriteString("    Files:\n")
				for _, f := range result.Files {
					sb.WriteString(fmt.Sprintf("      - %s\n", f))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Cross-Domain References
	if len(report.CrossDomainRefs) > 0 {
		sb.WriteString("Cross-Domain References:\n")
		for _, result := range report.CrossDomainRefs {
			sb.WriteString(fmt.Sprintf("  %s → %s:\n", result.From, result.To))
			for _, ref := range result.FoundRefs {
				locations := result.Locations[ref]
				if len(locations) > 0 {
					sb.WriteString(fmt.Sprintf("    ✓ %s found in %s\n", ref, strings.Join(locations, ", ")))
				} else {
					sb.WriteString(fmt.Sprintf("    ✓ %s found\n", ref))
				}
			}
			for _, ref := range result.MissingRefs {
				sb.WriteString(fmt.Sprintf("    ✗ %s NOT FOUND\n", ref))
			}
		}
		sb.WriteString("\n")
	}

	// File Comparisons
	if len(report.FileComparisons) > 0 {
		sb.WriteString("Expected File Comparison:\n")

		// Group by directory
		byDir := make(map[string][]FileComparisonResult)
		for _, result := range report.FileComparisons {
			dir := "."
			if idx := strings.Index(result.ExpectedFile, "/"); idx > 0 {
				dir = result.ExpectedFile[:idx]
			}
			byDir[dir] = append(byDir[dir], result)
		}

		for dir, results := range byDir {
			if dir != "." {
				sb.WriteString(fmt.Sprintf("  %s/:\n", dir))
			}
			for _, result := range results {
				icon := "✓"
				if !result.Passed {
					icon = "✗"
				}
				if result.Missing {
					icon = "✗"
				}

				name := result.ExpectedFile
				if dir != "." && strings.HasPrefix(name, dir+"/") {
					name = name[len(dir)+1:]
				}

				if result.Missing {
					sb.WriteString(fmt.Sprintf("    %s %s - MISSING\n", icon, name))
				} else if result.Passed {
					sb.WriteString(fmt.Sprintf("    %s %s - structure matches\n", icon, name))
				} else {
					sb.WriteString(fmt.Sprintf("    %s %s - differences found\n", icon, name))
					// Show first few differences
					maxDiffs := 3
					for i, diff := range result.Differences {
						if i >= maxDiffs {
							sb.WriteString(fmt.Sprintf("      ... and %d more\n", len(result.Differences)-maxDiffs))
							break
						}
						// Skip "extra key (allowed)" messages in short output
						if strings.Contains(diff, "(allowed)") {
							continue
						}
						sb.WriteString(fmt.Sprintf("      - %s\n", diff))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	// Errors
	if len(report.Errors) > 0 {
		sb.WriteString("Errors:\n")
		for _, err := range report.Errors {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", err))
		}
		sb.WriteString("\n")
	}

	// Score and summary
	sb.WriteString(fmt.Sprintf("Score: %d/12", report.Score))
	switch {
	case report.Score >= 11:
		sb.WriteString(" (Excellent)\n")
	case report.Score >= 9:
		sb.WriteString(" (Good)\n")
	case report.Score >= 7:
		sb.WriteString(" (Acceptable)\n")
	default:
		sb.WriteString(" (Needs Improvement)\n")
	}

	sb.WriteString("\n")
	if report.Passed {
		sb.WriteString("PASSED\n")
	} else {
		sb.WriteString("FAILED\n")
	}

	return sb.String()
}

// FormatReportMarkdown formats a validation report as markdown.
func FormatReportMarkdown(report *ValidationReport) string {
	var sb strings.Builder

	sb.WriteString("# Scenario Validation Report\n\n")

	// Summary
	status := "✅ PASSED"
	if !report.Passed {
		status = "❌ FAILED"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s  \n", status))
	sb.WriteString(fmt.Sprintf("**Score:** %d/12\n\n", report.Score))

	// Resource Counts
	if len(report.ResourceCounts) > 0 {
		sb.WriteString("## Resource Counts\n\n")
		sb.WriteString("| Domain | Type | Found | Constraint | Status |\n")
		sb.WriteString("|--------|------|-------|------------|--------|\n")
		for domain, result := range report.ResourceCounts {
			constraint := fmt.Sprintf("min: %d", result.Min)
			if result.Max > 0 {
				constraint += fmt.Sprintf(", max: %d", result.Max)
			}
			status := "✅"
			if !result.Passed {
				status = "❌"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n",
				domain, result.ResourceType, result.Found, constraint, status))
		}
		sb.WriteString("\n")
	}

	// Cross-Domain References
	if len(report.CrossDomainRefs) > 0 {
		sb.WriteString("## Cross-Domain References\n\n")
		for _, result := range report.CrossDomainRefs {
			sb.WriteString(fmt.Sprintf("### %s → %s\n\n", result.From, result.To))
			if len(result.FoundRefs) > 0 {
				sb.WriteString("**Found:**\n")
				for _, ref := range result.FoundRefs {
					locations := result.Locations[ref]
					if len(locations) > 0 {
						sb.WriteString(fmt.Sprintf("- ✅ `%s` in %s\n", ref, strings.Join(locations, ", ")))
					} else {
						sb.WriteString(fmt.Sprintf("- ✅ `%s`\n", ref))
					}
				}
				sb.WriteString("\n")
			}
			if len(result.MissingRefs) > 0 {
				sb.WriteString("**Missing:**\n")
				for _, ref := range result.MissingRefs {
					sb.WriteString(fmt.Sprintf("- ❌ `%s`\n", ref))
				}
				sb.WriteString("\n")
			}
		}
	}

	// File Comparisons
	if len(report.FileComparisons) > 0 {
		sb.WriteString("## Expected File Comparison\n\n")
		sb.WriteString("| Expected | Generated | Status | Notes |\n")
		sb.WriteString("|----------|-----------|--------|-------|\n")
		for _, result := range report.FileComparisons {
			status := "✅"
			notes := "matches"
			if result.Missing {
				status = "❌"
				notes = "missing"
			} else if !result.Passed {
				status = "⚠️"
				notes = fmt.Sprintf("%d differences", len(result.Differences))
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				result.ExpectedFile, result.GeneratedFile, status, notes))
		}
		sb.WriteString("\n")
	}

	// Errors
	if len(report.Errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, err := range report.Errors {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", err))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
