package config

import (
	"os"
	"strings"
)

func IsEnvironmentVariable(valueString string) bool {
	if valueString == "" {
		return false
	}

	return strings.TrimSpace(valueString)[0] == '$'
}

func ResolveValue(valueString string) string {
	if len(valueString) <= 1 {
		return valueString
	}

	trimmedValue := strings.TrimSpace(valueString)
	if trimmedValue[0] == '$' {
		return os.Getenv(trimmedValue[1:])
	}

	return valueString
}
