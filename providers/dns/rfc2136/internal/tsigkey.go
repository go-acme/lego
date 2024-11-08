package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

type Key struct {
	Name      string `mapstructure:"name"`
	Algorithm string `mapstructure:"algorithm"`
	Secret    string `mapstructure:"secret"`
}

// ReadTSIGFile reads TSIG key file generated with `tsig-keygen`.
func ReadTSIGFile(filename string) (*Key, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	data := make(map[string]string)

	var read bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), ";")

		if line == "}" {
			break
		}

		switch {
		case strings.HasPrefix(line, "key "):
			read = true
			fields := strings.Fields(line)

			if len(fields) != 3 {
				return nil, fmt.Errorf("invalid key line: %s", line)
			}

			data["name"] = safeUnquote(fields[1])

		case !read:
			continue

		default:
			fields := strings.Fields(line)

			if len(fields) != 2 {
				continue
			}

			data[safeUnquote(fields[0])] = safeUnquote(fields[1])
		}
	}

	key := &Key{}
	err = mapstructure.Decode(data, key)
	if err != nil {
		return nil, err
	}

	if key.Algorithm != "" {
		// to be compatible with https://github.com/miekg/dns/blob/master/tsig.go
		key.Algorithm += "."
	}

	return key, nil
}

func safeUnquote(v string) string {
	if v == "" {
		return v
	}

	if len(v)-1 != 0 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}

	return v
}
