package matching

import (
	"testing"
)

func TestFuzzyMatcher(t *testing.T) {
	fm := NewFuzzyMatcher()

	tests := []struct {
		name      string
		s1        string
		s2        string
		options   *FuzzyOptions
		expected  float64
		threshold float64
	}{
		{
			name:      "exact match",
			s1:        "device-01",
			s2:        "device-01",
			options:   &FuzzyOptions{MinSimilarity: 0.8},
			expected:  1.0,
			threshold: 1.0,
		},
		{
			name:      "case insensitive match",
			s1:        "Device-01",
			s2:        "device-01",
			options:   &FuzzyOptions{MinSimilarity: 0.8, CaseIgnore: true},
			expected:  1.0,
			threshold: 1.0,
		},
		{
			name:      "similar names (typo)",
			s1:        "device-01",
			s2:        "devise-01",
			options:   &FuzzyOptions{MinSimilarity: 0.7},
			expected:  0.85,
			threshold: 0.8,
		},
		{
			name:      "serial numbers with minor differences",
			s1:        "ABC123DEF456",
			s2:        "ABC123DF456",
			options:   &FuzzyOptions{MinSimilarity: 0.8},
			expected:  0.9,
			threshold: 0.85,
		},
		{
			name:      "completely different strings",
			s1:        "device-north",
			s2:        "switch-south",
			options:   &FuzzyOptions{MinSimilarity: 0.7},
			expected:  0.0,
			threshold: 0.3,
		},
		{
			name:      "normalized strings with punctuation",
			s1:        "device.01.prod",
			s2:        "device 01 prod",
			options:   &FuzzyOptions{MinSimilarity: 0.8, Normalize: true},
			expected:  0.95,
			threshold: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := fm.CalculateSimilarity(tt.s1, tt.s2, tt.options)

			if similarity < tt.threshold {
				t.Errorf("CalculateSimilarity(%q, %q) = %f, want >= %f",
					tt.s1, tt.s2, similarity, tt.threshold)
			}

			t.Logf("CalculateSimilarity(%q, %q) = %f", tt.s1, tt.s2, similarity)
		})
	}
}

func TestJaroWinklerSimilarity(t *testing.T) {
	fm := NewFuzzyMatcher()

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
		delta    float64
	}{
		{
			name:     "identical strings",
			s1:       "test",
			s2:       "test",
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "similar with common prefix",
			s1:       "device-01",
			s2:       "device-02",
			expected: 0.9,
			delta:    0.1,
		},
		{
			name:     "completely different",
			s1:       "abc",
			s2:       "xyz",
			expected: 0.0,
			delta:    0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := fm.jaroWinklerSimilarity(tt.s1, tt.s2)

			if similarity < tt.expected-tt.delta || similarity > tt.expected+tt.delta {
				t.Errorf("jaroWinklerSimilarity(%q, %q) = %f, want %f ± %f",
					tt.s1, tt.s2, similarity, tt.expected, tt.delta)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	fm := NewFuzzyMatcher()

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "test",
			s2:       "test",
			expected: 0,
		},
		{
			name:     "one character difference",
			s1:       "test",
			s2:       "best",
			expected: 1,
		},
		{
			name:     "insertion",
			s1:       "test",
			s2:       "tests",
			expected: 1,
		},
		{
			name:     "deletion",
			s1:       "tests",
			s2:       "test",
			expected: 1,
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "abc",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := fm.levenshteinDistance(tt.s1, tt.s2)

			if distance != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d",
					tt.s1, tt.s2, distance, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCalculateSimilarity(b *testing.B) {
	fm := NewFuzzyMatcher()
	options := &FuzzyOptions{MinSimilarity: 0.8, CaseIgnore: true, Normalize: true}

	for i := 0; i < b.N; i++ {
		fm.CalculateSimilarity("device-01-production", "devise-01-prod", options)
	}
}

func BenchmarkJaroWinklerSimilarity(b *testing.B) {
	fm := NewFuzzyMatcher()

	for i := 0; i < b.N; i++ {
		fm.jaroWinklerSimilarity("device-01-production", "devise-01-prod")
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	fm := NewFuzzyMatcher()

	for i := 0; i < b.N; i++ {
		fm.levenshteinDistance("device-01-production", "devise-01-prod")
	}
}
