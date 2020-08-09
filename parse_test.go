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
	var args struct {
		Foo string
		Ptr *string
	}

	err := parse(envsMap{
		"foo": "bar",
		"ptr": "baz",
	}, &args)
	require.NoError(t, err)
	assert.Equal(t, "bar", args.Foo)
	assert.Equal(t, "baz", *args.Ptr)
}

func TestBool(t *testing.T) {
	var args struct {
		A bool
		B bool
		C *bool
		D *bool
	}

	err := parse(envsMap{
		"a": "true",
		"c": "true",
	}, &args)
	require.NoError(t, err)
	assert.True(t, args.A)
	assert.False(t, args.B)
	assert.True(t, *args.C)
	assert.Nil(t, args.D)
}

func TestInt(t *testing.T) {
	var args struct {
		Foo int
		Ptr *int
	}

	err := parse(envsMap{
		"foo": "7",
		"ptr": "8",
	}, &args)
	require.NoError(t, err)
	assert.EqualValues(t, 7, args.Foo)
	assert.EqualValues(t, 8, *args.Ptr)
}

func TestNegativeInt(t *testing.T) {
	var args struct {
		Foo int
	}

	err := parse(envsMap{
		"foo": "-100",
	}, &args)
	require.NoError(t, err)
	assert.EqualValues(t, args.Foo, -100)
}

func TestNegativeIntAndFloatAndTricks(t *testing.T) {
	var args struct {
		Foo int
		Bar float64
		N   int `env:"name:100"`
	}

	err := parse(envsMap{
		"foo": "-100",
		"bar": "-60.14",
		"100": "-100",
	}, &args)
	require.NoError(t, err)
	assert.EqualValues(t, args.Foo, -100)
	assert.EqualValues(t, args.Bar, -60.14)
	assert.EqualValues(t, args.N, -100)
}

func TestUint(t *testing.T) {
	var args struct {
		Foo uint
		Ptr *uint
	}

	err := parse(envsMap{
		"foo": "7",
		"ptr": "8",
	}, &args)
	require.NoError(t, err)
	assert.EqualValues(t, 7, args.Foo)
	assert.EqualValues(t, 8, *args.Ptr)
}

func TestFloat(t *testing.T) {
	var args struct {
		Foo float32
		Ptr *float32
	}

	err := parse(envsMap{
		"foo": "3.4",
		"ptr": "3.5",
	}, &args)
	require.NoError(t, err)
	assert.EqualValues(t, 3.4, args.Foo)
	assert.EqualValues(t, 3.5, *args.Ptr)
}

func TestDuration(t *testing.T) {
	var args struct {
		Foo time.Duration
		Ptr *time.Duration
	}

	err := parse(envsMap{"foo": "3ms", "ptr": "4ms"}, &args)
	require.NoError(t, err)
	assert.Equal(t, 3*time.Millisecond, args.Foo)
	assert.Equal(t, 4*time.Millisecond, *args.Ptr)
}

func TestInvalidDuration(t *testing.T) {
	var args struct {
		Foo time.Duration
	}

	err := parse(envsMap{"foo": "xxx"}, &args)
	require.Error(t, err)
}

func TestIntPtr(t *testing.T) {
	var args struct {
		Foo *int
	}

	err := parse(envsMap{"foo": "123"}, &args)
	require.NoError(t, err)
	require.NotNil(t, args.Foo)
	assert.Equal(t, 123, *args.Foo)
}

func TestIntPtrNotPresent(t *testing.T) {
	var args struct {
		Foo *int
	}

	err := parse(envsMap{}, &args)
	require.NoError(t, err)
	assert.Nil(t, args.Foo)
}

