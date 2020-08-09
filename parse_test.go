package env

import (
	"errors"
	"net"
	"net/mail"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type envsMap = map[string]string

func parse(envs envsMap, dest interface{}) error {
	_, err := pparse(envs, dest)

	return err
}

func pparse(envs envsMap, dest interface{}) (*Parser, error) {
	p, err := NewParser(Config{}, dest)
	if err != nil {
		return nil, err
	}

	os.Clearenv()

	for k, v := range envs {
		_ = os.Setenv(k, v)
	}

	err = p.Parse()

	return p, err
}

func TestString(t *testing.T) {
	var envs struct {
		Foo string
		Ptr *string
	}

	err := parse(envsMap{
		"foo": "bar",
		"ptr": "baz",
	}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "bar", envs.Foo)
	assert.Equal(t, "baz", *envs.Ptr)
}

func TestBool(t *testing.T) {
	var envs struct {
		A bool
		B bool
		C *bool
		D *bool
	}

	err := parse(envsMap{
		"a": "true",
		"c": "true",
	}, &envs)
	require.NoError(t, err)
	assert.True(t, envs.A)
	assert.False(t, envs.B)
	assert.True(t, *envs.C)
	assert.Nil(t, envs.D)
}

func TestInt(t *testing.T) {
	var envs struct {
		Foo int
		Ptr *int
	}

	err := parse(envsMap{
		"foo": "7",
		"ptr": "8",
	}, &envs)
	require.NoError(t, err)
	assert.EqualValues(t, 7, envs.Foo)
	assert.EqualValues(t, 8, *envs.Ptr)
}

func TestNegativeInt(t *testing.T) {
	var envs struct {
		Foo int
	}

	err := parse(envsMap{
		"foo": "-100",
	}, &envs)
	require.NoError(t, err)
	assert.EqualValues(t, envs.Foo, -100)
}

func TestNegativeIntAndFloatAndTricks(t *testing.T) {
	var envs struct {
		Foo int
		Bar float64
		N   int `env:"100"`
	}

	err := parse(envsMap{
		"foo": "-100",
		"bar": "-60.14",
		"100": "-100",
	}, &envs)
	require.NoError(t, err)
	assert.EqualValues(t, envs.Foo, -100)
	assert.EqualValues(t, envs.Bar, -60.14)
	assert.EqualValues(t, envs.N, -100)
}

func TestUint(t *testing.T) {
	var envs struct {
		Foo uint
		Ptr *uint
	}

	err := parse(envsMap{
		"foo": "7",
		"ptr": "8",
	}, &envs)
	require.NoError(t, err)
	assert.EqualValues(t, 7, envs.Foo)
	assert.EqualValues(t, 8, *envs.Ptr)
}

func TestFloat(t *testing.T) {
	var envs struct {
		Foo float32
		Ptr *float32
	}

	err := parse(envsMap{
		"foo": "3.4",
		"ptr": "3.5",
	}, &envs)
	require.NoError(t, err)
	assert.EqualValues(t, 3.4, envs.Foo)
	assert.EqualValues(t, 3.5, *envs.Ptr)
}

func TestDuration(t *testing.T) {
	var envs struct {
		Foo time.Duration
		Ptr *time.Duration
	}

	err := parse(envsMap{"foo": "3ms", "ptr": "4ms"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, 3*time.Millisecond, envs.Foo)
	assert.Equal(t, 4*time.Millisecond, *envs.Ptr)
}

func TestInvalidDuration(t *testing.T) {
	var envs struct {
		Foo time.Duration
	}

	err := parse(envsMap{"foo": "xxx"}, &envs)
	require.Error(t, err)
}

func TestIntPtr(t *testing.T) {
	var envs struct {
		Foo *int
	}

	err := parse(envsMap{"foo": "123"}, &envs)
	require.NoError(t, err)
	require.NotNil(t, envs.Foo)
	assert.Equal(t, 123, *envs.Foo)
}

func TestIntPtrNotPresent(t *testing.T) {
	var envs struct {
		Foo *int
	}

	err := parse(envsMap{}, &envs)
	require.NoError(t, err)
	assert.Nil(t, envs.Foo)
}

func TestMixed(t *testing.T) {
	var envs struct {
		Foo  string `env:"f"`
		Bar  int
		Baz  uint
		Ham  bool
		Spam float32
	}

	envs.Bar = 3
	err := parse(envsMap{"baz": "123", "spam": "1.2", "ham": "true", "f": "xyz"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "xyz", envs.Foo)
	assert.Equal(t, 3, envs.Bar)
	assert.Equal(t, uint(123), envs.Baz)
	assert.Equal(t, true, envs.Ham)
	assert.EqualValues(t, 1.2, envs.Spam)
}

func TestRequired(t *testing.T) {
	var envs struct {
		Foo string `env:"required"`
	}

	err := parse(envsMap{}, &envs)
	require.Error(t, err, "foo is required")
}

func TestLongFlag(t *testing.T) {
	var envs struct {
		Foo string `env:"abc"`
	}

	err := parse(envsMap{"abc": "xyz"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "xyz", envs.Foo)
}

func TestCaseSensitive(t *testing.T) {
	var envs struct {
		Lower bool `env:"v"`
		Upper bool `env:"V"`
	}

	err := parse(envsMap{"v": "true"}, &envs)
	require.NoError(t, err)
	assert.True(t, envs.Lower)
	assert.False(t, envs.Upper)
}

func TestCaseSensitive2(t *testing.T) {
	var envs struct {
		Lower bool `env:"v"`
		Upper bool `env:"V"`
	}

	err := parse(envsMap{"V": "true"}, &envs)
	require.NoError(t, err)
	assert.False(t, envs.Lower)
	assert.True(t, envs.Upper)
}

func TestMultiple(t *testing.T) {
	var envs struct {
		Foo []int
		Bar []string
	}

	err := parse(envsMap{"foo": "1,2,3", "bar": "x,y,z"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, envs.Foo)
	assert.Equal(t, []string{"x", "y", "z"}, envs.Bar)
}

func TestMultipleWithEq(t *testing.T) {
	var envs struct {
		Foo []int
		Bar []string
	}

	err := parse(envsMap{"foo": "1,2,3", "bar": "x"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, envs.Foo)
	assert.Equal(t, []string{"x"}, envs.Bar)
}

func TestMultipleWithDefault(t *testing.T) {
	var envs struct {
		Foo []int
		Bar []string
	}

	envs.Foo = []int{42}
	envs.Bar = []string{"foo"}
	err := parse(envsMap{"foo": "1,2,3", "bar": "x,y,z"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, envs.Foo)
	assert.Equal(t, []string{"x", "y", "z"}, envs.Bar)
}

func TestExemptField(t *testing.T) {
	var envs struct {
		Foo string
		Bar interface{} `env:"-"`
	}

	err := parse(envsMap{"foo": "xyz"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "xyz", envs.Foo)
}

func TestMissingRequired(t *testing.T) {
	var envs struct {
		Foo string `env:"required"`
		X   string
	}

	err := parse(envsMap{"x": "bar"}, &envs)
	assert.Error(t, err)
}

func TestInvalidInt(t *testing.T) {
	var envs struct {
		Foo int
	}

	err := parse(envsMap{"foo": "xyz"}, &envs)
	assert.Error(t, err)
}

func TestInvalidUint(t *testing.T) {
	var envs struct {
		Foo uint
	}

	err := parse(envsMap{"foo": "xyz"}, &envs)
	assert.Error(t, err)
}

func TestInvalidFloat(t *testing.T) {
	var envs struct {
		Foo float64
	}

	err := parse(envsMap{"foo": "xyz"}, &envs)
	require.Error(t, err)
}

func TestInvalidBool(t *testing.T) {
	var envs struct {
		Foo bool
	}

	err := parse(envsMap{"foo": "xyz"}, &envs)
	require.Error(t, err)
}

func TestInvalidIntSlice(t *testing.T) {
	var envs struct {
		Foo []int
	}

	err := parse(envsMap{"foo": "1 2 xyz"}, &envs)
	require.Error(t, err)
}

func TestErrorOnNonPointer(t *testing.T) {
	var envs struct{}
	err := parse(envsMap{}, envs)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrorNotPointers))
}

func TestErrorOnNonStruct(t *testing.T) {
	var envs string
	err := parse(envsMap{}, &envs)
	assert.Error(t, err)
}

func TestUnsupportedType(t *testing.T) {
	var envs struct {
		Foo interface{}
	}

	err := parse(envsMap{"foo": ""}, &envs)
	assert.Error(t, err)
}

func TestUnsupportedSliceElement(t *testing.T) {
	var envs struct {
		Foo []interface{}
	}

	err := parse(envsMap{"foo": "3"}, &envs)
	assert.Error(t, err)
}

func TestUnknownTag(t *testing.T) {
	var envs struct {
		Foo string `env:"this_is_not_valid:1"`
	}

	err := parse(envsMap{"foo": "xyz"}, &envs)
	assert.Error(t, err)

	var envs2 struct {
		Foo string `env:"name,name2"`
	}

	err = parse(envsMap{"name": "xyz"}, &envs2)
	assert.Error(t, err)
	assert.EqualError(t, err, ".Foo: name2-unrecognized tag")
}

func TestParse(t *testing.T) {
	var envs struct {
		Foo string
	}

	err := parse(envsMap{"foo": "bar"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "bar", envs.Foo)
}

func TestParseError(t *testing.T) {
	var envs struct {
		Foo string `env:"this_is_not_valid:2"`
	}

	err := Parse(&envs)
	assert.Error(t, err)
}

func TestMustParse(t *testing.T) {
	var envs struct {
		Foo string
	}

	_ = os.Setenv("foo", "bar")
	parser, err := MustParse(&envs)
	require.NoError(t, err)
	assert.Equal(t, "bar", envs.Foo)
	assert.NotNil(t, parser)
}

type textUnmarshaler struct {
	val int
}

func (f *textUnmarshaler) UnmarshalText(b []byte) error {
	f.val = len(b)

	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var envs struct {
		Foo textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, 3, envs.Foo.val)
}

func TestPtrToTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var envs struct {
		Foo *textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, 3, envs.Foo.val)
}

func TestRepeatedTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var envs struct {
		Foo []textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc,d,ef"}, &envs)
	require.NoError(t, err)
	require.Len(t, envs.Foo, 3)
	assert.Equal(t, 3, envs.Foo[0].val)
	assert.Equal(t, 1, envs.Foo[1].val)
	assert.Equal(t, 2, envs.Foo[2].val)
}

func TestRepeatedPtrToTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var envs struct {
		Foo []*textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc,d,ef"}, &envs)
	require.NoError(t, err)
	require.Len(t, envs.Foo, 3)
	assert.Equal(t, 3, envs.Foo[0].val)
	assert.Equal(t, 1, envs.Foo[1].val)
	assert.Equal(t, 2, envs.Foo[2].val)
}

type boolUnmarshaler bool

func (p *boolUnmarshaler) UnmarshalText(b []byte) error {
	*p = len(b)%2 == 0

	return nil
}

func TestBoolUnmarhsaler(t *testing.T) {
	// test that a bool type that implements TextUnmarshaler is
	// handled as a TextUnmarshaler not as a bool
	var envs struct {
		Foo *boolUnmarshaler
	}

	err := parse(envsMap{"foo": "ab"}, &envs)
	require.NoError(t, err)
	assert.EqualValues(t, true, *envs.Foo)
}

type sliceUnmarshaler []int

func (p *sliceUnmarshaler) UnmarshalText(b []byte) error {
	*p = sliceUnmarshaler{len(b)}

	return nil
}

func TestSliceUnmarhsaler(t *testing.T) {
	// test that a slice type that implements TextUnmarshaler is
	// handled as a TextUnmarshaler not as a slice
	var envs struct {
		Foo *sliceUnmarshaler
		Bar string `env:""`
	}

	err := parse(envsMap{"foo": "abcde"}, &envs)
	require.NoError(t, err)
	require.Len(t, *envs.Foo, 1)
	assert.EqualValues(t, 5, (*envs.Foo)[0])
	assert.Equal(t, "", envs.Bar)
}

func TestIP(t *testing.T) {
	var envs struct {
		Host net.IP
	}

	err := parse(envsMap{"host": "192.168.0.1"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.1", envs.Host.String())
}

func TestPtrToIP(t *testing.T) {
	var envs struct {
		Host *net.IP
	}

	err := parse(envsMap{"host": "192.168.0.1"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.1", envs.Host.String())
}

func TestIPSlice(t *testing.T) {
	var envs struct {
		Host []net.IP
	}

	err := parse(envsMap{"host": "192.168.0.1,127.0.0.1"}, &envs)
	require.NoError(t, err)
	require.Len(t, envs.Host, 2)
	assert.Equal(t, "192.168.0.1", envs.Host[0].String())
	assert.Equal(t, "127.0.0.1", envs.Host[1].String())
}

func TestInvalidIPAddress(t *testing.T) {
	var envs struct {
		Host net.IP
	}

	err := parse(envsMap{"host": "xxx"}, &envs)
	assert.Error(t, err)
}

func TestMAC(t *testing.T) {
	var envs struct {
		Host net.HardwareAddr
	}

	err := parse(envsMap{"host": "0123.4567.89ab"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "01:23:45:67:89:ab", envs.Host.String())
}

func TestInvalidMac(t *testing.T) {
	var envs struct {
		Host net.HardwareAddr
	}

	err := parse(envsMap{"host": "xxx"}, &envs)
	assert.Error(t, err)
}

func TestMailAddr(t *testing.T) {
	var envs struct {
		Recipient mail.Address
	}

	err := parse(envsMap{"recipient": "foo@example.com"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "<foo@example.com>", envs.Recipient.String())
}

func TestInvalidMailAddr(t *testing.T) {
	var envs struct {
		Recipient mail.Address
	}

	err := parse(envsMap{"recipient": "xxx"}, &envs)
	assert.Error(t, err)
}

type A struct {
	X string
}

type B struct {
	Y int
}

func TestEmbedded(t *testing.T) {
	var envs struct {
		A
		B
		Z bool
	}

	err := parse(envsMap{"x": "hello", "y": "321", "z": "true"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "hello", envs.X)
	assert.Equal(t, 321, envs.Y)
	assert.Equal(t, true, envs.Z)
}

func TestEmbeddedPtr(t *testing.T) {
	// embedded pointer fields are not supported so this should return an error
	var envs struct {
		*A
	}

	err := parse(envsMap{"x": "hello"}, &envs)
	require.Error(t, err)
}

func TestEmbeddedPtrIgnored(t *testing.T) {
	// embedded pointer fields are not normally supported but here
	// we explicitly exclude it so the non-nil embedded structs
	// should work as expected
	var envs struct {
		*A `env:"-"`
		B
	}

	err := parse(envsMap{"y": "321"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, 321, envs.Y)
}

func TestEmbeddedWithDuplicateField(t *testing.T) {
	type T struct {
		A string `env:"cat"`
	}

	type U struct {
		A string `env:"dog"`
	}

	var envs struct {
		T
		U
	}

	err := parse(envsMap{"cat": "cat", "dog": "dog"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "cat", envs.T.A)
	assert.Equal(t, "dog", envs.U.A)
}

func TestEmbeddedWithDuplicateField2(t *testing.T) {
	type T struct {
		A string
	}

	type U struct {
		A string
	}

	var envs struct {
		T
		U
	}

	err := parse(envsMap{"a": "xyz"}, &envs)
	require.NoError(t, err)
	assert.Equal(t, "xyz", envs.T.A)
	assert.Equal(t, "xyz", envs.U.A)
}

func TestReuseParser(t *testing.T) {
	var envs struct {
		Foo string `env:"required"`
	}

	p, err := pparse(envsMap{"foo": "abc"}, &envs)
	require.NoError(t, err)

	err = p.Parse()
	require.NoError(t, err)
	assert.Equal(t, envs.Foo, "abc")

	os.Clearenv()

	err = p.Parse()
	assert.Error(t, err)
}

func TestDefaultOptionValues(t *testing.T) {
	var envs struct {
		A int      `default:"123"`
		B *int     `default:"123"`
		C string   `default:"abc"`
		D *string  `default:"abc"`
		E float64  `default:"1.23"`
		F *float64 `default:"1.23"`
		G bool     `default:"true"`
		H *bool    `default:"true"`
	}

	err := parse(envsMap{"c": "xyz", "e": "4.56"}, &envs)
	require.NoError(t, err)

	assert.Equal(t, 123, envs.A)
	assert.Equal(t, 123, *envs.B)
	assert.Equal(t, "xyz", envs.C)
	assert.Equal(t, "abc", *envs.D)
	assert.Equal(t, 4.56, envs.E)
	assert.Equal(t, 1.23, *envs.F)
	assert.True(t, envs.G)
	assert.True(t, envs.G)
}

func TestDefaultUnparseable(t *testing.T) {
	var envs struct {
		A int `default:"x"`
	}

	err := parse(envsMap{}, &envs)
	assert.EqualError(t, err, `error processing default value for a: strconv.ParseInt: parsing "x": invalid syntax`)
}

func TestDefaultValuesNotAllowedWithRequired(t *testing.T) {
	var envs struct {
		A int `env:"required" default:"123"` // required not allowed with default!
	}

	err := parse(envsMap{}, &envs)
	assert.EqualError(t, err, ".A: 'required' cannot be used when a default value is specified")
}

func TestDefaultValuesNotAllowedWithSlice(t *testing.T) {
	var envs struct {
		A []int `default:"123"` // required not allowed with default!
	}

	err := parse(envsMap{}, &envs)
	assert.EqualError(t, err, ".A: default values are not supported for slice fields")
}

func TestMultipleOptions(t *testing.T) {
	var envs struct {
		A string `env:"a, required"`
	}

	err := parse(envsMap{"a": "b"}, &envs)
	assert.NoError(t, err)
	err = parse(envsMap{}, &envs)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrorFieldIsRequired))
	assert.EqualError(t, err, "a: field is required")
}
