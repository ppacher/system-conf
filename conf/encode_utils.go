package conf

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// WriteSectionsTo writes all sections to w .
func WriteSectionsTo(sections Sections, w io.Writer) error {
	for _, sec := range sections {
		if _, err := fmt.Fprintf(w, "[%s]\n", sec.Name); err != nil {
			return err
		}

		for _, opt := range sec.Options {
			escaped := strings.ReplaceAll(opt.Value, "\n", "\\\n\t")
			if _, err := fmt.Fprintf(w, "%s= %s", opt.Name, escaped); err != nil {
				return err
			}
		}
	}

	return nil
}

// ConvertToFile converts x to a File. x is expected to be or point to a struct
// type.
func ConvertToFile(x interface{}, path string) (*File, error) {
	val := reflect.ValueOf(x)
	f := &File{
		Path: path,
	}

	if err := encodeFile(val, f); err != nil {
		return nil, err
	}

	return f, nil
}

func encodeFile(val reflect.Value, result *File) error {
	kind := getKind(val)

	if kind == reflect.Ptr {
		val = reflect.Indirect(val)
		return encodeFile(val, result)
	}

	if kind != reflect.Struct {
		return fmt.Errorf("cannot convert type %s to file", kind)
	}

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		fieldType := val.Type().Field(i)

		name := fieldType.Name

		// Skip unexported fields
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}

		if tagVal, ok := fieldType.Tag.Lookup("section"); ok {
			parts := strings.Split(tagVal, ",")
			if parts[0] != "" {
				name = parts[0]
			}

			if name == "-" {
				continue
			}
		}

		if err := encodeSection(fieldValue, name, result); err != nil {
			return fmt.Errorf("failed to encode %s: %w", name, err)
		}

	}
	return nil
}

func encodeSection(val reflect.Value, name string, result *File) error {
	kind := getKind(val)
	if kind == reflect.Ptr {
		return encodeSection(reflect.Indirect(val), name, result)
	}

	if kind == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := encodeSection(val.Index(i), name, result); err != nil {
				return fmt.Errorf("failed to encode section at index %d: %w", i, err)
			}
		}

		return nil
	}

	if kind != reflect.Struct {
		return fmt.Errorf("cannot decode section from %s, expected a struct", kind)
	}

	var opts Options

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		fieldType := val.Type().Field(i)

		name := fieldType.Name

		// skip unexported fields
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}

		if tagValue, ok := fieldType.Tag.Lookup("option"); ok {
			parts := strings.Split(tagValue, ",")
			if parts[0] != "" {
				name = parts[0]
			}

			if name == "-" {
				continue
			}
		}

		if err := encodeBasic(fieldValue, name, &opts); err != nil {
			return fmt.Errorf("cannot encode value of option %s: %w", name, err)
		}
	}

	result.Sections = append(result.Sections, Section{
		Name:    name,
		Options: opts,
	})

	return nil
}

func encodeBasic(val reflect.Value, name string, result *Options) error {
	kind := getKind(val)
	if kind == reflect.Ptr {
		return encodeBasic(reflect.Indirect(val), name, result)
	}

	var value string
	x := val.Interface()

	switch kind {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if err := encodeBasic(elem, name, result); err != nil {
				return fmt.Errorf("failed to encode slice: %w", err)
			}
		}
		return nil

	case reflect.Bool:
		value = strconv.FormatBool(x.(bool))
	case reflect.Int, reflect.Uint:
		value = fmt.Sprintf("%d", x)
	case reflect.Float32:
		value = strings.TrimRight(fmt.Sprintf("%f", x), "0")
	case reflect.String:
		value = x.(string)
	default:
		return fmt.Errorf("unsupported basic type %s", kind)
	}

	*result = append(*result, Option{
		Name:  name,
		Value: value,
	})

	return nil
}
