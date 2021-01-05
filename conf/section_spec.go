package conf

import "strings"

// SectionSpec describes all options that can be used in a
// given section.
type SectionSpec []OptionSpec

// GetOption searches for the OptionSpec with name optName.
func (specs SectionSpec) GetOption(optName string) (OptionSpec, bool) {
	for _, opt := range specs {
		if strings.ToLower(opt.Name) == optName {
			return opt, true
		}
	}

	return OptionSpec{}, false
}

// HasOption returns true if the section spec defines an option
// with name optName.
func (specs SectionSpec) HasOption(optName string) bool {
	_, ok := specs.GetOption(optName)
	return ok
}

// All returns all options defined for the section.
func (specs SectionSpec) All() []OptionSpec {
	return ([]OptionSpec)(specs)
}
