package wsl

import (
	"regexp"
	"strings"
)

// TranslatePath converts a Windows path (e.g., C:\Users) to a WSL path (e.g., /mnt/c/Users)
func TranslatePath(input string) string {
	// Root drive translation (e.g., C:\ -> /mnt/c/)
	re := regexp.MustCompile(`([A-Za-z]):\\`)
	result := re.ReplaceAllStringFunc(input, func(m string) string {
		drive := strings.ToLower(m[0:1])
		return "/mnt/" + drive + "/"
	})

	// Convert backslashes to forward slashes
	result = strings.ReplaceAll(result, "\\", "/")

	return result
}

// TranslateArgs iterates through arguments and translates those that look like Windows paths
func TranslateArgs(args []string) []string {
	translated := make([]string, len(args))
	for i, arg := range args {
		// Basic heuristic: if it contains ":\" or starts with a drive letter and has backslashes
		if strings.Contains(arg, ":\\") || (len(arg) >= 3 && arg[1] == ':' && strings.Contains(arg, "\\")) {
			translated[i] = TranslatePath(arg)
		} else {
			translated[i] = arg
		}
	}
	return translated
}
