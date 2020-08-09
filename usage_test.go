package env_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Alex616/go-env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errTestMissingPeriod = errors.New("missing period")
	errTestProblem       = errors.New("there was a problem")
)

type NameDotName struct {
	Head, Tail string
}

func (n *NameDotName) UnmarshalText(b []byte) error {
	s := string(b)
	pos := strings.Index(s, ".")

	if pos == -1 {
		return fmt.Errorf("%s: %w", s, errTestMissingPeriod)
	}

	n.Head = s[:pos]
	n.Tail = s[pos+1:]

	return nil
}

func (n *NameDotName) MarshalText() (text []byte, err error) {
	text = []byte(fmt.Sprintf("%s.%s", n.Head, n.Tail))

	return
}

func TestWriteUsage(t *testing.T) {
	expectedHelp := `Environments:
  input
  output                 list of outputs
  name                   name to use [default: Foo Bar]
  value                  secret value [default: 42]
  v                      verbosity level
  dataset                dataset to use
  O                      optimization level
  ids                    Ids
  values                 Values [default: [3.14 42 256]]
  WORKERS                number of workers to start [default: 10]
  TEST_ENV
  f                      File with mandatory extension [default: scratch.txt]
`

	var args struct {
		Input    string       `env:""`
		Output   []string     `env:"" help:"list of outputs"`
		Name     string       `help:"name to use"`
		Value    int          `help:"secret value"`
		Verbose  bool         `env:"name:v" help:"verbosity level"`
		Dataset  string       `help:"dataset to use"`
		Optimize int          `env:"name:O" help:"optimization level"`
		Ids      []int64      `help:"Ids"`
		Values   []float64    `help:"Values"`
		Workers  int          `env:"name:WORKERS" help:"number of workers to start" default:"10"`
		TestEnv  string       `env:"name:TEST_ENV"`
		File     *NameDotName `env:"name:f" help:"File with mandatory extension"`
	}

	args.Name = "Foo Bar"
	args.Value = 42
	args.Values = []float64{3.14, 42, 256}
	args.File = &NameDotName{"scratch", "txt"}
	p, err := env.NewParser(env.Config{}, &args)
	require.NoError(t, err)

	help := p.Help()
	assert.Equal(t, expectedHelp, help)
}

type MyEnum int

func (n *MyEnum) UnmarshalText(b []byte) error {
	return nil
}

func (n *MyEnum) MarshalText() ([]byte, error) {
	return nil, errTestProblem
}

func TestUsageWithDefaults(t *testing.T) {
	expectedHelp := `Environments:
  label [default: cat]
  content [default: dog]
`

	var args struct {
		Label   string
		Content string `default:"dog"`
	}

	args.Label = "cat"
	p, err := env.NewParser(env.Config{}, &args)
	require.NoError(t, err)

	args.Label = "should_ignore_this"

	help := p.Help()
	assert.Equal(t, expectedHelp, help)
}

func TestUsageCannotMarshalToString(t *testing.T) {
	var args struct {
		Name *MyEnum
	}

	v := MyEnum(42)
	args.Name = &v
	_, err := env.NewParser(env.Config{}, &args)
	assert.EqualError(t, err, `args.Name: error marshaling default value to string: there was a problem`)
}

type described struct{}

// Described returns the description for this program.
func (described) Description() string {
	return "this program does this and that"
}

func TestUsageWithDescription(t *testing.T) {
	expectedHelp := `this program does this and that
`
	p, err := env.NewParser(env.Config{}, &described{})
	require.NoError(t, err)

	help := p.Help()
	assert.Equal(t, expectedHelp, help)
}
