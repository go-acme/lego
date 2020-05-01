package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// ReadConfigFile reads a Plaintext Password struct from a JSON file.  This is loosely named
// because it's intended to eventually try different formats on parse failure (ie try JSON,
// fallback to YAML, fallback to XML, shoot: fallback to CSV if we need to provide the most options
// with sufficient return on investment.
func ReadConfigFile(filename string) (*PlaintextAuth, error) {
	var result = &PlaintextAuth{}

	if jsonFile, err := os.Open(filename); err == nil {
		defer jsonFile.Close()
		if byteValue, err := ioutil.ReadAll(jsonFile); err == nil {
			if err := json.Unmarshal([]byte(byteValue), result); err != nil {
				return nil, fmt.Errorf("Error parsing file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("Error reading file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("Error opening file: %v", err)
	}

	return result, nil
}
