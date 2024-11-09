package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Key struct {
	Name      string
	Algorithm string
	Secret    string
}

// ReadTSIGFile reads TSIG key file generated with `tsig-keygen`.
func ReadTSIGFile(filename string) (*Key, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	key := &Key{}

	var read bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), ";"))

		if line == "" {
			continue
		}

		if read && line == "}" {
			break
		}

		fields := strings.Fields(line)

		switch {
		case fields[0] == "key":
			read = true

			if len(fields) != 3 {
				return nil, fmt.Errorf("invalid key line: %s", line)
			}

			key.Name = safeUnquote(fields[1])

		case !read:
			continue

		default:
			if len(fields) != 2 {
				continue
			}

			v := safeUnquote(fields[1])

			switch safeUnquote(fields[0]) {
			case "algorithm":
				key.Algorithm = v
			case "secret":
				key.Secret = v
			default:
				continue
			}
		}
	}

	return key, nil
}

func safeUnquote(v string) string {
	if len(v) < 2 {
		// empty or single character string
		return v
	}

	if v[0] == '"' && v[len(v)-1] == '"' {
		// string wrapped in quotes
		return v[1 : len(v)-1]
	}

	return v
}
