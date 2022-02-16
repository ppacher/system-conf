package conf_test

import (
	"testing"

	"github.com/ppacher/system-conf/conf"
	"github.com/stretchr/testify/assert"
)

func TestSectionDecoder_EmptyOptions(t *testing.T) {
	specs := []conf.OptionSpec{
		{
			Name: "string slice",
			Type: conf.StringSliceType,
		},
		{
			Name: "string",
			Type: conf.StringType,
		},
	}

	decoder := conf.NewSectionDecoder(specs)

	res := decoder.AsMap(conf.Section{
		Options: []conf.Option{},
	})

	assert.Equal(t, map[string]interface{}{}, res)
}
