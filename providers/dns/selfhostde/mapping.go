package selfhostde

import (
	"errors"
	"fmt"
	"strings"
)

const (
	lineSep   = ","
	recordSep = ":"
)

type Seq struct {
	cursor int
	ids    []string
}

func NewSeq(ids ...string) *Seq {
	return &Seq{ids: ids}
}

func (s *Seq) Next() string {
	if len(s.ids) == 1 {
		return s.ids[0]
	}

	v := s.ids[s.cursor]

	if s.cursor < len(s.ids)-1 {
		s.cursor++
	} else {
		s.cursor = 0
	}

	return v
}

func parseRecordsMapping(v string) (map[string]*Seq, error) {
	v = strings.ReplaceAll(v, " ", "")

	if v == "" {
		return nil, errors.New("empty mapping")
	}

	acc := map[string]*Seq{}

	for {
		index, err := safeIndex(v, lineSep)
		if err != nil {
			return nil, err
		}

		if index != -1 {
			name, seq, err := parseLine(v[:index])
			if err != nil {
				return nil, err
			}

			acc[name] = seq

			v = v[index+1:]

			continue
		}

		name, seq, errP := parseLine(v)
		if errP != nil {
			return nil, errP
		}

		acc[name] = seq

		return acc, nil
	}
}

func parseLine(line string) (string, *Seq, error) {
	idx, err := safeIndex(line, recordSep)
	if err != nil {
		return "", nil, err
	}

	if idx == -1 {
		return "", nil, fmt.Errorf("missing %q: %s", recordSep, line)
	}

	name := line[:idx]
	rawIDs := line[idx+1:]

	var ids []string
	var count int

	for {
		idx, err = safeIndex(rawIDs, recordSep)
		if err != nil {
			return "", nil, err
		}

		if count == 2 {
			return "", nil, fmt.Errorf("too many record IDs for one domain: %s", line)
		}

		if idx == -1 {
			ids = append(ids, rawIDs)
			break
		}

		ids = append(ids, rawIDs[:idx])
		count++

		// Data for the next iteration.
		rawIDs = rawIDs[idx+1:]
	}

	return name, NewSeq(ids...), nil
}

func safeIndex(v, sep string) (int, error) {
	index := strings.Index(v, sep)
	if index == 0 {
		return 0, fmt.Errorf("first char is %q: %s", sep, v)
	}

	if index == len(v)-1 {
		return 0, fmt.Errorf("last char is %q: %s", sep, v)
	}

	return index, nil
}
