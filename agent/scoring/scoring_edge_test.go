// This file contains edge case tests for the scoring package
package scoring

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestScoreCompleteness_EdgeCases tests edge cases for completeness scoring
func TestScoreCompleteness_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected int
		actual   int
		rating   Rating
		notes    string
	}{
		{
			name:     "zero_expected_zero_actual",
			expected: 0,
			actual:   0,
			rating:   RatingExcellent,
			notes:    "No resources expected",
		},
		{
			name:     "zero_expected_some_actual",
			expected: 0,
			actual:   5,
			rating:   RatingExcellent,
			notes:    "No resources expected",
		},
		{
			name:     "exact_match",
			expected: 100,
			actual:   100,
			rating:   RatingExcellent,
			notes:    "",
		},
		{
			name:     "more_than_expected",
			expected: 10,
			actual:   15,
			rating:   RatingExcellent,
			notes:    "",
		},
		{
			name:     "exactly_80_percent",
			expected: 10,
			actual:   8,
			rating:   RatingGood,
			notes:    "",
		},
		{
			name:     "exactly_50_percent",
			expected: 10,
			actual:   5,
			rating:   RatingPartial,
			notes:    "",
		},
		{
			name:     "just_below_50_percent",
			expected: 10,
			actual:   4,
			rating:   RatingNone,
			notes:    "",
		},
		{
			name:     "one_of_many",
			expected: 100,
			actual:   1,
			rating:   RatingNone,
			notes:    "",
		},
		{
			name:     "very_large_numbers",
			expected: 10000,
			actual:   10000,
			rating:   RatingExcellent,
			notes:    "",
		},
		{
			name:     "floating_point_edge",
			expected: 7,
			actual:   5,
			rating:   RatingPartial, // 5/7 = 0.714... >= 0.5
			notes:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, notes := ScoreCompleteness(tt.expected, tt.actual)
			assert.Equal(t, tt.rating, rating, "Rating mismatch")
			if tt.notes != "" {
				assert.Contains(t, notes, tt.notes)
			}
		})
	}
}

// TestScoreLintQuality_EdgeCases tests edge cases for lint quality scoring
func TestScoreLintQuality_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		cycles int
		passed bool
		rating Rating
	}{
		{
			name:   "zero_cycles_passed",
			cycles: 0,
			passed: true,
			rating: RatingExcellent,
		},
		{
			name:   "zero_cycles_not_passed",
			cycles: 0,
			passed: false,
			rating: RatingNone,
		},
		{
			name:   "many_cycles_passed",
			cycles: 10,
			passed: true,
			rating: RatingPartial,
		},
		{
			name:   "many_cycles_not_passed",
			cycles: 10,
			passed: false,
			rating: RatingNone,
		},
		{
			name:   "exactly_max_cycles",
			cycles: 3,
			passed: true,
			rating: RatingPartial,
		},
		{
			name:   "negative_cycles",
			cycles: -1,
			passed: true,
			rating: RatingPartial, // Negative values fall through to default case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, notes := ScoreLintQuality(tt.cycles, tt.passed)
			assert.Equal(t, tt.rating, rating)
			assert.NotEmpty(t, notes)
		})
	}
}

// TestScoreOutputValidity_EdgeCases tests edge cases for output validity scoring
func TestScoreOutputValidity_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errors   int
		warnings int
		rating   Rating
	}{
		{
			name:     "zero_errors_zero_warnings",
			errors:   0,
			warnings: 0,
			rating:   RatingExcellent,
		},
		{
			name:     "zero_errors_one_warning",
			errors:   0,
			warnings: 1,
			rating:   RatingGood,
		},
		{
			name:     "zero_errors_exactly_2_warnings",
			errors:   0,
			warnings: 2,
			rating:   RatingGood,
		},
		{
			name:     "zero_errors_many_warnings",
			errors:   0,
			warnings: 100,
			rating:   RatingPartial,
		},
		{
			name:     "one_error_zero_warnings",
			errors:   1,
			warnings: 0,
			rating:   RatingNone,
		},
		{
			name:     "many_errors_many_warnings",
			errors:   50,
			warnings: 50,
			rating:   RatingNone,
		},
		{
			name:     "negative_values",
			errors:   -1,
			warnings: -1,
			rating:   RatingGood, // errors not > 0, so checks warnings: -1 <= 2, so RatingGood
		},
		{
			name:     "very_large_numbers",
			errors:   math.MaxInt32,
			warnings: math.MaxInt32,
			rating:   RatingNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, notes := ScoreOutputValidity(tt.errors, tt.warnings)
			assert.Equal(t, tt.rating, rating)
			assert.NotEmpty(t, notes)
		})
	}
}

