package conf_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ppacher/system-conf/conf"
	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	cases := []struct {
		I string
		V conf.OptionType
		E error
	}{
		{
			`{"type": "string"}`,
			conf.StringType,
			nil,
		},
		{
			`{"type": ""}`,
			nil,
			nil,
		},
		{
			`{}`,
			nil,
			nil,
		},
		{
			`{"type": "[]int"}`,
			conf.IntSliceType,
			nil,
		},
		{
			`{"type": "unknown"}`,
			nil,
			conf.ErrUnknownOptionType,
		},
	}

	for idx, c := range cases {
		var o conf.OptionSpec
		err := json.Unmarshal([]byte(c.I), &o)

		if !errors.Is(err, c.E) {
			t.Logf("case %d:Expected error to be %v but got %v", idx, c.E, err)
			t.Fail()
		}

		if c.E == nil {
			assert.Equalf(t, c.V, o.Type, "case %d", idx)
		}
	}
}
