package conf

// SectionRegistry is used to validate sections when
// parsing files. It's default implementation is FileSpec.
type SectionRegistry interface {
	// OptionsForSection returns the option registry that defines all options
	// allowed in the section name. It returns a boolean value to
	// indicate if a section with name was found. If false is returned
	// the section is treated as unknown and an error is returned.
	// The section name is always in lowercase.
	OptionsForSection(name string) (OptionRegistry, bool)
}

// OptionRegistry is used to validate all options in a section.
// It's default implementation is SectionSpec.
type OptionRegistry interface {
	// hasOptoin returns true if the option with name is defined
	// in the option registry. optName is always lowercase.
	HasOption(optName string) bool

	// GetOption returns the definition of the option defined by optName.
	// optName is always lowercase.
	GetOption(optName string) (OptionSpec, bool)

	// All returns all options defined in the option registry (if supported).
	All() []OptionSpec
}
