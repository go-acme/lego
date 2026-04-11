package dotenv

import (
	"log/slog"
	"maps"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/log"
	"github.com/joho/godotenv"
)

const BaseFilePrefix = ".env"

func Load(filenames ...string) (func(), error) {
	// ONLY FOR TESTING PURPOSE: DON'T USE IT!!
	prefix, ok := os.LookupEnv("12b79c45_2153_4e99_9518_67b3350d878b")
	if ok {
		var prefixed []string

		for _, filename := range filenames {
			prefixed = append(prefixed, filepath.Join(prefix, filename))
		}

		filenames = prefixed
	}

	envs, err := read(filenames)
	if err != nil {
		return noopCleanUp, err
	}

	backup := make(map[string]string)

	var toDelete []string

	for k, v := range envs {
		existingValue, found := os.LookupEnv(k)
		if found {
			backup[k] = existingValue
		} else {
			toDelete = append(toDelete, k)
		}

		err = os.Setenv(k, v)
		if err != nil {
			return noopCleanUp, err
		}
	}

	return func() {
		for k, v := range backup {
			_ = os.Setenv(k, v)
		}

		for _, k := range toDelete {
			_ = os.Unsetenv(k)
		}
	}, nil
}

func noopCleanUp() {}

func read(filenames []string) (map[string]string, error) {
	envMap := make(map[string]string)

	for _, filename := range filenames {
		_, err := os.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug("Environment file not found", slog.String("filename", filename))

				continue
			}
		}

		data, err := readFile(filename)
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			continue
		}

		maps.Copy(envMap, data)
	}

	return envMap, nil
}

func readFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	return godotenv.Parse(file)
}
