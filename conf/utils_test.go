package conf_test

import (
	"testing"

	"github.com/ppacher/system-conf/conf"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		var x int

		err := conf.DecodeValues([]string{"10"}, conf.IntType, &x)
		assert.NoError(t, err)
		assert.Equal(t, 10, x)
	})

	t.Run("float", func(t *testing.T) {
		var x float64

		err := conf.DecodeValues([]string{"10.2"}, conf.FloatType, &x)
		assert.NoError(t, err)
		assert.Equal(t, 10.2, x)
	})

	t.Run("bool", func(t *testing.T) {
		var x bool

		err := conf.DecodeValues([]string{"yes"}, conf.BoolType, &x)
		assert.NoError(t, err)
		assert.Equal(t, true, x)
	})

	t.Run("interface", func(t *testing.T) {
		var x interface{}

		err := conf.DecodeValues([]string{"yes"}, conf.BoolType, &x)
		assert.NoError(t, err)
		assert.Equal(t, true, x)

		x = nil
		err = conf.DecodeValues([]string{"100"}, conf.IntType, &x)
		assert.NoError(t, err)
		assert.Equal(t, 100, x)

		x = nil
		err = conf.DecodeValues([]string{"1.0", "2.1", "3.2"}, conf.FloatSliceType, &x)
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.0, 2.1, 3.2}, x)
	})
}
