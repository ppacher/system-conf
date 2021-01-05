package conf

import "errors"

// Commonly used validation and error messages.
var (
	ErrOptionRequired          = errors.New("option is required")
	ErrOptionAllowedOnce       = errors.New("option is only allowed once")
	ErrOptionNotExists         = errors.New("option does not exist")
	ErrInvalidBoolean          = errors.New("invalid boolean value")
	ErrInvalidFloat            = errors.New("invalid floating point number)")
	ErrInvalidNumber           = errors.New("invalid number")
	ErrInvalidDuration         = errors.New("invalid duration")
	ErrNoSections              = errors.New("task does not contain any sections")
	ErrUnknownSection          = errors.New("unknown section")
	ErrDropInSectionNotExists  = errors.New("section defined in drop-in does not exist")
	ErrDropInSectionNotAllowed = errors.New("drop-ins not allowed for not-unique sections")
	ErrNoOptions               = errors.New("no options defined")
)
