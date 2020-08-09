package env

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertCanParse(t *testing.T, typ reflect.Type, parseable, boolean, multiple bool) {
	p, b, m := canParse(typ)
	assert.Equal(t, parseable, p, "expected %v to have parseable=%v but was %v", typ, parseable, p)
	assert.Equal(t, boolean, b, "expected %v to have boolean=%v but was %v", typ, boolean, b)
	assert.Equal(t, multiple, m, "expected %v to have multiple=%v but was %v", typ, multiple, m)
}

func TestCanParse(t *testing.T) {
	var (
		b  bool
		i  int
		s  string
		f  float64
		bs []bool
		is []int
	)

	assertCanParse(t, reflect.TypeOf(b), true, true, false)
	assertCanParse(t, reflect.TypeOf(i), true, false, false)
	assertCanParse(t, reflect.TypeOf(s), true, false, false)
	assertCanParse(t, reflect.TypeOf(f), true, false, false)

	assertCanParse(t, reflect.TypeOf(&b), true, true, false)
	assertCanParse(t, reflect.TypeOf(&s), true, false, false)
	assertCanParse(t, reflect.TypeOf(&i), true, false, false)
	assertCanParse(t, reflect.TypeOf(&f), true, false, false)

	assertCanParse(t, reflect.TypeOf(bs), true, true, true)
	assertCanParse(t, reflect.TypeOf(&bs), true, true, true)

	assertCanParse(t, reflect.TypeOf(is), true, false, true)
	assertCanParse(t, reflect.TypeOf(&is), true, false, true)
}

type implementsTextUnmarshaler struct{}

func (*implementsTextUnmarshaler) UnmarshalText(text []byte) error {
	return nil
}

func TestCanParseTextUnmarshaler(t *testing.T) {
	var (
		u  implementsTextUnmarshaler
		su []implementsTextUnmarshaler
	)

	assertCanParse(t, reflect.TypeOf(u), true, false, false)
	assertCanParse(t, reflect.TypeOf(&u), true, false, false)
	assertCanParse(t, reflect.TypeOf(su), true, false, true)
	assertCanParse(t, reflect.TypeOf(&su), true, false, true)
}

func TestCanNotParse(t *testing.T) {
	var envs struct{}

	assertCanParse(t, reflect.TypeOf(envs), false, false, false)
	assertCanParse(t, reflect.TypeOf(&envs), false, false, false)
}
