package conf

import (
	"fmt"
	"strings"
)

// OptionType describes the type of an option. It cannot
// be implemented outside the deploy package.
type OptionType interface {
	option() // ensure types can only be specified by this package.

	IsSliceType() bool

	fmt.Stringer
}

// All supported option types.
var (
	StringType      = option("string    ", false)
	StringSliceType = option("[]string  ", true)
	BoolType        = option("bool      ", false)
	IntType         = option("int       ", false)
	IntSliceType    = option("[]int     ", true)
	FloatType       = option("float     ", false)
	FloatSliceType  = option("[]float   ", true)
)

type optionType struct {
	name  string
	slice bool
}

func option(name string, slice bool) OptionType {
	return &optionType{
		name:  strings.Trim(name, " "),
		slice: slice,
	}
}

func (*optionType) option() {}

func (o *optionType) IsSliceType() bool { return o.slice }

func (o *optionType) String() string { return o.name }

// TypeFromString returns the option type described by str.
func TypeFromString(str string) *OptionType {
	switch str {
	case "string":
		return &StringType
	case "[]string":
		return &StringSliceType
	case "bool":
		return &BoolType
	case "int":
		return &IntType
	case "[]int":
		return &IntSliceType
	case "float":
		return &FloatType
	case "[]float":
		return &FloatSliceType
	}

	return nil
}
