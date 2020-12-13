package conf

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// decodeFile decodes all sections of file as defined in spec into
// outVal. Note that outVal must be a direct or indirect struct
// type. outVal may be a nil struct-type value.
func decodeFile(file *File, spec FileSpec, outVal reflect.Value) error {
	kind := getKind(outVal)

	if kind == reflect.Ptr {
		valType := outVal.Type()
		valElemType := valType.Elem()

		if outVal.CanSet() {
			realVal := outVal
			if realVal.IsNil() {
				realVal = reflect.New(valElemType)
			}

			if err := decodeFile(file, spec, reflect.Indirect(realVal)); err != nil {
				return err
			}

			outVal.Set(realVal)
			return nil
		}

		return decodeFile(file, spec, reflect.Indirect(outVal))
	}

	if kind != reflect.Struct {
		return fmt.Errorf("target must be of type %s", reflect.Struct)
	}

	return decodeFileToStruct(file, spec, outVal)
}

func decodeFileToStruct(file *File, spec FileSpec, outVal reflect.Value) error {
	for i := 0; i < outVal.NumField(); i++ {
		fieldType := outVal.Type().Field(i)
		name := fieldType.Name
		required := false

		if sectionValue, ok := fieldType.Tag.Lookup("section"); ok {
			parts := strings.Split(sectionValue, ",")
			if parts[0] != "" {
				name = parts[0]
			}

			if name == "-" {
				continue
			}

			if len(parts) > 1 {
				for _, p := range parts[1:] {
					if p == "required" {
						required = true
					}
				}
			}
		}

		secSpec, ok := spec.FindSection(name)
		if !ok {
			return fmt.Errorf("no specification for section %q", name)
		}

		sections := file.GetAll(name)
		if len(sections) == 0 {
			if required {
				return fmt.Errorf("required section %q is missing", name)
			}

			continue
		}

		if err := decodeSections(sections, secSpec, outVal.Field(i)); err != nil {
			return fmt.Errorf("failed to decode section %s: %w", name, err)
		}
	}

	return nil
}

func decodeSections(sections Sections, spec SectionSpec, outVal reflect.Value) error {
	kind := getKind(outVal)

	if kind == reflect.Ptr {
		valType := outVal.Type()
		valElemType := valType.Elem()

		// if we can set outVal we may also need to create a new
		// value of the type it's pointing to, decode into that
		// and then set the new value as the target of outVal pointer
		if outVal.CanSet() {
			realVal := outVal
			if realVal.IsNil() {
				realVal = reflect.New(valElemType)
			}

			if err := decodeSections(sections, spec, reflect.Indirect(realVal)); err != nil {
				return err
			}

			outVal.Set(realVal)
			return nil
		}

		// Try to decode into the actual element outVal
		// points to.
		return decodeSections(sections, spec, reflect.Indirect(outVal))
	}

	// we might need to decode multiple sections
	if kind == reflect.Slice || kind == reflect.Array {
		valType := outVal.Type()
		valElemType := valType.Elem()
		sliceVal := outVal

		for i := 0; i < len(sections); i++ {
			// ensure there are enough elements available in
			// the target outVal slice.
			for sliceVal.Len() <= i {
				sliceVal = reflect.Append(sliceVal, reflect.Zero(valElemType))
			}
			currentField := sliceVal.Index(i)

			// decode the section into the current field. We are
			// calling into decodeSections again as it handles
			// currentField being a pointer or nil-value and will
			// eventually call decodeSectionToStruct and expect
			// only one section being passed.
			if err := decodeSections(Sections{sections[i]}, spec, currentField); err != nil {
				return err
			}
		}

		outVal.Set(sliceVal)
		return nil
	}

	// We only support decoding sections into struct-types here.
	// TODO(ppacher): may add support for maps as well.
	if kind != reflect.Struct {
		return fmt.Errorf("target must be of type %s", reflect.Struct)
	}

	// There must be exactly one section to decode now. Otherwise
	// there are multiple sections defined but the user expects
	// only one. Bail out here.
	if len(sections) != 1 {
		return fmt.Errorf("invalid number of sections, expected 1 but got %d", len(sections))
	}

	return decodeSectionToStruct(sections[0], spec, outVal)
}

func decodeSectionToStruct(section Section, spec SectionSpec, outVal reflect.Value) error {
	// If outVal is addressable and implements a SectionUnmarshaler
	// than we use UnmarshalSection instead of a reflection based
	// method.
	// Note that only the Ptr version of a value can implement
	// SectionUnmarshaler.
	var u SectionUnmarshaler
	if outVal.CanAddr() {
		if m, ok := outVal.Addr().Interface().(SectionUnmarshaler); ok {
			u = m
		}
	} else if m, ok := outVal.Interface().(SectionUnmarshaler); ok {
		u = m
	}

	if u != nil {
		if err := u.UnmarshalSection(section, spec); err != nil {
			return err
		}
	}

	for i := 0; i < outVal.NumField(); i++ {
		fieldType := outVal.Type().Field(i)
		name := fieldType.Name

		// Skip unexported struct fields.
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}

		// if we have an struct type here we may need to unmarshal the section into
		// and embedded struct.
		if fieldType.Anonymous {
			if fieldType.Type.Kind() == reflect.Struct || (fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct) {
				if err := decodeSections(Sections{section}, spec, outVal.Field(i)); err != nil {
					return fmt.Errorf("failed to unmarshal into anonymous field %s: %w", fieldType.Name, err)
				}
				continue
			}
		}

		if optionValue, ok := fieldType.Tag.Lookup("option"); ok && optionValue != "" {
			name = optionValue
			if name == "-" {
				continue
			}
		}

		optionSpec, ok := spec.FindOption(name)
		if !ok {
			// TODO(ppacher): add a strict mode that errors out here.
			continue
		}

		values := section.GetStringSlice(optionSpec.Name)
		if len(values) == 0 && !optionSpec.Required {
			continue
		}
		if err := decode(values, optionSpec.Type, outVal.Field(i)); err != nil {
			return fmt.Errorf("failed to unmarshal into field %s: %w", fieldType.Name, err)
		}
	}
	return nil
}

