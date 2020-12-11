package conf

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

	specCopy := make(FileSpec)
	for name, secSpec := range spec {
		specCopy[strings.ToLower(name)] = secSpec
	}

	return decodeFileToStruct(file, specCopy, outVal)
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

		name = strings.ToLower(name)

		secSpec, ok := spec[name]
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

		return decodeSections(sections, spec, reflect.Indirect(outVal))
	}

	// we might need to decode multiple sections
	if kind == reflect.Slice || kind == reflect.Array {
		valType := outVal.Type()
		valElemType := valType.Elem()
		sliceVal := outVal

		for i := 0; i < len(sections); i++ {
			for sliceVal.Len() <= i {
				sliceVal = reflect.Append(sliceVal, reflect.Zero(valElemType))
			}
			currentField := sliceVal.Index(i)

			if err := decodeSections(Sections{sections[i]}, spec, currentField); err != nil {
				return err
			}
		}

		outVal.Set(sliceVal)
		return nil
	}

	if kind != reflect.Struct {
		return fmt.Errorf("target must be of type %s", reflect.Struct)
	}

	if len(sections) != 1 {
		return fmt.Errorf("invalid number of sections, expected 1 but got %d", len(sections))
	}

	return decodeSectionToStruct(sections[0], spec, outVal)
}

func decodeSectionToStruct(section Section, spec SectionSpec, outVal reflect.Value) error {
	for i := 0; i < outVal.NumField(); i++ {
		fieldType := outVal.Type().Field(i)
		name := fieldType.Name

		if optionValue, ok := fieldType.Tag.Lookup("option"); ok && optionValue != "" {
			name = optionValue
			if name == "-" {
				continue
			}
		}

		name = strings.ToLower(name)

		var optionSpec *OptionSpec
		for _, opt := range spec {
			if strings.ToLower(opt.Name) == name {
				optionSpec = &opt
				break
			}
		}
		if optionSpec == nil {
			return fmt.Errorf("Cannot decode into unknown option %q", name)
		}

		if err := decode(section.GetStringSlice(optionSpec.Name), optionSpec.Type, outVal.Field(i)); err != nil {
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

func errInvalidType(specType OptionType, receiverType reflect.Type) error {
	return fmt.Errorf("failed to decode option of type %s into variable type %s", specType.String(), receiverType.String())
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
