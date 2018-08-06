package placeholders

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	vars := map[string]interface{}{"ph1": "foo", "ph2": "bar", "num": 9}

	res, err := Resolve("foo", vars)
	assert.Nil(t, err)
	assert.Equal(t, "foo", res, "should return string with no placeholders as is")

	res, err = Resolve("/{ph1}/{ph2}", vars)
	assert.Nil(t, err)
	assert.Equal(t, "/foo/bar", res, "should resolve multiple placeholders in one var")

	res, err = Resolve("/{num}", vars)
	assert.Nil(t, err)
	assert.Equal(t, "/9", res, "should resolve number as string because of prefix /")

	res, err = Resolve("{num}", vars)
	assert.Nil(t, err)
	assert.Equal(t, 9, res, "should resolve as int")

	res, err = Resolve("{ph3}", vars)
	assert.Error(t, err, "should fail on non existing fields")
	assert.Nil(t, res)

	// resolve from extra
	ext1 := map[string]interface{}{"z": "baz"}
	ext2 := map[string]interface{}{"ph3": "qux"}
	ext3 := map[string]interface{}{"ph3": "baz"}
	res, err = Resolve("{ph3}", vars, ext1, ext2, ext3)
	assert.Nil(t, err)
	assert.Equal(t, "qux", res, "should resolve by the order of the given extras")
}
