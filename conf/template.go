package conf

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// IsSymlink returns true if file is a symlink.
func IsSymlink(file string) (bool, error) {
	stat, err := os.Lstat(file)
	if err != nil {
		return false, err
	}

	return stat.Mode() == os.ModeSymlink, nil
}

// TemplateInstanceName parses path and returns the
// template instance name if path represents a systemd
// template unit name. For my-webserver@config-1.service
// TemplateInstanceName returns "config-1", true.
func TemplateInstanceName(path string) (string, bool) {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	unitName := strings.TrimSuffix(base, ext)

	parts := strings.Split(unitName, "@")
	if len(parts) == 1 {
		return "", false
	}

	return strings.Join(parts[1:], "@"), true
}

var specifierRe = regexp.MustCompile("%.")

// Specifiers maps a alpha-numerical rune to a value.
type Specifiers map[rune]string

// Replace replaces all specifiers from sm in str and returns the result.
// If an specifier is unknown an error is returned.
func (sm Specifiers) Replace(str string) (string, error) {
	var err error
	res := specifierRe.ReplaceAllStringFunc(str, func(id string) string {
		r := []rune(id)[1]
		if r == '%' {
			return "%"
		}
		val, ok := sm[rune(r)]
		if !ok {
			err = errors.New("Unknown specifier " + id)
			return id
		}
		return val
	})

	return res, err
}

// Get returns the value fro val.
func (sm Specifiers) Get(val rune) (string, error) {
	ret, ok := sm[val]
	if !ok {
		return "", errors.New("Unknown specifier %" + string(val))
	}

	return ret, nil
}

// ReplaceSpecifiers replaces all specifiers in all section options
// of f. If an unknown identifier is encountered an error is returned.
// The original File f remains untouched.
func ReplaceSpecifiers(f *File, sm Specifiers) (*File, error) {
	copy := f.Clone()
	var err error
	for _, sec := range copy.Sections {
		for optIdx, opt := range sec.Options {
			sec.Options[optIdx].Value, err = sm.Replace(opt.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	return copy, nil
}
