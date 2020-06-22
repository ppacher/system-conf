package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecifiers(t *testing.T) {
	spec := Specifiers{
		'i': "value",
		'I': "other value",
	}

	val, err := spec.Replace("Some %i and some %I %%")
	assert.NoError(t, err)
	assert.Equal(t, "Some value and some other value %", val)

	_, err = spec.Replace("Some unknown %u")
	assert.Error(t, err)
}

func TestReplaceSpecifiers(t *testing.T) {
	f := &File{
		Sections: Sections{
			{
				Name: "Section1",
				Options: Options{
					{
						Name:  "Key",
						Value: "Value %1",
					},
				},
			},
			{
				Name: "Section2",
				Options: Options{
					{
						Name:  "Key",
						Value: "Value %2",
					},
				},
			},
		},
	}

	f, err := ReplaceSpecifiers(f, Specifiers{
		'1': "SP1",
		'2': "SP2",
	})
	assert.NoError(t, err)
	assert.Equal(t, &File{
		Sections: Sections{
			{
				Name: "Section1",
				Options: Options{
					{
						Name:  "Key",
						Value: "Value SP1",
					},
				},
			},
			{
				Name: "Section2",
				Options: Options{
					{
						Name:  "Key",
						Value: "Value SP2",
					},
				},
			},
		},
	}, f)
}
