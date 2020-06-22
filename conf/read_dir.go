package conf

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// ReadDir parses all files in directory that end in suffix. The sections of each file
// are validated against the spec map using the lowercase section name as the map key.
// If spec is nil no validation is performed.
func ReadDir(directory, suffix string, spec map[string][]OptionSpec) ([]*File, error) {
	entries, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var files []*File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), suffix) {
			continue
		}

		f, err := LoadFile(e.Name())
		if err != nil {
			return files, fmt.Errorf("%s: %w", e.Name(), err)
		}

		if err := ValidateFile(f, spec); err != nil {
			return files, fmt.Errorf("%s: %w", e.Name(), err)
		}

		files = append(files, f)
	}

	return files, nil
}