func decode(data []string, specType OptionType, outVal reflect.Value) error {
	kind := getKind(outVal)

	if !specType.IsSliceType() && len(data) != 1 {
		return fmt.Errorf("cannot convert %d values into basic value %s", len(data), kind)
	}

	switch kind {
	case reflect.Bool:
		return decodeBool(data[0], specType, outVal)
	case reflect.Int:
		return decodeInt(data[0], specType, outVal)
	case reflect.Float32:
		return decodeFloat(data[0], specType, outVal)
	case reflect.String:
		return decodeString(data[0], specType, outVal)
	case reflect.Interface:
		return decodeBasic(data, specType, outVal)
	case reflect.Ptr:
		return decodePtr(data, specType, outVal)
	case reflect.Slice:
		return decodeSlice(data, specType, outVal)
	}

	return fmt.Errorf("unsupported type: %s", kind.String())
}

func decodeBasic(data []string, specType OptionType, outVal reflect.Value) error {
	if outVal.IsValid() && outVal.Elem().IsValid() {
		elem := outVal.Elem()

		// if elem is not addressable we make a copy of it's value
		// and replace that.
		copied := false
		if !elem.CanAddr() {
			copied = true
			copy := reflect.New(elem.Type())
			copy.Elem().Set(elem)

			elem = copy
		}

		if err := decode(data, specType, elem); err != nil || !copied {
			return err
		}

		// If we are a copy we need to overwrite the original
		// value.
		outVal.Set(elem.Elem())
		return nil
	}

	var decodeType reflect.Type

	switch specType {
	case StringType:
		decodeType = reflect.TypeOf("")
	case StringSliceType:
		decodeType = reflect.TypeOf([]string{})
	case IntType:
		decodeType = reflect.TypeOf(int(0))
	case IntSliceType:
		decodeType = reflect.TypeOf([]int{})
	case FloatType:
		decodeType = reflect.TypeOf(float64(0))
	case FloatSliceType:
		decodeType = reflect.TypeOf([]float64{})
	case BoolType:
		decodeType = reflect.TypeOf(bool(true))
	default:
		return fmt.Errorf("unsupported type: %s", specType.String())
	}

	decoded := reflect.New(decodeType)

	if err := decode(data, specType, reflect.Indirect(decoded)); err != nil {
		return err
	}

	outVal.Set(reflect.Indirect(decoded))
	return nil
}

func decodeBool(data string, specType OptionType, outVal reflect.Value) error {
	if specType != BoolType {
		return errors.New("type mismatch")
	}

	b, err := ConvertBool(data)
	if err != nil {
		return err
	}

	outVal.SetBool(b)

	return nil
}

func decodeInt(data string, specType OptionType, outVal reflect.Value) error {
	if specType != IntType && specType != IntSliceType {
		return errors.New("invalid type")
	}

	val, err := strconv.ParseInt(data, 0, 64)
	if err != nil {
		return err
	}

	outVal.SetInt(val)
	return nil
}

func decodeFloat(data string, specType OptionType, outVal reflect.Value) error {
	if specType != FloatType && specType != FloatSliceType {
		return errors.New("invalid type")
	}

	val, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return err
	}

	outVal.SetFloat(val)
	return nil
}

func decodeString(data string, specType OptionType, outVal reflect.Value) error {
	if specType != StringType && specType != StringSliceType {
		return errors.New("invalid type")
	}

	outVal.SetString(data)
	return nil
}

func decodePtr(data []string, specType OptionType, outVal reflect.Value) error {
	valType := outVal.Type()
	valElemType := valType.Elem()

	if outVal.CanSet() {
		realVal := outVal
		if realVal.IsNil() {
			realVal = reflect.New(valElemType)
		}

		if err := decode(data, specType, reflect.Indirect(realVal)); err != nil {
			return err
		}

		outVal.Set(realVal)
	} else {
		if err := decode(data, specType, reflect.Indirect(outVal)); err != nil {
			return err
		}
	}
	return nil
}

func decodeSlice(data []string, specType OptionType, outVal reflect.Value) error {
	if !specType.IsSliceType() {
		return fmt.Errorf("cannot decode into %s, %s is not a slice type", getKind(outVal), specType)
	}

	valType := outVal.Type()
	valElemType := valType.Elem()

	sliceVal := outVal

	for i := 0; i < len(data); i++ {
		for sliceVal.Len() <= i {
			sliceVal = reflect.Append(sliceVal, reflect.Zero(valElemType))
		}
		currentField := sliceVal.Index(i)

		if err := decode([]string{data[i]}, specType, currentField); err != nil {
			return err
		}
	}

	outVal.Set(sliceVal)

	return nil
}

// getKind returns the kind of value but normalized Int, Uint and Float varaints
// to their base type.
func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}
