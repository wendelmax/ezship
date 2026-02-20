package wsl

import (
	"testing"
)

func TestTranslatePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"C:\\Users", "/mnt/c/Users"},
		{"D:\\Projetos\\ezship", "/mnt/d/Projetos/ezship"},
		{"C:\\", "/mnt/c/"},
		{"random string", "random string"},
	}

	for _, tt := range tests {
		result := TranslatePath(tt.input)
		if result != tt.expected {
			t.Errorf("TranslatePath(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestTranslateArgs(t *testing.T) {
	args := []string{"run", "-v", "C:\\Users:/data", "ubuntu"}
	expected := []string{"run", "-v", "/mnt/c/Users:/data", "ubuntu"}

	result := TranslateArgs(args)
	for i, r := range result {
		if r != expected[i] {
			t.Errorf("TranslateArgs[%d] = %s; want %s", i, r, expected[i])
		}
	}
}
