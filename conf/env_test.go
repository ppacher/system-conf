package conf_test

import (
	"testing"

	"github.com/ppacher/system-conf/conf"
	"github.com/stretchr/testify/assert"
)

func TestParseEnv(t *testing.T) {
	fileSpec := conf.FileSpec{
		"foo": conf.SectionSpec{
			conf.OptionSpec{
				Name: "String",
				Type: conf.StringType,
			},
			conf.OptionSpec{
				Name: "Slice",
				Type: conf.StringSliceType,
			},
		},
		"Bar": conf.SectionSpec{
			conf.OptionSpec{
				Name: "String",
				Type: conf.StringType,
			},
			conf.OptionSpec{
				Name: "Slice",
				Type: conf.StringSliceType,
			},
		},
	}

	f, err := conf.ParseFromEnv("TEST", []string{
		"SOME_OTHER_ENV=test",
		"TEST_FOO_STRING=one value",
		"TEST_FOO_SLICE=first second",
		"TEST_BAR_Slice=first second",
		"TEST_BAR_1_Slice=third forth",
	}, fileSpec)
	assert.NoError(t, err)

	assert.NotNil(t, f.Get("foo"))
	assert.Equal(t, []string{"one value"}, f.Get("foo").GetStringSlice("string"))
	assert.Equal(t, []string{"first", "second"}, f.Get("foo").GetStringSlice("slice"))

	assert.Len(t, f.GetAll("bar"), 2)
	assert.Equal(t, []string{"first", "second"}, f.GetAll("bar")[0].GetStringSlice("slice"))
	assert.Equal(t, []string{"third", "forth"}, f.GetAll("bar")[1].GetStringSlice("slice"))
}
