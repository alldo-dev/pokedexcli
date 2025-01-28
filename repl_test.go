package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "  mixedCase Words   With Spaces  ",
			expected: []string{"mixedcase", "words", "with", "spaces"},
		},
		{
			input:    "singleword",
			expected: []string{"singleword"},
		},
		{
			input:    "   ",
			expected: []string{},
		},
		{
			input:    "",
			expected: []string{},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		// Check the length of the actual slice
		// if they don't match, use t.Errorf to print an error message
		// and fail the test
		if len(actual) != len(c.expected) {
			t.Errorf("For input '%s', expected length %d but got %d", c.input, len(c.expected), len(actual))
			continue
		}
		for i := range actual {
			// Check each word in the slice
			// if they don't match, use t.Errorf to print an error message
			// and fail the test
			if actual[i] != c.expected[i] {
				t.Errorf("For input '%s', expected word '%s' at index %d but got '%s'", c.input, c.expected[i], i, actual[i])
			}
		}

	}
}
