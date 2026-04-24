package rfc2136

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	altEnvRFC2136Namespace    = "RFC2136_"
	altEnvRFC3645SubNamespace = "RFC3645_"
)

func altEnvNames(v string) []string {
	if strings.HasPrefix(v, envTSIGGSS) {
		return []string{
			strings.ReplaceAll(v,
				envTSIGGSS,
				envNamespace+altEnvRFC3645SubNamespace,
			),
			strings.ReplaceAll(v,
				envTSIGGSS,
				altEnvRFC2136Namespace+envSubTSIGGSS,
			),
		}
	}

	return []string{
		strings.ReplaceAll(v, envNamespace, altEnvRFC2136Namespace),
	}
}

func getEnvString(name string) string {
	return getOrDefaultString(name, "")
}

func getEnvStringSlice(name string) []string {
	v := getEnvString(name)
	if v == "" {
		return nil
	}

	return strings.Split(v, ",")
}

func getOrDefaultString(name, defaultValue string) string {
	return getOneWithFallback(name, defaultValue, env.ParseString)
}

func getOrDefaultSecond(name string, defaultValue time.Duration) time.Duration {
	return getOneWithFallback(name, defaultValue, env.ParseSecond)
}

func getOrDefaultInt(name string, defaultValue int) int {
	return getOneWithFallback(name, defaultValue, strconv.Atoi)
}

func getOneWithFallback[T any](main string, defaultValue T, fn func(string) (T, error)) T {
	return env.GetOneWithFallback(main, defaultValue, fn, altEnvNames(main)...)
}
