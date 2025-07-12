package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2m3u/internal/config"
)

func TestFindConfigFile(t *testing.T) {
	testCases := map[string]struct {
		xdgEnvValue   string
		homeEnvValue  string
		expected      string
		expectedError error
	}{
		"ReturnsXdgFileWhenExists": {
			xdgEnvValue:   "testdata/xdg/valid",
			homeEnvValue:  "testdata/home/",
			expected:      "testdata/xdg/valid/plex2m3u/config.json",
			expectedError: nil,
		},
		"ReturnsHomeFileWhenExists": {
			xdgEnvValue:   "",
			homeEnvValue:  "testdata/home/",
			expected:      "testdata/home/.config/plex2m3u/config.json",
			expectedError: nil,
		},
		"ReturnsEmptyStringWhenNoEnvVarsAreSet": {
			xdgEnvValue:   "",
			homeEnvValue:  "",
			expected:      "",
			expectedError: nil,
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_DIR", tc.xdgEnvValue)
			t.Setenv("HOME", tc.homeEnvValue)

			file := config.FindConfigFile()
			assert.Equal(t, tc.expected, file)
		})
	}
}

func TestGet(t *testing.T) {
	testCases := map[string]struct {
		fileFlag      string
		xdgEnvValue   string
		expectedError error
		expected      config.Config
	}{
		"ReturnsErrorWhenFileIsInvalid": {
			fileFlag:      "",
			xdgEnvValue:   "testdata/xdg/invalid",
			expectedError: config.ErrUnmarshalingConfig,
			expected:      config.Config{},
		},
		"ReturnsFileSpecifiedByFileFlagIfValid": {
			fileFlag:      "testdata/xdg/valid/plex2m3u/config.json",
			xdgEnvValue:   "",
			expectedError: nil,
			expected:      config.Config{},
		},
		"ReturnsErrorIfFileFlagFileIsNotFound": {
			fileFlag:      "testdata/xdg/valid/plex2m3u/foo.json",
			xdgEnvValue:   "testdata/xdg/valid",
			expectedError: config.ErrConfigNotFound,
			expected:      config.Config{},
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_DIR", tc.xdgEnvValue)

			actual, err := config.Get(tc.fileFlag)
			require.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
