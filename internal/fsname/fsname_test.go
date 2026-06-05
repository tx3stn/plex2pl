package fsname_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tx3stn/plex2pl/internal/fsname"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    string
		expected string
	}{
		"ReturnsSafeNameUnchanged": {
			input:    "2026 jamz",
			expected: "2026 jamz",
		},
		"ReplacesForwardSlashes": {
			input:    "rock/metal mix",
			expected: "rock-metal mix",
		},
		"ReplacesBackslashes": {
			input:    `rock\metal mix`,
			expected: "rock-metal mix",
		},
		"TrimsLeadingAndTrailingDotsAndSpaces": {
			input:    " .hidden mix. ",
			expected: "hidden mix",
		},
		"ReturnsFallbackForPathTraversalName": {
			input:    "..",
			expected: "playlist",
		},
		"ReturnsFallbackForEmptyName": {
			input:    "",
			expected: "playlist",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, fsname.Sanitize(testCase.input))
		})
	}
}
