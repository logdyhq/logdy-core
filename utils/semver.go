package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ComparisonResult represents the result of comparing two semver strings
type ComparisonResult int

const (
	Less    ComparisonResult = -1
	Equal   ComparisonResult = 0
	Greater ComparisonResult = 1
)

// String returns a string representation of the comparison result
func (c ComparisonResult) String() string {
	switch c {
	case Less:
		return "less than"
	case Equal:
		return "equal to"
	case Greater:
		return "greater than"
	default:
		return "unknown"
	}
}

// CompareOptions defines which version components to include in comparison
type CompareOptions struct {
	CheckMajor bool
	CheckMinor bool
	CheckPatch bool
}

// DefaultCompareOptions returns the standard semver comparison options
// which checks all three components (major, minor, patch)
func DefaultCompareOptions() CompareOptions {
	return CompareOptions{
		CheckMajor: true,
		CheckMinor: true,
		CheckPatch: true,
	}
}

// CompareSemver compares two semantic version strings (A and B)
// with customizable options for which components to check
// Returns:
//
//	-1 (Less) if A < B
//	 0 (Equal) if A = B
//	 1 (Greater) if A > B
//
// Follows semantic versioning rules where versions are in the format: MAJOR.MINOR.PATCH
func CompareSemver(versionA, versionB string, options CompareOptions) (ComparisonResult, error) {
	// Parse version A
	partsA := strings.Split(versionA, ".")
	if len(partsA) != 3 {
		return 0, fmt.Errorf("version A (%s) is not in the format MAJOR.MINOR.PATCH", versionA)
	}

	// Parse version B
	partsB := strings.Split(versionB, ".")
	if len(partsB) != 3 {
		return 0, fmt.Errorf("version B (%s) is not in the format MAJOR.MINOR.PATCH", versionB)
	}

	// Compare MAJOR version if enabled
	if options.CheckMajor {
		majorA, err := strconv.Atoi(partsA[0])
		if err != nil {
			return 0, fmt.Errorf("invalid MAJOR version in A: %s", partsA[0])
		}

		majorB, err := strconv.Atoi(partsB[0])
		if err != nil {
			return 0, fmt.Errorf("invalid MAJOR version in B: %s", partsB[0])
		}

		if majorA != majorB {
			if majorA > majorB {
				return Greater, nil
			}
			return Less, nil
		}
	}

	// Compare MINOR version if enabled
	if options.CheckMinor {
		minorA, err := strconv.Atoi(partsA[1])
		if err != nil {
			return 0, fmt.Errorf("invalid MINOR version in A: %s", partsA[1])
		}

		minorB, err := strconv.Atoi(partsB[1])
		if err != nil {
			return 0, fmt.Errorf("invalid MINOR version in B: %s", partsB[1])
		}

		if minorA != minorB {
			if minorA > minorB {
				return Greater, nil
			}
			return Less, nil
		}
	}

	// Compare PATCH version if enabled
	if options.CheckPatch {
		patchA, err := strconv.Atoi(partsA[2])
		if err != nil {
			return 0, fmt.Errorf("invalid PATCH version in A: %s", partsA[2])
		}

		patchB, err := strconv.Atoi(partsB[2])
		if err != nil {
			return 0, fmt.Errorf("invalid PATCH version in B: %s", partsB[2])
		}

		if patchA != patchB {
			if patchA > patchB {
				return Greater, nil
			}
			return Less, nil
		}
	}

	// All checked components are equal
	return Equal, nil
}

// SimplifiedCompareSemver provides the original function signature for backward compatibility
// It compares all three version components (major, minor, patch)
func SimplifiedCompareSemver(versionA, versionB string) (ComparisonResult, error) {
	return CompareSemver(versionA, versionB, DefaultCompareOptions())
}
