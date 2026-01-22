package matching

import (
	"math"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// FuzzyMatcher provides database-agnostic fuzzy string matching algorithms
type FuzzyMatcher struct{}

// NewFuzzyMatcher creates a new fuzzy matcher instance
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{}
}

// CalculateSimilarity calculates similarity between two strings using multiple algorithms
// Returns a score between 0.0 and 1.0 where 1.0 is an exact match
func (fm *FuzzyMatcher) CalculateSimilarity(s1, s2 string, options *FuzzyOptions) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Apply normalization options
	if options != nil {
		if options.CaseIgnore {
			s1 = strings.ToLower(s1)
			s2 = strings.ToLower(s2)
		}

		if options.Normalize {
			s1 = fm.normalizeString(s1)
			s2 = fm.normalizeString(s2)
		}
	}

	// Use Jaro-Winkler similarity which works well for names and identifiers
	jaroWinkler := fm.jaroWinklerSimilarity(s1, s2)

	// For very short strings or when Jaro-Winkler gives low scores,
	// also consider Levenshtein distance
	if len(s1) < 4 || len(s2) < 4 || jaroWinkler < 0.7 {
		levenshtein := fm.levenshteinSimilarity(s1, s2)
		// Take the maximum of both algorithms
		return math.Max(jaroWinkler, levenshtein)
	}

	return jaroWinkler
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (fm *FuzzyMatcher) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a matrix to store distances
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = minInt(
				minInt(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion, insertion
				matrix[i-1][j-1]+cost,                      // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// levenshteinSimilarity converts Levenshtein distance to similarity score (0.0-1.0)
func (fm *FuzzyMatcher) levenshteinSimilarity(s1, s2 string) float64 {
	distance := fm.levenshteinDistance(s1, s2)
	maxLen := maxInt(len(s1), len(s2))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(distance)/float64(maxLen)
}

// jaroSimilarity calculates the Jaro similarity between two strings
func (fm *FuzzyMatcher) jaroSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate the match window
	matchWindow := maxInt(len1, len2)/2 - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matches := 0
	transpositions := 0

	// Find matches
	for i := 0; i < len1; i++ {
		start := maxInt(0, i-matchWindow)
		end := minInt(i+matchWindow+1, len2)

		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Find transpositions
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	// Calculate Jaro similarity
	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	return jaro
}

// jaroWinklerSimilarity calculates the Jaro-Winkler similarity between two strings
// This gives more weight to strings that have a common prefix
func (fm *FuzzyMatcher) jaroWinklerSimilarity(s1, s2 string) float64 {
	jaro := fm.jaroSimilarity(s1, s2)

	if jaro < 0.7 {
		return jaro
	}

	// Calculate common prefix length (up to 4 characters)
	prefixLen := 0
	for i := 0; i < minInt(minInt(len(s1), len(s2)), 4); i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	// Jaro-Winkler formula
	return jaro + (0.1 * float64(prefixLen) * (1.0 - jaro))
}

// normalizeString performs Unicode normalization and removes extra whitespace/punctuation
func (fm *FuzzyMatcher) normalizeString(s string) string {
	// Perform Unicode normalization (NFC - canonical composition)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, s)

	// Remove extra whitespace and normalize punctuation
	normalized = strings.TrimSpace(normalized)

	// Replace multiple whitespace with single space
	var result strings.Builder
	var prevSpace bool

	for _, r := range normalized {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
