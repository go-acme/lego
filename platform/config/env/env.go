package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/log"
)

// Get environment variables.
func Get(names ...string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string

	for _, envVar := range names {
		value := GetOrFile(envVar)
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

// GetWithFallback Get environment variable values.
// The first name in each group is use as key in the result map.
//
// case 1:
//
//	// LEGO_ONE="ONE"
//	// LEGO_TWO="TWO"
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => "LEGO_ONE" = "ONE"
//
// case 2:
//
//	// LEGO_ONE=""
//	// LEGO_TWO="TWO"
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => "LEGO_ONE" = "TWO"
//
// case 3:
//
//	// LEGO_ONE=""
//	// LEGO_TWO=""
//	env.GetWithFallback([]string{"LEGO_ONE", "LEGO_TWO"})
//	// => error
func GetWithFallback(groups ...[]string) (map[string]string, error) {
	values := map[string]string{}

	var missingEnvVars []string

	for _, names := range groups {
		if len(names) == 0 {
			return nil, errors.New("undefined environment variable names")
		}

		value, envVar := getOneWithFallback(names[0], names[1:]...)
		if value == "" {
			missingEnvVars = append(missingEnvVars, envVar)
			continue
		}

		values[envVar] = value
	}

	if len(missingEnvVars) > 0 {
		return nil, fmt.Errorf("some credentials information are missing: %s", strings.Join(missingEnvVars, ","))
	}

	return values, nil
}

func GetOneWithFallback[T any](main string, defaultValue T, fn func(string) (T, error), names ...string) T {
	v, _ := getOneWithFallback(main, names...)

	value, err := fn(v)
	if err != nil {
		return defaultValue
	}

	return value
}

func getOneWithFallback(main string, names ...string) (string, string) {
	value := GetOrFile(main)
	if value != "" {
		return value, main
	}

	for _, name := range names {
		value := GetOrFile(name)
		if value != "" {
			return value, main
		}
	}

	return "", main
}

// GetOrDefaultString returns the given environment variable value as a string.
// Returns the default if the env var cannot be found.
func GetOrDefaultString(envVar, defaultValue string) string {
	return getOrDefault(envVar, defaultValue, ParseString)
}

// GetOrDefaultBool returns the given environment variable value as a boolean.
// Returns the default if the env var cannot be coopered to a boolean, or is not found.
func GetOrDefaultBool(envVar string, defaultValue bool) bool {
	return getOrDefault(envVar, defaultValue, strconv.ParseBool)
}

// GetOrDefaultInt returns the given environment variable value as an integer.
// Returns the default if the env var cannot be coopered to an int, or is not found.
func GetOrDefaultInt(envVar string, defaultValue int) int {
	return getOrDefault(envVar, defaultValue, strconv.Atoi)
}

// GetOrDefaultSecond returns the given environment variable value as a time.Duration (second).
// Returns the default if the env var cannot be coopered to an int, or is not found.
func GetOrDefaultSecond(envVar string, defaultValue time.Duration) time.Duration {
	return getOrDefault(envVar, defaultValue, ParseSecond)
}

func getOrDefault[T any](envVar string, defaultValue T, fn func(string) (T, error)) T {
	v, err := fn(GetOrFile(envVar))
	if err != nil {
		return defaultValue
	}

	return v
}

// GetOrFile Attempts to resolve 'key' as an environment variable.
// Failing that, it will check to see if '<key>_FILE' exists.
// If so, it will attempt to read from the referenced file to populate a value.
func GetOrFile(envVar string) string {
	envVarValue := os.Getenv(envVar)
	if envVarValue != "" {
		return envVarValue
	}

	fileVar := envVar + "_FILE"

	fileVarValue := os.Getenv(fileVar)
	if fileVarValue == "" {
		return envVarValue
	}

	fileContents, err := os.ReadFile(fileVarValue)
	if err != nil {
		log.Printf("Failed to read the file %s (defined by env var %s): %s", fileVarValue, fileVar, err)
		return ""
	}

	return strings.TrimSuffix(string(fileContents), "\n")
}

// ParseSecond parses env var value (string) to a second (time.Duration).
func ParseSecond(s string) (time.Duration, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	if v < 0 {
		return 0, fmt.Errorf("unsupported value: %d", v)
	}

	return time.Duration(v) * time.Second, nil
}

// ParseString parses env var value (string) to a string but throws an error when the string is empty.
func ParseString(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty string")
	}

	return s, nil
}

// ParsePairs parses a raw string of comma-separated key-value pairs into a map.
// Keys and values are separated by a colon and are trimmed of whitespace.
func ParsePairs(raw string) (map[string]string, error) {
	result := make(map[string]string)

	for pair := range strings.SplitSeq(strings.TrimSuffix(raw, ","), ",") {
		data := strings.Split(pair, ":")
		if len(data) != 2 {
			return nil, fmt.Errorf("incorrect pair: %s", pair)
		}

		result[strings.TrimSpace(data[0])] = strings.TrimSpace(data[1])
	}

	return result, nil
}
