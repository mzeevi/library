package data

import (
	"testing"
)

func TestAddFormatSuffix(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		format   string
		expected string
	}{
		{
			name:     "No suffix present",
			filename: "file",
			format:   "txt",
			expected: "file.txt",
		},
		{
			name:     "Suffix already present",
			filename: "file.txt",
			format:   "txt",
			expected: "file.txt",
		},
		{
			name:     "Different format",
			filename: "file.csv",
			format:   "txt",
			expected: "file.csv.txt",
		},
		{
			name:     "Empty filename",
			filename: "",
			format:   "txt",
			expected: ".txt",
		},
		{
			name:     "Empty format",
			filename: "file",
			format:   "",
			expected: "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addFormatSuffix(tt.filename, tt.format)
			if result != tt.expected {
				t.Errorf("addFormatSuffix(%v, %v) = %v; want %v", tt.filename, tt.format, result, tt.expected)
			}
		})
	}
}
