package lint

import "testing"

func TestSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		expected int
	}{
		{"error is 0", SeverityError, 0},
		{"warning is 1", SeverityWarning, 1},
		{"info is 2", SeverityInfo, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.severity) != tt.expected {
				t.Errorf("Severity = %d, want %d", int(tt.severity), tt.expected)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		expected string
	}{
		{"error string", SeverityError, "error"},
		{"warning string", SeverityWarning, "warning"},
		{"info string", SeverityInfo, "info"},
		{"unknown severity", Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.expected {
				t.Errorf("Severity.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIssue(t *testing.T) {
	issue := Issue{
		Rule:       "TEST001",
		Message:    "test message",
		File:       "test.go",
		Line:       10,
		Column:     5,
		Severity:   SeverityError,
		Suggestion: "fix suggestion",
		Fixable:    true,
	}

	if issue.Rule != "TEST001" {
		t.Errorf("Issue.Rule = %q, want %q", issue.Rule, "TEST001")
	}
	if issue.Message != "test message" {
		t.Errorf("Issue.Message = %q, want %q", issue.Message, "test message")
	}
	if issue.File != "test.go" {
		t.Errorf("Issue.File = %q, want %q", issue.File, "test.go")
	}
	if issue.Line != 10 {
		t.Errorf("Issue.Line = %d, want %d", issue.Line, 10)
	}
	if issue.Column != 5 {
		t.Errorf("Issue.Column = %d, want %d", issue.Column, 5)
	}
	if issue.Severity != SeverityError {
		t.Errorf("Issue.Severity = %v, want %v", issue.Severity, SeverityError)
	}
	if issue.Suggestion != "fix suggestion" {
		t.Errorf("Issue.Suggestion = %q, want %q", issue.Suggestion, "fix suggestion")
	}
	if !issue.Fixable {
		t.Error("Issue.Fixable = false, want true")
	}
}

func TestConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := Config{}
		if cfg.MinSeverity != SeverityError {
			t.Errorf("default MinSeverity = %v, want %v", cfg.MinSeverity, SeverityError)
		}
		if len(cfg.DisabledRules) != 0 {
			t.Errorf("default DisabledRules = %v, want empty", cfg.DisabledRules)
		}
	})

	t.Run("configured config", func(t *testing.T) {
		cfg := Config{
			DisabledRules: []string{"TEST001", "TEST002"},
			MinSeverity:   SeverityWarning,
		}
		if len(cfg.DisabledRules) != 2 {
			t.Errorf("DisabledRules length = %d, want 2", len(cfg.DisabledRules))
		}
		if cfg.MinSeverity != SeverityWarning {
			t.Errorf("MinSeverity = %v, want %v", cfg.MinSeverity, SeverityWarning)
		}
	})
}

func TestConfigIsRuleDisabled(t *testing.T) {
	cfg := Config{
		DisabledRules: []string{"TEST001", "TEST002"},
	}

	tests := []struct {
		rule     string
		expected bool
	}{
		{"TEST001", true},
		{"TEST002", true},
		{"TEST003", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			if got := cfg.IsRuleDisabled(tt.rule); got != tt.expected {
				t.Errorf("IsRuleDisabled(%q) = %v, want %v", tt.rule, got, tt.expected)
			}
		})
	}
}

func TestConfigShouldReport(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		issue    Issue
		expected bool
	}{
		{
			name: "error severity matches error min",
			cfg:  Config{MinSeverity: SeverityError},
			issue: Issue{
				Rule:     "TEST001",
				Severity: SeverityError,
			},
			expected: true,
		},
		{
			name: "warning severity below error min",
			cfg:  Config{MinSeverity: SeverityError},
			issue: Issue{
				Rule:     "TEST001",
				Severity: SeverityWarning,
			},
			expected: false,
		},
		{
			name: "error severity exceeds warning min",
			cfg:  Config{MinSeverity: SeverityWarning},
			issue: Issue{
				Rule:     "TEST001",
				Severity: SeverityError,
			},
			expected: true,
		},
		{
			name: "disabled rule not reported",
			cfg: Config{
				MinSeverity:   SeverityInfo,
				DisabledRules: []string{"TEST001"},
			},
			issue: Issue{
				Rule:     "TEST001",
				Severity: SeverityError,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ShouldReport(tt.issue); got != tt.expected {
				t.Errorf("ShouldReport() = %v, want %v", got, tt.expected)
			}
		})
	}
}
