package placeholders

import (
	"testing"
		"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	vars := map[string]interface{}{"ph1": "foo", "ph2": "bar", "num": 9}
	res, err := Resolve("/{ph1}/{ph2}", vars)
	assert.Nil(t, err)
	assert.Equal(t, "/foo/bar", res)

	res, err = Resolve("/{num}", vars)
	assert.Nil(t, err)
	assert.Equal(t, "/9", res)

	res, err = Resolve("{num}", vars)
	assert.Nil(t, err)
	assert.Equal(t, 9, res)
}
