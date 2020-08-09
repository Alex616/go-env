package env

import (
	"encoding"
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strings"

	scalar "github.com/alexflint/go-scalar"
)

// path represents a sequence of steps to find the output location for an
// argument or subcommand in the final destination struct.
type path struct {
	root   int                   // index of the destination struct
	fields []reflect.StructField // sequence of struct fields to traverse

}

// String gets a string representation of the given path.
func (p path) String() string {
	s := "args"
	for _, f := range p.fields {
		s += "." + f.Name
	}

	return s
}

// Child gets a new path representing a child of this path.
func (p path) Child(f *reflect.StructField) path {
	// copy the entire slice of fields to avoid possible slice overwrite
	subfields := make([]reflect.StructField, len(p.fields)+1)
	copy(subfields, p.fields)

	if f != nil {
		subfields[len(subfields)-1] = *f
	}

	return path{
		root:   p.root,
		fields: subfields,
	}
}

// spec represents a command line option.
type spec struct {
	dest       path
	typ        reflect.Type
	name       string
	help       string
	defaultVal string // default value for this option
	multiple   bool
	required   bool
	boolean    bool

	hasDefault bool
}

func (s *spec) SetDefault(def string) {
	s.defaultVal = def
	s.hasDefault = true
}

// MustParse processes command line arguments and exits upon failure.
func MustParse(dest ...interface{}) (*Parser, error) {
	p, err := NewParser(Config{}, dest...)
	if err != nil {
		return nil, err
	}

	err = p.Parse()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Parse processes command line arguments and stores them in dest.
func Parse(dest ...interface{}) error {
	p, err := NewParser(Config{}, dest...)
	if err != nil {
		return err
	}

	return p.Parse()
}

// Config represents configuration options for an argument parser.
type Config struct{}

// Parser represents a set of command line options with destination values.
type Parser struct {
	specs       []*spec
	roots       []reflect.Value
	config      Config
	description string
}

// Described is the interface that the destination struct should implement to
// make a description string appear at the top of the help message.
type Described interface {
	// Description returns the string that will be printed on a line by itself
	// at the top of the help message.
	Description() string
}

type visitorFn func(field reflect.StructField, owner reflect.Type) (bool, error)

// walkFields calls a function for each field of a struct, recursively expanding struct fields.
func walkFields(t reflect.Type, visit visitorFn) error {
	return walkFieldsImpl(t, visit, nil)
}

func walkFieldsImpl(t reflect.Type, visit visitorFn, path []int) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		field.Index = make([]int, len(path)+1)
		copy(field.Index, append(path, i))
		expand, err := visit(field, t)

		if err != nil {
			return err
		}

		if expand {
			var subpath []int
			if field.Anonymous {
				subpath = append(path, i) // nolint:gocritic
			}

			err := walkFieldsImpl(field.Type, visit, subpath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NewParser constructs a parser from a list of destination structs.
func NewParser(config Config, dests ...interface{}) (*Parser, error) {
	// construct a parser
	p := Parser{
		config: config,
		specs:  make([]*spec, 0),
	}

	// make a list of roots
	for _, dest := range dests {
		p.roots = append(p.roots, reflect.ValueOf(dest))
	}

	// process each of the destination values
	for i, dest := range dests {
		t := reflect.TypeOf(dest)

		specs, err := specsFromStruct(path{root: i}, t)
		if err != nil {
			return nil, err
		}

		// add nonzero field values as defaults
		for _, spec := range specs {
			if v := p.val(spec.dest); v.IsValid() && !isZero(v) {
				spec.defaultVal = fmt.Sprintf("%v", v)

				if defaultVal, ok := v.Interface().(encoding.TextMarshaler); ok {
					str, err := defaultVal.MarshalText()
					if err != nil {
						return nil, fmt.Errorf("%v: error marshaling default value to string: %w", spec.dest, err)
					}

					spec.defaultVal = string(str)
				}
			}
		}

		p.specs = append(p.specs, specs...)

		if dest, ok := dest.(Described); ok {
			p.description = dest.Description()
		}
	}

	return &p, nil
}

func specsFromStruct(dest path, t reflect.Type) ([]*spec, error) {
	// commands can only be created from pointers to structs
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%s:%s - %w",
			dest, t.Kind(), ErrorNotPointers)
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s:%s - %w",
			dest, t.Kind(), ErrorNotStruct)
	}

	specs := make([]*spec, 0)

	err := walkFields(t, func(field reflect.StructField, t reflect.Type) (bool, error) {
		sp, expand, err := walker(dest, &field, t)
		if sp != nil {
			specs = append(specs, sp)
		}

		return expand, err
	})

	return specs, err
}

