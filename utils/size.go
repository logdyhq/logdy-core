package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

// ParseRotateSize parses a size string like "1G" or "250M" into bytes
func ParseRotateSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size specification")
	}

	// Regular expression to match a number followed by a unit (K, M, G, T)
	re := regexp.MustCompile(`^(\d+)([KMGT]?)$`)
	matches := re.FindStringSubmatch(sizeStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	// Parse the number part
	size, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %s", matches[1])
	}

	// Apply the unit multiplier
	switch matches[2] {
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	case "G":
		size *= 1024 * 1024 * 1024
	case "T":
		size *= 1024 * 1024 * 1024 * 1024
	}

	return size, nil
}
