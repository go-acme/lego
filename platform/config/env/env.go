package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get environment variables
func Get(names ...string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string
	for _, envVar := range names {
		value := os.Getenv(envVar)
		if value == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
		values[envVar] = value
	}

	if len(missingEnvVars) > 0 {
		return nil, fmt.Errorf("some credentials information are missing: %s", strings.Join(missingEnvVars, ","))
	}

	return values, nil
}

// GetOrDefaultInt returns the given environment variable value as an integer.
// Returns the default if the envvar cannot be cooered to an int, or is not
// found.
func GetOrDefaultInt(envVar string, def int) int {
	s := os.Getenv(envVar)

	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return i
}
