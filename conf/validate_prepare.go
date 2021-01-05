package conf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Prepare prepares the sec by applying default values and validating
// options against a set of option specs.
func Prepare(sec Section, specs OptionRegistry) (Section, error) {
	var copy = Section{
		Name:    sec.Name,
		Options: ApplyDefaults(sec.Options, specs),
	}

	if err := ValidateOptions(sec.Options, specs); err != nil {
		return copy, err
	}

	return copy, nil
}

// ValidateFile validates all sections in file and applies any
// default option values. If specs is nil then ValidateFile is
// a no-op.
func ValidateFile(file *File, specs SectionRegistry) error {
	if specs == nil {
		return nil
	}

	for idx, section := range file.Sections {
		secSpec, ok := specs.OptionsForSection(strings.ToLower(section.Name))
		if !ok {
			return fmt.Errorf("%s: %w", section.Name, ErrUnknownSection)
		}

		sec, err := Prepare(section, secSpec)
		if err != nil {
			return err
		}
		file.Sections[idx] = sec
	}

	return nil
}

// ApplyDefaults will add the default value for each option that
// is not specified but has an default set in it's spec.
func ApplyDefaults(options Options, specs OptionRegistry) Options {
	for _, spec := range specs.All() {
		if spec.Required {
			// if it's required we can skip that here because
			// Validate() would return an error anyway.
			continue
		}

		if spec.Default == "" {
			continue
		}

		var err error
		if spec.Type.IsSliceType() {
			// we use Required here because we need to get
			// the ErrOptionNotSet error
			_, err = options.GetRequiredStringSlice(spec.Name)
		} else {
			// GetString could actually return ErrOptionAllowedOnce too
			// be we don't care here because it means a value is set and
			// validate would fail anyway.
			_, err = options.GetString(spec.Name)
		}

		if err == ErrOptionNotSet {
			// we don't validate if spec.Default actually matches
			// spec.Type because Validate() would do it anyway.
			options = append(options, Option{
				Name:  spec.Name,
				Value: spec.Default,
			})
		}
	}

	return options
}

// ValidateOptions validates if all unit options specified in sec conform
// to the specification options.
func ValidateOptions(options Options, specs OptionRegistry) error {
	lm := make(map[string]OptionSpec)
	for _, spec := range specs.All() {
		lm[strings.ToLower(spec.Name)] = spec
	}

	// group option values by option name.
	gv := make(map[string][]string)
	for _, opt := range options {
		n := strings.ToLower(opt.Name)
		gv[n] = append(gv[n], opt.Value)
	}

	// validate
	for name, values := range gv {
		spec, ok := lm[strings.ToLower(name)]
		if !ok {
			// TODO(ppacher): we always use the lowercase version for the
			// error message here, use the original one instead.
			return fmt.Errorf("%s: %w", name, ErrOptionNotExists)
		}

		if err := ValidateOption(values, spec); err != nil {
			return fmt.Errorf("%s: %w", spec.Name, err)
		}

		// delete the spec from the lookup map
		// so any spec left-over may cause a Required
		// error.
		delete(lm, name)
	}

	// check if any option that is required is
	// missing completely
	for _, spec := range lm {
		if spec.Required {
			return fmt.Errorf("%s: %w", spec.Name, ErrOptionRequired)
		}
	}

	return nil
}

// ValidateOption validates if values matches spec.
func ValidateOption(values []string, spec OptionSpec) error {
	if len(values) > 1 && !spec.Type.IsSliceType() {
		return ErrOptionAllowedOnce
	}

	if spec.Required && len(values) == 0 {
		return ErrOptionRequired
	}

	for _, v := range values {
		// all occurences must have a value set
		// if the option is required.
		if spec.Required && v == "" {
			return ErrOptionRequired
		}

		// ensure the value matches the types expecations.
		if err := checkValue(v, spec.Type); err != nil {
			return err
		}
	}

	return nil
}

func checkValue(val string, optType OptionType) error {
	switch optType {
	case BoolType:
		if _, err := ConvertBool(val); err != nil {
			return ErrInvalidBoolean
		}
	case StringSliceType, StringType:
		// we cannot validate anything here
		return nil
	case FloatSliceType, FloatType:
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return ErrInvalidFloat
		}
	case IntSliceType, IntType:
		// we support all number formats supported by ParseInt.
		// That is, hex (0xYY), binary (0bYY) and octal (0YY)
		if _, err := strconv.ParseInt(val, 0, 64); err != nil {
			return ErrInvalidNumber
		}
	case DurationType, DurationSliceType:
		if _, err := time.ParseDuration(val); err != nil {
			return ErrInvalidDuration
		}
	}

	return nil
}
