// Package scoring provides a rubric-based scoring system for agent sessions.
//
// Sessions are scored on 4 dimensions (0-3 scale each) for a max score of 12.
// Thresholds:
//   - 0-4:   Failure (CI fails)
//   - 5-7:   Partial (needs review)
//   - 8-10:  Success
//   - 11-12: Excellent
package scoring

import (
	"fmt"
	"strings"
)

// Rating is a score from 0-3 for a single dimension.
type Rating int

const (
	RatingNone      Rating = 0 // Did not meet criteria
	RatingPartial   Rating = 1 // Partially met criteria
	RatingGood      Rating = 2 // Met criteria with minor issues
	RatingExcellent Rating = 3 // Fully met or exceeded criteria
)

// String returns a human-readable rating.
func (r Rating) String() string {
	switch r {
	case RatingNone:
		return "None (0)"
	case RatingPartial:
		return "Partial (1)"
	case RatingGood:
		return "Good (2)"
	case RatingExcellent:
		return "Excellent (3)"
	default:
		return fmt.Sprintf("Unknown (%d)", r)
	}
}

// Dimension represents a scoring dimension.
type Dimension struct {
	Name        string
	Description string
	Rating      Rating
	Notes       string
}

// Score contains the complete scoring for a session.
type Score struct {
	// Dimensions are the individual scores
	Completeness       Dimension
	LintQuality        Dimension
	OutputValidity     Dimension
	QuestionEfficiency Dimension

	// Metadata
	Persona       string
	Scenario      string
	LintCycles    int
	QuestionCount int
}

// Total returns the sum of all dimension scores (0-12).
func (s Score) Total() int {
	return int(s.Completeness.Rating) +
		int(s.LintQuality.Rating) +
		int(s.OutputValidity.Rating) +
		int(s.QuestionEfficiency.Rating)
}

// Threshold returns the quality threshold name.
func (s Score) Threshold() string {
	total := s.Total()
	switch {
	case total >= 11:
		return "Excellent"
	case total >= 8:
		return "Success"
	case total >= 5:
		return "Partial"
	default:
		return "Failure"
	}
}

// Passed returns true if the score meets the minimum threshold for CI.
func (s Score) Passed() bool {
	return s.Total() >= 5
}

// NewScore creates a new Score with initialized dimensions.
func NewScore(persona, scenario string) *Score {
	return &Score{
		Persona:  persona,
		Scenario: scenario,
		Completeness: Dimension{
			Name:        "Completeness",
			Description: "Were all required resources generated?",
		},
		LintQuality: Dimension{
			Name:        "Lint Quality",
			Description: "Did the code pass linting?",
		},
		OutputValidity: Dimension{
			Name:        "Output Validity",
			Description: "Is the generated output valid?",
		},
		QuestionEfficiency: Dimension{
			Name:        "Question Efficiency",
			Description: "Did the agent ask an appropriate number of questions?",
		},
	}
}

// ScoreCompleteness scores based on how many resources were generated.
func ScoreCompleteness(expected, actual int) (Rating, string) {
	if expected == 0 {
		return RatingExcellent, "No resources expected"
	}

	ratio := float64(actual) / float64(expected)
	switch {
	case ratio >= 1.0:
		return RatingExcellent, fmt.Sprintf("All %d resources generated", expected)
	case ratio >= 0.8:
		return RatingGood, fmt.Sprintf("%d/%d resources generated", actual, expected)
	case ratio >= 0.5:
		return RatingPartial, fmt.Sprintf("%d/%d resources generated", actual, expected)
	default:
		return RatingNone, fmt.Sprintf("Only %d/%d resources generated", actual, expected)
	}
}

// ScoreLintQuality scores based on lint cycles needed.
func ScoreLintQuality(cycles int, passed bool) (Rating, string) {
	if !passed {
		return RatingNone, "Lint never passed"
	}

	switch cycles {
	case 0, 1:
		return RatingExcellent, fmt.Sprintf("Passed on cycle %d", cycles)
	case 2:
		return RatingGood, "Passed on cycle 2"
	case 3:
		return RatingPartial, "Passed on cycle 3 (max)"
	default:
		return RatingPartial, fmt.Sprintf("Passed on cycle %d", cycles)
	}
}

// ScoreOutputValidity scores based on lint results.
func ScoreOutputValidity(errors, warnings int) (Rating, string) {
	if errors > 0 {
		return RatingNone, fmt.Sprintf("%d errors in output", errors)
	}
	if warnings == 0 {
		return RatingExcellent, "Output is valid"
	}
	if warnings <= 2 {
		return RatingGood, fmt.Sprintf("%d warnings in output", warnings)
	}
	return RatingPartial, fmt.Sprintf("%d warnings in output", warnings)
}

// ScoreQuestionEfficiency scores based on number of clarifying questions.
func ScoreQuestionEfficiency(questions int) (Rating, string) {
	switch {
	case questions <= 2:
		return RatingExcellent, fmt.Sprintf("%d questions asked", questions)
	case questions <= 4:
		return RatingGood, fmt.Sprintf("%d questions asked", questions)
	case questions <= 6:
		return RatingPartial, fmt.Sprintf("%d questions asked", questions)
	default:
		return RatingNone, fmt.Sprintf("%d questions asked (too many)", questions)
	}
}

// String returns a formatted score summary.
func (s Score) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Score: %d/12 (%s)\n", s.Total(), s.Threshold()))
	b.WriteString(fmt.Sprintf("Persona: %s, Scenario: %s\n\n", s.Persona, s.Scenario))

	dims := []Dimension{
		s.Completeness,
		s.LintQuality,
		s.OutputValidity,
		s.QuestionEfficiency,
	}

	for _, d := range dims {
		b.WriteString(fmt.Sprintf("  %s: %s\n", d.Name, d.Rating))
		if d.Notes != "" {
			b.WriteString(fmt.Sprintf("    %s\n", d.Notes))
		}
	}

	return b.String()
}
