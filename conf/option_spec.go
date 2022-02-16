package conf

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// OptionSpec describes an option
type OptionSpec struct {
	// Name is the name of the option.
	Name string `json:"name"`

	// Aliases is a set of aliases supported by this
	// option spec. If set, there should be a "Interal"
	// option for each alias name. Otherwise system-deploy
	// will throw an error if an alias is used in the
	// configuration. Use with care!
	Aliases []string `json:"aliases,omitempty"`

	// Description is a human readable description of
	// the option.
	Description string `json:"description,omitempty"`

	// Type defines the type of the option.
	Type OptionType `json:"type,omitempty" option:"-"`

	// Required may be set to true if deploy tasks must
	// specify this option.
	Required bool `json:"required,omitempty"`

	// Default may hold the default value for this option.
	// This value is only for help purposes and is NOT SET
	// as the default for that option.
	Default string `json:"default,omitempty"`

	// Internal may be set to true to omit the option from
	// the help page.
	Internal bool `json:"internal,omitempty"`

	// Annotations can be used to add arbitrary metadata to
	// option definitions. For example, such annotations can
	// be later used in help or documentation generators.
	// Note that it is recommended to ensure annotation values
	// are JSON or Gob serializable.
	Annotations Annotation `json:"annotation,omitempty"`
}

// HasAnnotation returns true if spec has an annotation with the
// given name.
func (spec *OptionSpec) HasAnnotation(name string) bool {
	if spec.Annotations == nil {
		return false
	}
	_, ok := spec.Annotations[name]
	return ok
}

// UnmarshalSection implements SectionUnmarshaller.
func (spec *OptionSpec) UnmarshalSection(sec Section, sectionSpec OptionRegistry) error {
	type alias OptionSpec
	if err := decodeSectionToStruct(sec, sectionSpec, reflect.ValueOf((*alias)(spec)).Elem()); err != nil {
		return err
	}

	// only parse the type member if it's specified in the specs.
	if sectionSpec.HasOption("type") {
		opt, err := sec.GetString("Type")
		if err != nil {
			return fmt.Errorf("Type: %w", err)
		}

		typePtr := TypeFromString(opt)
		if typePtr == nil {
			return fmt.Errorf("invalid type %q", opt)
		}

		spec.Type = *typePtr
	}

	return nil
}

// UnmarshalJSON unmarshals blob into spec.
func (spec *OptionSpec) UnmarshalJSON(blob []byte) error {
	type embed OptionSpec
	var wrapped struct {
		embed
		Type string `json:"type"` // must equal the json tag from OptionSpec
	}

	if err := json.Unmarshal(blob, &wrapped); err != nil {
		return err
	}

	*spec = OptionSpec(wrapped.embed)
	if wrapped.Type != "" {
		t := TypeFromString(wrapped.Type)
		if t == nil {
			return ErrUnknownOptionType
		}
		spec.Type = *t
	}
	return nil
}
