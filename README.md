<h1 align="center">
  <br>
  go-env
  </br>
</h1>
<h4 align="center">Struct-based environment parsing for Go</h4>
<br>

Declare command line envuments for your program by defining a struct.

```go
var envs struct {
	Foo string
	Bar bool
}
env.MustParse(&envs)
fmt.Println(envs.Foo, envs.Bar)
```

```shell
$ foo=hello bar=true ./example
hello true
```

### Installation

```shell
go get github.com/Alex616/go-env
```

### Required environments

```go
var envs struct {
	ID      int `env:"required"`
	Timeout time.Duration
}
env.MustParse(&envs)
```

```shell
$ ./example
error: id is required
```

### Help strings
```go
var envs struct {
	Input    string
	Output   []string
	Dataset  string   `help:"dataset to use"`
	Optimize int      `help:"optimization level"`
}
p, _ := env.MustParse(&envs)
fmt.Fprintln(p.Help())
```

```shell
$ ./example
Environments:
  dataset    dataset to use
  optimize   optimization level
```

### Default values

```go
var envs struct {
	Foo string `default:"abc"`
	Bar bool
}
env.MustParse(&envs)
```

```go
var envs struct {
	Foo string
	Bar bool
}
env.Foo = "abc"
env.MustParse(&envs)
```

### Environments with multiple values
```go
var envs struct {
	Database string
	IDs      []int64
}
env.MustParse(&envs)
fmt.Printf("Fetching the following IDs from %s: %q", envs.Database, envs.IDs)
```

```shell
database=foo ids=1,2,3 ./example
Fetching the following IDs from foo: [1 2 3]
```


### Overriding option names

```go
var envs struct {
	Short         string  `env:"s"`
	Long          string  `env:"custom-long-option"`
	ShortAndLong  string  `env:"my-option"`
}
env.MustParse(&envs)
```

```shell
$ ./example

Environments:
  short
  custom-long-option
  my-option
```


### Embedded structs

The fields of embedded structs are treated just like regular fields:

```go

type DatabaseOptions struct {
	Host     string
	Username string
	Password string
}

type LogOptions struct {
	LogFile string
	Verbose bool
}

func main() {
	var envs struct {
		DatabaseOptions
		LogOptions
	}
	env.MustParse(&envs)
}
```

As usual, any field tagged with `env:"-"` is ignored.

### Custom parsing

Implement `encoding.TextUnmarshaler` to define your own parsing logic.

```go
// Accepts command line envuments of the form "head.tail"
type NameDotName struct {
	Head, Tail string
}

func (n *NameDotName) UnmarshalText(b []byte) error {
	s := string(b)
	pos := strings.Index(s, ".")
	if pos == -1 {
		return fmt.Errorf("missing period in %s", s)
	}
	n.Head = s[:pos]
	n.Tail = s[pos+1:]
	return nil
}

func main() {
	var envs struct {
		Name NameDotName
	}
	env.MustParse(&envs)
	fmt.Printf("%#v\n", envs.Name)
}
```
```shell
$ name=foo.bar ./example
main.NameDotName{Head:"foo", Tail:"bar"}

$ name=oops ./example
error: error processing name: missing period in "oops"
```

### Custom parsing with default values

Implement `encoding.TextMarshaler` to define your own default value strings:

```go
// Accepts command line envuments of the form "head.tail"
type NameDotName struct {
	Head, Tail string
}

func (n *NameDotName) UnmarshalText(b []byte) error {
	// same as previous example
}

// this is only needed if you want to display a default value in the usage string
func (n *NameDotName) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s.%s", n.Head, n.Tail)), nil
}

func main() {
	var envs struct {
		Name NameDotName `default:"file.txt"`
	}
	env.MustParse(&envs)
	fmt.Printf("%#v\n", envs.Name)
}
```
```shell
$ ./example
main.NameDotName{Head:"file", Tail:"txt"}
```


### Description strings

```go
type envs struct {
	Foo string
}

func (envs) Description() string {
	return "this program does this and that"
}

func main() {
	var envs envs
	p, _ := env.MustParse(&envs)
	fmt.Fprintln(p.Help())
}
```

```shell
$ ./example
this program does this and that

Environments:
  foo FOO
```
