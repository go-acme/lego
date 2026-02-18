package internal

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type Records struct {
	Data   map[string]json.RawMessage `json:"data"`
	Status string                     `json:"status"`
}

type RecordValue map[string]string

func (r RecordValue) Set(key, value string) {
	r[key] = strconv.Quote(value)
}

func (r RecordValue) Add(key, value string) {
	r[key] = strings.TrimSpace(r[key] + " " + strconv.Quote(value))
}

func (r RecordValue) Delete(key string) {
	delete(r, key)
}

func (r RecordValue) RemoveValue(key, value string) {
	if r[key] == "" {
		return
	}

	quotedValue := strconv.Quote(value)

	data := strings.ReplaceAll(r[key], " "+quotedValue, "")
	data = strings.ReplaceAll(data, quotedValue+" ", "")
	data = strings.ReplaceAll(data, quotedValue, "")

	r[key] = data

	if r[key] == "" {
		r.Delete(key)
	}
}

func (r RecordValue) String() string {
	var parts []string

	for _, key := range slices.Sorted(maps.Keys(r)) {
		parts = append(parts, key+" "+r[key])
	}

	return strings.Join(parts, "\n")
}

func ParseRecordValue(lines string) RecordValue {
	data := make(RecordValue)

	for line := range strings.Lines(lines) {
		line = strings.TrimSpace(line)

		idx := strings.IndexFunc(line, unicode.IsSpace)

		data[line[:idx]] = line[idx+1:]
	}

	return data
}

func parseDomains(input string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(input))
	reader.Comma = '\t'
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	var data []string

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		if len(record) < 1 {
			// Malformed line
			continue
		}

		data = append(data, record[0])
	}

	return data, nil
}