func walker(dest path, field *reflect.StructField, t reflect.Type) (*spec, bool, error) {
	// Check for the ignore switch in the tag
	tag := field.Tag.Get("env")
	if tag == "-" {
		return nil, false, nil
	}

	// If this is an embedded struct then recurse into its fields
	if field.Anonymous && field.Type.Kind() == reflect.Struct {
		return nil, true, nil
	}

	// duplicate the entire path to avoid slice overwrites
	subdest := dest.Child(field)
	sp := &spec{
		dest: subdest,
		name: strings.ToLower(field.Name),
		typ:  field.Type,
	}

	if help, exists := field.Tag.Lookup("help"); exists {
		sp.help = help
	}

	if defaultVal, exists := field.Tag.Lookup("default"); exists {
		sp.SetDefault(defaultVal)
	}

	// Look at the tag
	err := lookAtTag(tag, sp)
	if err != nil {
		return nil, false, fmt.Errorf("%s.%s: %w", t.Name(), field.Name, err)
	}

	var parseable bool
	parseable, sp.boolean, sp.multiple = canParse(field.Type)

	if !parseable {
		return sp, false, fmt.Errorf("%s.%s: %s - %w", t.Name(), field.Name, field.Type.String(), ErrorFieldsAreNotSupported)
	}

	if sp.multiple && sp.hasDefault {
		return sp, false, fmt.Errorf("%s.%s: %w", t.Name(), field.Name, ErrorDefaultValueForSlice)
	}

	// if this was an embedded field then we already returned true up above
	return sp, false, nil
}

// lookAtTag fill spec from tag annotation.
func lookAtTag(tag string, sp *spec) error {
	for _, key := range strings.Split(tag, ",") {
		if key == "" {
			continue
		}

		key = strings.TrimLeft(key, " ")

		var value string
		if pos := strings.Index(key, ":"); pos != -1 {
			value = key[pos+1:]
			key = key[:pos]
		}

		switch {
		case key == "name" && value != "":
			sp.name = value
		case key == "required":
			if sp.hasDefault {
				return ErrorRequiredWithDefault
			}

			sp.required = true
		default:
			return fmt.Errorf("%s-%w", key, ErrorUnrecognizedTag)
		}
	}

	return nil
}

// Parse processes the given command line option, storing the results in the field
// of the structs from which NewParser was constructed.
func (p *Parser) Parse() error {
	return p.process()
}

// process environment vars for the given arguments.
func (p *Parser) captureEnvVars(specs []*spec, wasPresent map[*spec]bool) error {
	for _, spec := range specs {
		value, found := os.LookupEnv(spec.name)
		if !found {
			continue
		}

		if spec.multiple {
			// expect a CSV string in an environment
			// variable in the case of multiple values
			values, err := csv.NewReader(strings.NewReader(value)).Read()
			if err != nil {
				return fmt.Errorf( // nolint:goerr113
					"error reading a CSV string from environment variable %s with multiple values: %w",
					spec.name,
					err,
				)
			}

			if err = setSlice(p.val(spec.dest), values); err != nil {
				return fmt.Errorf(
					"error processing environment variable %s with multiple values: %w",
					spec.name,
					err,
				)
			}
		} else if err := scalar.ParseValue(p.val(spec.dest), value); err != nil {
			return fmt.Errorf("error processing environment variable %s: %w", spec.name, err)
		}

		wasPresent[spec] = true
	}

	return nil
}

// process goes through arguments one-by-one, parses them, and assigns the result to
// the underlying struct field.
func (p *Parser) process() error {
	// track the options we have seen
	wasPresent := make(map[*spec]bool)

	// make a copy of the specs because we will add to this list each time we expand a subcommand
	specs := make([]*spec, len(p.specs))
	copy(specs, p.specs)

	// deal with environment vars
	err := p.captureEnvVars(specs, wasPresent)
	if err != nil {
		return err
	}

	// fill in defaults and check that all the required args were provided
	for _, spec := range specs {
		if wasPresent[spec] {
			continue
		}

		name := spec.name

		if spec.required {
			return fmt.Errorf("%s: %w", name, ErrorFieldIsRequired)
		}

		if spec.defaultVal != "" {
			err := scalar.ParseValue(p.val(spec.dest), spec.defaultVal)
			if err != nil {
				return fmt.Errorf("error processing default value for %s: %w", name, err)
			}
		}
	}

	return nil
}

// val returns a reflect.Value corresponding to the current value for the
// given path.
func (p *Parser) val(dest path) reflect.Value {
	v := p.roots[dest.root]

	for _, field := range dest.fields {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}
			}

			v = v.Elem()
		}

		next := v.FieldByIndex(field.Index)
		if !next.IsValid() {
			// it is appropriate to panic here because this can only happen due to
			// an internal bug in this library (since we construct the path ourselves
			// by reflecting on the same struct)
			panic(fmt.Sprintf("error resolving path %v: %v has no field named %v",
				dest.fields, v.Type(), field))
		}

		v = next
	}

	return v
}

// parse a value as the appropriate type and store it in the struct.
func setSlice(dest reflect.Value, values []string) error {
	if !dest.CanSet() {
		return ErrorFieldIsNotWritable
	}

	var ptr bool

	elem := dest.Type().Elem()
	if elem.Kind() == reflect.Ptr && !elem.Implements(textUnmarshalerType) {
		ptr = true
		elem = elem.Elem()
	}

	// Truncate the dest slice in case default values exist
	if !dest.IsNil() {
		dest.SetLen(0)
	}

	for _, s := range values {
		v := reflect.New(elem)
		if err := scalar.ParseValue(v.Elem(), s); err != nil {
			return err
		}

		if !ptr {
			v = v.Elem()
		}

		dest.Set(reflect.Append(dest, v))
	}

	return nil
}

// isZero returns true if v contains the zero value for its type.
func isZero(v reflect.Value) bool {
	t := v.Type()
	if t.Kind() == reflect.Slice {
		return v.IsNil()
	}

	if !t.Comparable() {
		return false
	}

	return v.Interface() == reflect.Zero(t).Interface()
}
