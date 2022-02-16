package conf

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckValue(t *testing.T) {
	cases := []struct {
		T OptionType
		V string
		E error
	}{
		{BoolType, "yes", nil},
		{BoolType, "false", nil},
		{BoolType, "0", nil},
		{BoolType, "foo", ErrInvalidBoolean},

		{IntType, "0x10", nil},
		{IntType, "0600", nil},
		{IntSliceType, "0b1100", nil},
		{IntType, "INVALID", ErrInvalidNumber},
		{IntSliceType, "INVALID2", ErrInvalidNumber},

		{FloatType, "0.5", nil},
		{FloatType, ".5", nil},
		{FloatSliceType, "0.1e10", nil},
		{FloatType, ".INVALID", ErrInvalidFloat},
		{FloatSliceType, "0.1eINVALID", ErrInvalidFloat},
		{DurationType, "5m", nil},
		{DurationType, "10h6s", nil},
		{DurationType, "asdf", ErrInvalidDuration},
		{DurationSliceType, "5m", nil},
		{DurationSliceType, "10h6s", nil},
		{DurationSliceType, "asdf", ErrInvalidDuration},

		{StringType, "", nil}, // empty strings ARE VALID
	}

	for idx, c := range cases {
		err := ValidateValue(c.V, c.T)

		if !errors.Is(err, c.E) {
			t.Errorf("case #%d (input=%v) expected error to be %s but got %s", idx, c.V, c.E, err)
		}
	}
}

func TestApplyDefaults(t *testing.T) {
	cases := []struct {
		I OptionSpec
		O Options
		V []string
	}{
		{
			I: OptionSpec{
				Name:    "test",
				Default: "some-value",
				Type:    StringType,
			},
			O: Options{
				{
					Name:  "Key2",
					Value: "something",
				},
			},
			V: []string{"some-value"},
		},
		// Ignore if there's no default
		{
			I: OptionSpec{
				Name:    "Test",
				Default: "",
				Type:    StringType,
			},
			O: Options{},
			V: nil,
		},
		// Ignore if required.
		{
			I: OptionSpec{
				Name:     "Test",
				Default:  "",
				Type:     StringType,
				Required: true,
			},
			O: Options{},
			V: nil,
		},
		{
			I: OptionSpec{
				Name:    "Test",
				Default: "val1",
				Type:    StringSliceType,
			},
			O: Options{},
			V: []string{"val1"},
		},
		{
			I: OptionSpec{
				Name:    "Test",
				Default: "val1",
				Type:    StringSliceType,
			},
			O: Options{
				{
					Name:  "Test",
					Value: "val2",
				},
			},
			V: []string{"val2"},
		},
	}

	for idx, c := range cases {
		opts := ApplyDefaults(c.O, SectionSpec{c.I})

		values := opts.GetStringSlice(c.I.Name)

		assert.Equal(t, c.V, values, "case #d", idx)
	}
}

func TestValidateOption(t *testing.T) {
	cases := []struct {
		I OptionSpec
		V []string
		E error
	}{
		{
			OptionSpec{
				Required: true,
				Type:     BoolType,
			},
			nil,
			ErrOptionRequired,
		},
		{
			OptionSpec{
				Required: true,
				Type:     BoolType,
			},
			[]string{""},
			ErrOptionRequired,
		},
		{
			OptionSpec{
				Required: true,
				Type:     StringSliceType,
			},
			[]string{"value", ""},
			ErrOptionRequired,
		},
		{
			OptionSpec{
				Type: StringType,
			},
			[]string{"one", "two"},
			ErrOptionAllowedOnce,
		},
		{
			OptionSpec{
				Type: IntSliceType,
			},
			[]string{"1", "2", "", "0.5"},
			ErrInvalidNumber,
		},
	}

	for idx, c := range cases {
		err := ValidateOption(c.V, c.I)
		if !errors.Is(err, c.E) {
			t.Errorf("cases #%d (input=%v): expected error to be '%v', got '%v'", idx, c.V, c.E, err)
		}
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		I SectionSpec
		V []Option
		E error
	}{
		{
			SectionSpec{
				{Name: "Option1", Type: StringType},
			},
			[]Option{
				{Name: "Option1", Value: "value"},
			},
			nil,
		},
		{
			SectionSpec{
				{Name: "Option1", Type: StringSliceType},
			},
			[]Option{
				{Name: "Option1", Value: "value"},
				{Name: "Option1", Value: "value"},
			},
			nil,
		},
		{
			SectionSpec{
				{Name: "Option1", Type: StringType},
			},
			[]Option{
				{Name: "Option1", Value: "value"},
				{Name: "Option1", Value: "value"},
			},
			ErrOptionAllowedOnce,
		},
		{
			SectionSpec{
				{Name: "Option1", Type: StringType},
			},
			[]Option{
				{Name: "Option2", Value: "value"},
			},
			ErrOptionNotExists,
		},
		{
			SectionSpec{
				{Name: "Option1", Type: StringType, Required: true},
			},
			nil,
			ErrOptionRequired,
		},
	}

	for idx, c := range cases {
		err := ValidateOptions(c.V, c.I)
		if !errors.Is(err, c.E) {
			t.Errorf("cases #%d (input=%v): expected errot to be '%v' but got '%v'", idx, c.V, c.E, err)
		}
	}
}
