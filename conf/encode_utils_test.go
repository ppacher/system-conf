package conf_test

import (
	"testing"

	"github.com/ppacher/system-conf/conf"
	"github.com/stretchr/testify/assert"
)

func TestEncodeSections(t *testing.T) {
	type Sec1 struct {
		Bool   bool
		Int    int
		Float  float32
		String string
	}

	type Sec_2 struct {
		Bool2   bool    `option:"Bool"`
		Int2    int     `option:"Int"`
		Float2  float64 `option:"Float"`
		String2 string  `option:"String"`
	}

	type Sec_3 struct {
		StringSlice []string
	}

	s := struct {
		Sec1
		Sec_2 Sec_2 `section:"Sec2"`
		Sec_3 `section:"Sec3"`
	}{
		Sec1: Sec1{
			Bool:   false,
			Int:    10,
			Float:  0.1,
			String: "test",
		},
		Sec_2: Sec_2{
			Bool2:   true,
			Int2:    11,
			Float2:  0.2,
			String2: "test2",
		},
		Sec_3: Sec_3{
			StringSlice: []string{"foo", "bar"},
		},
	}

	file, err := conf.ConvertToFile(s, "")
	assert.NoError(t, err)
	assert.Equal(t, conf.Sections{
		{
			Name: "Sec1",
			Options: conf.Options{
				// Bool is missing because we drop zero-values
				{Name: "Int", Value: "10"},
				{Name: "Float", Value: "0.1"},
				{Name: "String", Value: "test"},
			},
		},
		{
			Name: "Sec2",
			Options: conf.Options{
				{Name: "Bool", Value: "true"},
				{Name: "Int", Value: "11"},
				{Name: "Float", Value: "0.2"},
				{Name: "String", Value: "test2"},
			},
		},
		{
			Name: "Sec3",
			Options: conf.Options{
				{Name: "StringSlice", Value: "foo"},
				{Name: "StringSlice", Value: "bar"},
			},
		},
	}, file.Sections)
}
