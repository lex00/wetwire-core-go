// Package lint provides shared linting infrastructure for wetwire domain packages.
package lint

// Severity indicates the severity level of a lint issue.
type Severity int

const (
	// SeverityError indicates a critical issue that should be fixed.
	SeverityError Severity = iota
	// SeverityWarning indicates a potential problem that should be reviewed.
	SeverityWarning
	// SeverityInfo indicates a suggestion or informational message.
	SeverityInfo
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// Issue represents a single lint issue found during analysis.
type Issue struct {
	// Rule is the unique identifier of the rule that found this issue.
	Rule string
	// Message describes the issue.
	Message string
	// File is the path to the file containing the issue.
	File string
	// Line is the line number (1-based) where the issue was found.
	Line int
	// Column is the column number (1-based) where the issue was found.
	Column int
	// Severity indicates how serious the issue is.
	Severity Severity
	// Suggestion provides a recommended fix for the issue.
	Suggestion string
	// Fixable indicates whether this issue can be automatically fixed.
	Fixable bool
}

// Config controls linting behavior.
type Config struct {
	// DisabledRules is a list of rule IDs to skip.
	DisabledRules []string
	// MinSeverity is the minimum severity level to report.
	// Issues with lower severity will be filtered out.
	MinSeverity Severity
}

// IsRuleDisabled returns true if the given rule ID is disabled.
func (c *Config) IsRuleDisabled(ruleID string) bool {
	for _, id := range c.DisabledRules {
		if id == ruleID {
			return true
		}
	}
	return false
}

// ShouldReport returns true if the issue should be reported based on config.
func (c *Config) ShouldReport(issue Issue) bool {
	if c.IsRuleDisabled(issue.Rule) {
		return false
	}
	// Lower severity value means higher priority (Error=0 is most severe)
	return issue.Severity <= c.MinSeverity
}
