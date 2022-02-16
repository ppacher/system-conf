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
			if _, err := fmt.Fprintf(w, "%s= %s\n", opt.Name, escaped); err != nil {
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

// EncodeToOptions encodes the value from x into one or more options
// with the given name.
func EncodeToOptions(name string, x interface{}) (Options, error) {
	val := reflect.ValueOf(x)
	opts := new(Options)

	if err := encodeBasic(val, name, opts, false); err != nil {
		return nil, err
	}

	return *opts, nil
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

		if err := encodeSection(fieldValue, name, result, nil); err != nil {
			return fmt.Errorf("failed to encode %s: %w", name, err)
		}

	}
	return nil
}

func encodeSection(val reflect.Value, name string, result *File, opts *Options) error {
	kind := getKind(val)
	if kind == reflect.Ptr {
		return encodeSection(reflect.Indirect(val), name, result, opts)
	}

	if kind == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := encodeSection(val.Index(i), name, result, opts); err != nil {
				return fmt.Errorf("failed to encode section at index %d: %w", i, err)
			}
		}

		return nil
	}

	if kind != reflect.Struct {
		return fmt.Errorf("cannot encode section from %s, expected a struct", kind)
	}

	inline := true
	if opts == nil {
		opts = new(Options)
		inline = false
	}

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

		if err := encodeBasic(fieldValue, name, opts, false); err != nil {
			return fmt.Errorf("cannot encode value of option %s: %w", name, err)
		}
	}

	if !inline && len(*opts) > 0 {
		result.Sections = append(result.Sections, Section{
			Name:    name,
			Options: *opts,
		})
	}

	return nil
}

func encodeBasic(val reflect.Value, name string, result *Options, includeZeroValues bool) error {
	// skip encoding if we have the zero value
	if !includeZeroValues {
		if val.IsZero() {
			return nil
		}
	}

	kind := getKind(val)
	if kind == reflect.Ptr {
		return encodeBasic(reflect.Indirect(val), name, result, includeZeroValues)
	}

	var value string
	x := val.Interface()

	switch kind {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			elem := reflect.ValueOf(val.Index(i).Interface())
			if err := encodeBasic(elem, name, result, true); err != nil {
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
	case reflect.Struct: // TODO(ppacher): we should only allow anonymous fields here
		return encodeSection(val, name, nil, result)
	default:
		return fmt.Errorf("unsupported basic type %s", kind)
	}

	*result = append(*result, Option{
		Name:  name,
		Value: value,
	})

	return nil
}
