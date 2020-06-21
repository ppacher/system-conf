package conf

import "strings"

// Sections is a convenience type for working with a slice
// of sections.
type Sections []Section

// Get returns the section identified by name. Section names
// are compared using equal fold. If no section matches name
// nil is returned.
func (ss Sections) Get(name string) *Section {
	lower := strings.ToLower(name)
	for idx, s := range ss {
		if strings.ToLower(s.Name) == lower {
			return &ss[idx]
		}
	}
	return nil
}

// Has checks if a section with name is available. Section names
// are compared using equal fold.
func (ss Sections) Has(name string) bool {
	return ss.Get(name) != nil
}
