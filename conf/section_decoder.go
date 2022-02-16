package conf

import "strings"

// SectionDecoder supports decoding a single section matching
// the decoders option specifications.
type SectionDecoder struct {
	specs map[string]OptionSpec
}

// NewSectionDecoder returns a new decoder.
func NewSectionDecoder(specs []OptionSpec) *SectionDecoder {
	dec := &SectionDecoder{}

	dec.specs = make(map[string]OptionSpec)
	for _, spec := range specs {
		dec.specs[strings.ToLower(spec.Name)] = spec
	}

	return dec
}

// AsMap returns a map representation of the section.
func (dec *SectionDecoder) AsMap(sec Section) map[string]interface{} {
	res := make(map[string]interface{})

	for _, opt := range dec.specs {
		val := sec.GetAs(opt.Name, opt.Type)
		if val == nil {
			continue
		}
		res[opt.Name] = val
	}

	return res
}

// Get returns the value of the option name in sec. If name is not set
// or does not exist nil is returned.
func (dec *SectionDecoder) Get(sec Section, name string) interface{} {
	spec, ok := dec.specs[strings.ToLower(name)]
	if !ok {
		return nil
	}

	return sec.GetAs(spec.Name, spec.Type)
}
