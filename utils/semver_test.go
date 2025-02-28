package utils

import (
	"testing"
)

func TestCompareSemver(t *testing.T) {
	// Full comparison tests (all components checked)
	fullComparisonTests := []struct {
		name       string
		versionA   string
		versionB   string
		expected   ComparisonResult
		shouldFail bool
	}{
		// Equal versions
		{
			name:       "equal versions",
			versionA:   "1.2.3",
			versionB:   "1.2.3",
			expected:   Equal,
			shouldFail: false,
		},

		// Major version differences
		{
			name:       "greater major version",
			versionA:   "2.0.0",
			versionB:   "1.9.9",
			expected:   Greater,
			shouldFail: false,
		},
		{
			name:       "lesser major version",
			versionA:   "1.0.0",
			versionB:   "2.0.0",
			expected:   Less,
			shouldFail: false,
		},

		// Minor version differences
		{
			name:       "greater minor version",
			versionA:   "1.2.0",
			versionB:   "1.1.9",
			expected:   Greater,
			shouldFail: false,
		},
		{
			name:       "lesser minor version",
			versionA:   "1.1.0",
			versionB:   "1.2.0",
			expected:   Less,
			shouldFail: false,
		},

		// Patch version differences
		{
			name:       "greater patch version",
			versionA:   "1.2.3",
			versionB:   "1.2.2",
			expected:   Greater,
			shouldFail: false,
		},
		{
			name:       "lesser patch version",
			versionA:   "1.2.2",
			versionB:   "1.2.3",
			expected:   Less,
			shouldFail: false,
		},

		// Zero versions
		{
			name:       "zero versions",
			versionA:   "0.0.0",
			versionB:   "0.0.0",
			expected:   Equal,
			shouldFail: false,
		},

		// Large version numbers
		{
			name:       "large version numbers",
			versionA:   "999.999.999",
			versionB:   "999.999.998",
			expected:   Greater,
			shouldFail: false,
		},

		// Error cases
		{
			name:       "invalid format A - too few parts",
			versionA:   "1.2",
			versionB:   "1.2.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "invalid format B - too few parts",
			versionA:   "1.2.3",
			versionB:   "1.2",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "invalid format A - too many parts",
			versionA:   "1.2.3.4",
			versionB:   "1.2.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "invalid format B - too many parts",
			versionA:   "1.2.3",
			versionB:   "1.2.3.4",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "non-numeric major version A",
			versionA:   "a.2.3",
			versionB:   "1.2.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "non-numeric minor version A",
			versionA:   "1.b.3",
			versionB:   "1.2.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "non-numeric patch version A",
			versionA:   "1.2.c",
			versionB:   "1.2.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "non-numeric major version B",
			versionA:   "1.2.3",
			versionB:   "a.2.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "non-numeric minor version B",
			versionA:   "1.2.3",
			versionB:   "1.b.3",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
		{
			name:       "non-numeric patch version B",
			versionA:   "1.2.3",
			versionB:   "1.2.c",
			expected:   Equal, // doesn't matter, should fail
			shouldFail: true,
		},
	}

	t.Run("using SimplifiedCompareSemver", func(t *testing.T) {
		for _, test := range fullComparisonTests {
			t.Run(test.name, func(t *testing.T) {
				result, err := SimplifiedCompareSemver(test.versionA, test.versionB)

				// Check error expectations
				if test.shouldFail {
					if err == nil {
						t.Errorf("Expected error, but got nil")
					}
					return
				}

				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
					return
				}

				if result != test.expected {
					t.Errorf("Comparing %s and %s: Expected %s, got %s",
						test.versionA, test.versionB, test.expected, result)
				}
			})
		}
	})

	t.Run("using CompareSemver with default options", func(t *testing.T) {
		for _, test := range fullComparisonTests {
			t.Run(test.name, func(t *testing.T) {
				result, err := CompareSemver(test.versionA, test.versionB, DefaultCompareOptions())

				// Check error expectations
				if test.shouldFail {
					if err == nil {
						t.Errorf("Expected error, but got nil")
					}
					return
				}

				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
					return
				}

				if result != test.expected {
					t.Errorf("Comparing %s and %s: Expected %s, got %s",
						test.versionA, test.versionB, test.expected, result)
				}
			})
		}
	})

	// Test specific component comparison scenarios
	t.Run("custom comparison options", func(t *testing.T) {
		customTests := []struct {
			name     string
			versionA string
			versionB string
			options  CompareOptions
			expected ComparisonResult
		}{
			// Major-only comparison
			{
				name:     "major only - different major",
				versionA: "2.5.7",
				versionB: "1.9.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: false, CheckPatch: false},
				expected: Greater,
			},
			{
				name:     "major only - same major",
				versionA: "1.5.7",
				versionB: "1.9.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: false, CheckPatch: false},
				expected: Equal,
			},

			// Minor-only comparison
			{
				name:     "minor only - different minor",
				versionA: "1.5.7",
				versionB: "1.3.9",
				options:  CompareOptions{CheckMajor: false, CheckMinor: true, CheckPatch: false},
				expected: Greater,
			},
			{
				name:     "minor only - same minor",
				versionA: "2.5.7",
				versionB: "1.5.9",
				options:  CompareOptions{CheckMajor: false, CheckMinor: true, CheckPatch: false},
				expected: Equal,
			},

			// Patch-only comparison
			{
				name:     "patch only - different patch",
				versionA: "1.5.7",
				versionB: "2.3.4",
				options:  CompareOptions{CheckMajor: false, CheckMinor: false, CheckPatch: true},
				expected: Greater,
			},
			{
				name:     "patch only - same patch",
				versionA: "1.5.7",
				versionB: "2.3.7",
				options:  CompareOptions{CheckMajor: false, CheckMinor: false, CheckPatch: true},
				expected: Equal,
			},

			// Major and minor comparison
			{
				name:     "major and minor - different major",
				versionA: "2.5.7",
				versionB: "1.9.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: true, CheckPatch: false},
				expected: Greater,
			},
			{
				name:     "major and minor - same major, different minor",
				versionA: "1.5.7",
				versionB: "1.9.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: true, CheckPatch: false},
				expected: Less,
			},
			{
				name:     "major and minor - same major and minor",
				versionA: "1.5.7",
				versionB: "1.5.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: true, CheckPatch: false},
				expected: Equal,
			},

			// Major and patch comparison
			{
				name:     "major and patch - different major",
				versionA: "2.5.7",
				versionB: "1.9.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: false, CheckPatch: true},
				expected: Greater,
			},
			{
				name:     "major and patch - same major, different patch",
				versionA: "1.5.7",
				versionB: "1.9.9",
				options:  CompareOptions{CheckMajor: true, CheckMinor: false, CheckPatch: true},
				expected: Less,
			},

			// Minor and patch comparison
			{
				name:     "minor and patch - different minor",
				versionA: "1.6.7",
				versionB: "2.5.9",
				options:  CompareOptions{CheckMajor: false, CheckMinor: true, CheckPatch: true},
				expected: Greater,
			},
			{
				name:     "minor and patch - same minor, different patch",
				versionA: "1.5.7",
				versionB: "2.5.9",
				options:  CompareOptions{CheckMajor: false, CheckMinor: true, CheckPatch: true},
				expected: Less,
			},

			// No components checked
			{
				name:     "no components checked",
				versionA: "1.5.7",
				versionB: "2.9.3",
				options:  CompareOptions{CheckMajor: false, CheckMinor: false, CheckPatch: false},
				expected: Equal,
			},
		}

		for _, test := range customTests {
			t.Run(test.name, func(t *testing.T) {
				result, err := CompareSemver(test.versionA, test.versionB, test.options)

				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
					return
				}

				if result != test.expected {
					t.Errorf("Comparing %s and %s with options %+v: Expected %s, got %s",
						test.versionA, test.versionB, test.options, test.expected, result)
				}
			})
		}
	})
}

// Test the String method of ComparisonResult
func TestComparisonResultString(t *testing.T) {
	tests := []struct {
		result   ComparisonResult
		expected string
	}{
		{Less, "less than"},
		{Equal, "equal to"},
		{Greater, "greater than"},
		{ComparisonResult(99), "unknown"}, // Invalid value should return "unknown"
	}

	for _, test := range tests {
		result := test.result.String()
		if result != test.expected {
			t.Errorf("ComparisonResult(%d).String(): expected %q, got %q",
				test.result, test.expected, result)
		}
	}
}

func TestDefaultCompareOptions(t *testing.T) {
	options := DefaultCompareOptions()

	if !options.CheckMajor {
		t.Errorf("Expected DefaultCompareOptions().CheckMajor to be true")
	}

	if !options.CheckMinor {
		t.Errorf("Expected DefaultCompareOptions().CheckMinor to be true")
	}

	if !options.CheckPatch {
		t.Errorf("Expected DefaultCompareOptions().CheckPatch to be true")
	}
}