func TestMixed(t *testing.T) {
	var args struct {
		Foo  string `env:"name:f"`
		Bar  int
		Baz  uint `env:""`
		Ham  bool
		Spam float32
	}

	args.Bar = 3
	err := parse(envsMap{"baz": "123", "spam": "1.2", "ham": "true", "f": "xyz"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "xyz", args.Foo)
	assert.Equal(t, 3, args.Bar)
	assert.Equal(t, uint(123), args.Baz)
	assert.Equal(t, true, args.Ham)
	assert.EqualValues(t, 1.2, args.Spam)
}

func TestRequired(t *testing.T) {
	var args struct {
		Foo string `env:"required"`
	}

	err := parse(envsMap{}, &args)
	require.Error(t, err, "--foo is required")
}

func TestLongFlag(t *testing.T) {
	var args struct {
		Foo string `env:"name:abc"`
	}

	err := parse(envsMap{"abc": "xyz"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "xyz", args.Foo)
}

func TestCaseSensitive(t *testing.T) {
	var args struct {
		Lower bool `env:"name:v"`
		Upper bool `env:"name:V"`
	}

	err := parse(envsMap{"v": "true"}, &args)
	require.NoError(t, err)
	assert.True(t, args.Lower)
	assert.False(t, args.Upper)
}

func TestCaseSensitive2(t *testing.T) {
	var args struct {
		Lower bool `env:"name:v"`
		Upper bool `env:"name:V"`
	}

	err := parse(envsMap{"V": "true"}, &args)
	require.NoError(t, err)
	assert.False(t, args.Lower)
	assert.True(t, args.Upper)
}

func TestMultiple(t *testing.T) {
	var args struct {
		Foo []int
		Bar []string
	}

	err := parse(envsMap{"foo": "1,2,3", "bar": "x,y,z"}, &args)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, args.Foo)
	assert.Equal(t, []string{"x", "y", "z"}, args.Bar)
}

func TestMultipleWithEq(t *testing.T) {
	var args struct {
		Foo []int
		Bar []string
	}

	err := parse(envsMap{"foo": "1,2,3", "bar": "x"}, &args)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, args.Foo)
	assert.Equal(t, []string{"x"}, args.Bar)
}

func TestMultipleWithDefault(t *testing.T) {
	var args struct {
		Foo []int
		Bar []string
	}

	args.Foo = []int{42}
	args.Bar = []string{"foo"}
	err := parse(envsMap{"foo": "1,2,3", "bar": "x,y,z"}, &args)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, args.Foo)
	assert.Equal(t, []string{"x", "y", "z"}, args.Bar)
}

func TestExemptField(t *testing.T) {
	var args struct {
		Foo string
		Bar interface{} `env:"-"`
	}

	err := parse(envsMap{"foo": "xyz"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "xyz", args.Foo)
}

func TestMissingRequired(t *testing.T) {
	var args struct {
		Foo string `env:"required"`
		X   string
	}

	err := parse(envsMap{"x": "bar"}, &args)
	assert.Error(t, err)
}

func TestNonsenseKey(t *testing.T) {
	var args struct {
		X string `env:"nonsense"`
	}

	err := parse(envsMap{"x": "bar"}, &args)
	assert.Error(t, err)
}

func TestInvalidInt(t *testing.T) {
	var args struct {
		Foo int
	}

	err := parse(envsMap{"foo": "xyz"}, &args)
	assert.Error(t, err)
}

func TestInvalidUint(t *testing.T) {
	var args struct {
		Foo uint
	}

	err := parse(envsMap{"foo": "xyz"}, &args)
	assert.Error(t, err)
}

func TestInvalidFloat(t *testing.T) {
	var args struct {
		Foo float64
	}

	err := parse(envsMap{"foo": "xyz"}, &args)
	require.Error(t, err)
}

func TestInvalidBool(t *testing.T) {
	var args struct {
		Foo bool
	}

	err := parse(envsMap{"foo": "xyz"}, &args)
	require.Error(t, err)
}

func TestInvalidIntSlice(t *testing.T) {
	var args struct {
		Foo []int
	}

	err := parse(envsMap{"foo": "1 2 xyz"}, &args)
	require.Error(t, err)
}

func TestErrorOnNonPointer(t *testing.T) {
	var args struct{}
	err := parse(envsMap{}, args)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrorNotPointers))
}

func TestErrorOnNonStruct(t *testing.T) {
	var args string
	err := parse(envsMap{}, &args)
	assert.Error(t, err)
}

func TestUnsupportedType(t *testing.T) {
	var args struct {
		Foo interface{}
	}

	err := parse(envsMap{"foo": ""}, &args)
	assert.Error(t, err)
}

func TestUnsupportedSliceElement(t *testing.T) {
	var args struct {
		Foo []interface{}
	}

	err := parse(envsMap{"foo": "3"}, &args)
	assert.Error(t, err)
}

func TestUnknownTag(t *testing.T) {
	var args struct {
		Foo string `env:"this_is_not_valid"`
	}

	err := parse(envsMap{"foo": "xyz"}, &args)
	assert.Error(t, err)
}

