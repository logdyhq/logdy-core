package utils

import "testing"

func TestParseRotateSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format",
			input:    "abc",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid number",
			input:    "abc123M",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "no unit",
			input:    "1024",
			expected: 1024,
			wantErr:  false,
		},
		{
			name:     "kilobytes",
			input:    "2K",
			expected: 2 * 1024,
			wantErr:  false,
		},
		{
			name:     "megabytes",
			input:    "3M",
			expected: 3 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "gigabytes",
			input:    "4G",
			expected: 4 * 1024 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "terabytes",
			input:    "5T",
			expected: 5 * 1024 * 1024 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "large number",
			input:    "1024M",
			expected: 1024 * 1024 * 1024,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRotateSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRotateSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseRotateSize() = %v, want %v", got, tt.expected)
			}
		})
	}
}
