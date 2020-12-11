package conf

import "reflect"

// SectionUnmarshaler describes the interface that can be
// implemented to provide custom decoding of sections when
// using FileSpec.Decode.
// See OptionSpec for an example of how to use and implement
// UnmarshalSection.
type SectionUnmarshaler interface {
	UnmarshalSection(sec Section, spec SectionSpec) error
}

// Decode a file into target following the file specification.
func (spec FileSpec) Decode(file *File, target interface{}) error {
	return decodeFile(file, spec, reflect.ValueOf(target).Elem())
}

// DecodeValues decodes data into receiver. If receiver is a pointer to a
// nil interface a new value of the correct type will be created
// and stored. If receiver already has a type that Decode tries to parse
// data into receiver. If specType does not match receiver an error is
// returned.
func DecodeValues(data []string, specType OptionType, receiver interface{}) error {
	return decode(data, specType, reflect.ValueOf(receiver).Elem())
}