func TestParse(t *testing.T) {
	var args struct {
		Foo string
	}

	err := parse(envsMap{"foo": "bar"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "bar", args.Foo)
}

func TestParseError(t *testing.T) {
	var args struct {
		Foo string `env:"this_is_not_valid"`
	}

	err := Parse(&args)
	assert.Error(t, err)
}

func TestMustParse(t *testing.T) {
	var args struct {
		Foo string
	}

	_ = os.Setenv("foo", "bar")
	parser, err := MustParse(&args)
	require.NoError(t, err)
	assert.Equal(t, "bar", args.Foo)
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
	var args struct {
		Foo textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc"}, &args)
	require.NoError(t, err)
	assert.Equal(t, 3, args.Foo.val)
}

func TestPtrToTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var args struct {
		Foo *textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc"}, &args)
	require.NoError(t, err)
	assert.Equal(t, 3, args.Foo.val)
}

func TestRepeatedTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var args struct {
		Foo []textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc,d,ef"}, &args)
	require.NoError(t, err)
	require.Len(t, args.Foo, 3)
	assert.Equal(t, 3, args.Foo[0].val)
	assert.Equal(t, 1, args.Foo[1].val)
	assert.Equal(t, 2, args.Foo[2].val)
}

func TestRepeatedPtrToTextUnmarshaler(t *testing.T) {
	// fields that implement TextUnmarshaler should be parsed using that interface
	var args struct {
		Foo []*textUnmarshaler
	}

	err := parse(envsMap{"foo": "abc,d,ef"}, &args)
	require.NoError(t, err)
	require.Len(t, args.Foo, 3)
	assert.Equal(t, 3, args.Foo[0].val)
	assert.Equal(t, 1, args.Foo[1].val)
	assert.Equal(t, 2, args.Foo[2].val)
}

type boolUnmarshaler bool

func (p *boolUnmarshaler) UnmarshalText(b []byte) error {
	*p = len(b)%2 == 0

	return nil
}

func TestBoolUnmarhsaler(t *testing.T) {
	// test that a bool type that implements TextUnmarshaler is
	// handled as a TextUnmarshaler not as a bool
	var args struct {
		Foo *boolUnmarshaler
	}

	err := parse(envsMap{"foo": "ab"}, &args)
	require.NoError(t, err)
	assert.EqualValues(t, true, *args.Foo)
}

type sliceUnmarshaler []int

func (p *sliceUnmarshaler) UnmarshalText(b []byte) error {
	*p = sliceUnmarshaler{len(b)}

	return nil
}

func TestSliceUnmarhsaler(t *testing.T) {
	// test that a slice type that implements TextUnmarshaler is
	// handled as a TextUnmarshaler not as a slice
	var args struct {
		Foo *sliceUnmarshaler
		Bar string `env:""`
	}

	err := parse(envsMap{"foo": "abcde"}, &args)
	require.NoError(t, err)
	require.Len(t, *args.Foo, 1)
	assert.EqualValues(t, 5, (*args.Foo)[0])
	assert.Equal(t, "", args.Bar)
}

func TestIP(t *testing.T) {
	var args struct {
		Host net.IP
	}

	err := parse(envsMap{"host": "192.168.0.1"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.1", args.Host.String())
}

func TestPtrToIP(t *testing.T) {
	var args struct {
		Host *net.IP
	}

	err := parse(envsMap{"host": "192.168.0.1"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.1", args.Host.String())
}

func TestIPSlice(t *testing.T) {
	var args struct {
		Host []net.IP
	}

	err := parse(envsMap{"host": "192.168.0.1,127.0.0.1"}, &args)
	require.NoError(t, err)
	require.Len(t, args.Host, 2)
	assert.Equal(t, "192.168.0.1", args.Host[0].String())
	assert.Equal(t, "127.0.0.1", args.Host[1].String())
}

func TestInvalidIPAddress(t *testing.T) {
	var args struct {
		Host net.IP
	}

	err := parse(envsMap{"host": "xxx"}, &args)
	assert.Error(t, err)
}

func TestMAC(t *testing.T) {
	var args struct {
		Host net.HardwareAddr
	}

	err := parse(envsMap{"host": "0123.4567.89ab"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "01:23:45:67:89:ab", args.Host.String())
}

func TestInvalidMac(t *testing.T) {
	var args struct {
		Host net.HardwareAddr
	}

	err := parse(envsMap{"host": "xxx"}, &args)
	assert.Error(t, err)
}

func TestMailAddr(t *testing.T) {
	var args struct {
		Recipient mail.Address
	}

	err := parse(envsMap{"recipient": "foo@example.com"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "<foo@example.com>", args.Recipient.String())
}

func TestInvalidMailAddr(t *testing.T) {
	var args struct {
		Recipient mail.Address
	}

	err := parse(envsMap{"recipient": "xxx"}, &args)
	assert.Error(t, err)
}

type A struct {
	X string
}

type B struct {
	Y int
}

func TestEmbedded(t *testing.T) {
	var args struct {
		A
		B
		Z bool
	}

	err := parse(envsMap{"x": "hello", "y": "321", "z": "true"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "hello", args.X)
	assert.Equal(t, 321, args.Y)
	assert.Equal(t, true, args.Z)
}

func TestEmbeddedPtr(t *testing.T) {
	// embedded pointer fields are not supported so this should return an error
	var args struct {
		*A
	}

	err := parse(envsMap{"x": "hello"}, &args)
	require.Error(t, err)
}

func TestEmbeddedPtrIgnored(t *testing.T) {
	// embedded pointer fields are not normally supported but here
	// we explicitly exclude it so the non-nil embedded structs
	// should work as expected
	var args struct {
		*A `env:"-"`
		B
	}

	err := parse(envsMap{"y": "321"}, &args)
	require.NoError(t, err)
	assert.Equal(t, 321, args.Y)
}

func TestEmbeddedWithDuplicateField(t *testing.T) {
	// see https://github.com/alexflint/go-arg/issues/100
	type T struct {
		A string `env:"name:cat"`
	}

	type U struct {
		A string `env:"name:dog"`
	}

	var args struct {
		T
		U
	}

	err := parse(envsMap{"cat": "cat", "dog": "dog"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "cat", args.T.A)
	assert.Equal(t, "dog", args.U.A)
}

func TestEmbeddedWithDuplicateField2(t *testing.T) {
	type T struct {
		A string
	}

	type U struct {
		A string
	}

	var args struct {
		T
		U
	}

	err := parse(envsMap{"a": "xyz"}, &args)
	require.NoError(t, err)
	assert.Equal(t, "xyz", args.T.A)
	assert.Equal(t, "xyz", args.U.A)
}

func TestReuseParser(t *testing.T) {
	var args struct {
		Foo string `env:"required"`
	}

	p, err := pparse(envsMap{"foo": "abc"}, &args)
	require.NoError(t, err)

	err = p.Parse()
	require.NoError(t, err)
	assert.Equal(t, args.Foo, "abc")

	os.Clearenv()

	err = p.Parse()
	assert.Error(t, err)
}

func TestDefaultOptionValues(t *testing.T) {
	var args struct {
		A int      `default:"123"`
		B *int     `default:"123"`
		C string   `default:"abc"`
		D *string  `default:"abc"`
		E float64  `default:"1.23"`
		F *float64 `default:"1.23"`
		G bool     `default:"true"`
		H *bool    `default:"true"`
	}

	err := parse(envsMap{"c": "xyz", "e": "4.56"}, &args)
	require.NoError(t, err)

	assert.Equal(t, 123, args.A)
	assert.Equal(t, 123, *args.B)
	assert.Equal(t, "xyz", args.C)
	assert.Equal(t, "abc", *args.D)
	assert.Equal(t, 4.56, args.E)
	assert.Equal(t, 1.23, *args.F)
	assert.True(t, args.G)
	assert.True(t, args.G)
}

func TestDefaultUnparseable(t *testing.T) {
	var args struct {
		A int `default:"x"`
	}

	err := parse(envsMap{}, &args)
	assert.EqualError(t, err, `error processing default value for a: strconv.ParseInt: parsing "x": invalid syntax`)
}

func TestDefaultValuesNotAllowedWithRequired(t *testing.T) {
	var args struct {
		A int `env:"required" default:"123"` // required not allowed with default!
	}

	err := parse(envsMap{}, &args)
	assert.EqualError(t, err, ".A: 'required' cannot be used when a default value is specified")
}

func TestDefaultValuesNotAllowedWithSlice(t *testing.T) {
	var args struct {
		A []int `default:"123"` // required not allowed with default!
	}

	err := parse(envsMap{}, &args)
	assert.EqualError(t, err, ".A: default values are not supported for slice fields")
}