// TestScoreQuestionEfficiency_EdgeCases tests edge cases for question efficiency scoring
func TestScoreQuestionEfficiency_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		questions int
		rating    Rating
	}{
		{
			name:      "zero_questions",
			questions: 0,
			rating:    RatingExcellent,
		},
		{
			name:      "exactly_2_questions",
			questions: 2,
			rating:    RatingExcellent,
		},
		{
			name:      "exactly_4_questions",
			questions: 4,
			rating:    RatingGood,
		},
		{
			name:      "exactly_6_questions",
			questions: 6,
			rating:    RatingPartial,
		},
		{
			name:      "exactly_7_questions",
			questions: 7,
			rating:    RatingNone,
		},
		{
			name:      "many_questions",
			questions: 100,
			rating:    RatingNone,
		},
		{
			name:      "negative_questions",
			questions: -5,
			rating:    RatingExcellent, // Treated as less than or equal to 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, notes := ScoreQuestionEfficiency(tt.questions)
			assert.Equal(t, tt.rating, rating)
			assert.NotEmpty(t, notes)
		})
	}
}

// TestScore_Total_EdgeCases tests edge cases for total score calculation
func TestScore_Total_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ratings  [4]Rating
		expected int
	}{
		{
			name:     "all_zero",
			ratings:  [4]Rating{0, 0, 0, 0},
			expected: 0,
		},
		{
			name:     "all_max",
			ratings:  [4]Rating{3, 3, 3, 3},
			expected: 12,
		},
		{
			name:     "mixed",
			ratings:  [4]Rating{3, 2, 1, 3},
			expected: 9,
		},
		{
			name:     "all_partial",
			ratings:  [4]Rating{1, 1, 1, 1},
			expected: 4,
		},
		{
			name:     "all_good",
			ratings:  [4]Rating{2, 2, 2, 2},
			expected: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScore("test", "test")
			s.Completeness.Rating = tt.ratings[0]
			s.LintQuality.Rating = tt.ratings[1]
			s.OutputValidity.Rating = tt.ratings[2]
			s.QuestionEfficiency.Rating = tt.ratings[3]

			total := s.Total()
			assert.Equal(t, tt.expected, total)
		})
	}
}

// TestScore_Threshold_Boundaries tests threshold boundary conditions
func TestScore_Threshold_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		total     int
		threshold string
	}{
		{
			name:      "minimum_failure",
			total:     0,
			threshold: "Failure",
		},
		{
			name:      "maximum_failure",
			total:     4,
			threshold: "Failure",
		},
		{
			name:      "minimum_partial",
			total:     5,
			threshold: "Partial",
		},
		{
			name:      "maximum_partial",
			total:     7,
			threshold: "Partial",
		},
		{
			name:      "minimum_success",
			total:     8,
			threshold: "Success",
		},
		{
			name:      "maximum_success",
			total:     10,
			threshold: "Success",
		},
		{
			name:      "minimum_excellent",
			total:     11,
			threshold: "Excellent",
		},
		{
			name:      "maximum_excellent",
			total:     12,
			threshold: "Excellent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScore("test", "test")

			// Distribute the total across 4 dimensions
			remainder := tt.total
			for i := 0; i < 4 && remainder > 0; i++ {
				rating := min(3, remainder)
				remainder -= rating

				switch i {
				case 0:
					s.Completeness.Rating = Rating(rating)
				case 1:
					s.LintQuality.Rating = Rating(rating)
				case 2:
					s.OutputValidity.Rating = Rating(rating)
				case 3:
					s.QuestionEfficiency.Rating = Rating(rating)
				}
			}

			threshold := s.Threshold()
			assert.Equal(t, tt.threshold, threshold)
		})
	}
}

