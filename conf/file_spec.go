package conf

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// FileSpec describes all sections and the allowed options
// for each section. It implements the SectionRegistry
// interface.
type FileSpec map[string]SectionSpec

// OptionsForSection searches the FileSpec for the section spec with
// the given name. It implements the SectionRegistry.
func (spec FileSpec) OptionsForSection(name string) (OptionRegistry, bool) {
	key := strings.ToLower(name)
	if sec, ok := spec[key]; ok {
		return sec, true
	}

	for mk, mv := range spec {
		if strings.ToLower(mk) == key {
			return mv, true
		}
	}

	return nil, false
}

// Parse parses a configuration file from r, validates it against the spec
// and unmarshals it into target. Users that want to utilize drop-in files
// should take care of deserializing, validating, applying drop-ins
// and decoding into target themself.
func (spec FileSpec) Parse(path string, r io.Reader, target interface{}) error {
	content, err := Deserialize(path, r)
	if err != nil {
		return fmt.Errorf("failed to load: %w", err)
	}

	if err := ValidateFile(content, spec); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return spec.Decode(content, target)
}

// Decode a file into target following the file specification.
func (spec FileSpec) Decode(file *File, target interface{}) error {
	return decodeFile(file, spec, reflect.ValueOf(target).Elem())
}

// ParseFile is like Parse but opens the file at path.
func (spec FileSpec) ParseFile(path string, target interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer f.Close()

	return spec.Parse(path, f, target)
}