// TestScore_Passed_Boundaries tests passed/failed boundary
func TestScore_Passed_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		total  int
		passed bool
	}{
		{
			name:   "exactly_4_fails",
			total:  4,
			passed: false,
		},
		{
			name:   "exactly_5_passes",
			total:  5,
			passed: true,
		},
		{
			name:   "zero_fails",
			total:  0,
			passed: false,
		},
		{
			name:   "maximum_passes",
			total:  12,
			passed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScore("test", "test")

			// Set ratings to achieve target total (4 dimensions)
			s.Completeness.Rating = Rating(min(3, tt.total))
			remainder := tt.total - int(s.Completeness.Rating)
			s.LintQuality.Rating = Rating(min(3, remainder))
			remainder -= int(s.LintQuality.Rating)
			s.OutputValidity.Rating = Rating(min(3, remainder))
			remainder -= int(s.OutputValidity.Rating)
			s.QuestionEfficiency.Rating = Rating(remainder)

			passed := s.Passed()
			assert.Equal(t, tt.passed, passed)
		})
	}
}

// TestRating_String_UnknownValue tests unknown rating values
func TestRating_String_UnknownValue(t *testing.T) {
	t.Parallel()

	tests := []Rating{
		Rating(4),
		Rating(10),
		Rating(-1),
		Rating(255),
	}

	for _, r := range tests {
		t.Run(r.String(), func(t *testing.T) {
			s := r.String()
			assert.Contains(t, s, "Unknown")
			assert.Contains(t, s, "(")
			assert.Contains(t, s, ")")
		})
	}
}

// TestScore_String_Output tests the string formatting
func TestScore_String_Output(t *testing.T) {
	t.Parallel()

	s := NewScore("test-persona", "test-scenario")
	s.Completeness.Rating = RatingExcellent
	s.Completeness.Notes = "Test notes"
	s.LintQuality.Rating = RatingGood
	s.OutputValidity.Rating = RatingNone
	s.QuestionEfficiency.Rating = RatingExcellent

	output := s.String()

	// Verify key components are present
	assert.Contains(t, output, "Score:")
	assert.Contains(t, output, "/12")
	assert.Contains(t, output, "test-persona")
	assert.Contains(t, output, "test-scenario")
	assert.Contains(t, output, "Completeness")
	assert.Contains(t, output, "Lint Quality")
	assert.Contains(t, output, "Output Validity")
	assert.Contains(t, output, "Question Efficiency")
	assert.Contains(t, output, "Test notes")
}

// TestNewScore_Initialization tests that NewScore properly initializes all dimensions
func TestNewScore_Initialization(t *testing.T) {
	t.Parallel()

	s := NewScore("expert", "complex-scenario")

	assert.Equal(t, "expert", s.Persona)
	assert.Equal(t, "complex-scenario", s.Scenario)
	assert.Equal(t, 0, s.LintCycles)
	assert.Equal(t, 0, s.QuestionCount)

	// Check all dimensions are initialized with proper names and descriptions
	assert.Equal(t, "Completeness", s.Completeness.Name)
	assert.NotEmpty(t, s.Completeness.Description)
	assert.Equal(t, RatingNone, s.Completeness.Rating)

	assert.Equal(t, "Lint Quality", s.LintQuality.Name)
	assert.NotEmpty(t, s.LintQuality.Description)

	assert.Equal(t, "Output Validity", s.OutputValidity.Name)
	assert.NotEmpty(t, s.OutputValidity.Description)

	assert.Equal(t, "Question Efficiency", s.QuestionEfficiency.Name)
	assert.NotEmpty(t, s.QuestionEfficiency.Description)
}

// TestScoreWithEmptyStrings tests scoring with empty persona/scenario
func TestScoreWithEmptyStrings(t *testing.T) {
	t.Parallel()

	s := NewScore("", "")

	assert.Equal(t, "", s.Persona)
	assert.Equal(t, "", s.Scenario)
	assert.Equal(t, 0, s.Total())

	output := s.String()
	assert.NotEmpty(t, output)
}

// TestScoreWithVeryLongStrings tests scoring with very long strings
func TestScoreWithVeryLongStrings(t *testing.T) {
	t.Parallel()

	longString := string(make([]byte, 10000))
	s := NewScore(longString, longString)

	assert.Equal(t, longString, s.Persona)
	assert.Equal(t, longString, s.Scenario)

	output := s.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, longString)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
