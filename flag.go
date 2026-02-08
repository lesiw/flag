package flag

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

var errHelp = errors.New("help requested")

type Value interface {
	String() string
	Set(string) error
}

type boolValue bool

func newBoolValue(p *bool) *boolValue {
	return (*boolValue)(p)
}
func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = fmt.Errorf("error parsing bool: %s", s)
	}
	*b = boolValue(v)
	return err
}
func (b *boolValue) Get() bool        { return bool(*b) }
func (b *boolValue) String() string   { return strconv.FormatBool(bool(*b)) }
func (b *boolValue) IsBoolFlag() bool { return true }

type stringValue string

func newStringValue(p *string) *stringValue {
	return (*stringValue)(p)
}
func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}
func (s *stringValue) Get() string    { return string(*s) }
func (s *stringValue) String() string { return string(*s) }

type stringsValue []string

func newStringsValue(p *[]string) *stringsValue {
	return (*stringsValue)(p)
}
func (s *stringsValue) Set(v string) error {
	*s = append(*s, v)
	return nil
}
func (s *stringsValue) Get() []string { return *s }
func (s *stringsValue) String() string {
	return "[" + strings.Join(*s, ", ") + "]"
}

type intValue int

func newIntValue(p *int) *intValue {
	return (*intValue)(p)
}
func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = fmt.Errorf("error parsing int: %s", s)
	}
	*i = intValue(v)
	return err
}
func (i *intValue) Get() int       { return int(*i) }
func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

type Flag struct {
	Names []string
	Usage string
	Value Value
	isSet bool
}

func (f *Flag) set(s string) error {
	f.isSet = true
	return f.Value.Set(s)
}

type Set struct {
	args   []string
	output io.Writer
	flags  map[string]*Flag
	usage  string
	Args   []string
}

func NewSet(output io.Writer, usage string) *Set {
	return &Set{
		output: output,
		usage:  usage,
		flags:  make(map[string]*Flag),
	}
}

func (s *Set) Var(value Value, name, usage string) {
	names := strings.Split(name, ",")
	if len(names) < 1 {
		panic("tried to create flag with no name")
	}
	flag := &Flag{names, usage, value, false}
	for _, name := range names {
		s.flags[name] = flag
	}
}

func (s *Set) Bool(name, usage string) *bool {
	var b bool
	s.BoolVar(&b, name, usage)
	return &b
}

func (s *Set) BoolVar(p *bool, name, usage string) {
	s.Var(newBoolValue(p), name, usage)
}

func (s *Set) String(name, usage string) *string {
	var str string
	s.StringVar(&str, name, usage)
	return &str
}

func (s *Set) StringVar(p *string, name, usage string) {
	s.Var(newStringValue(p), name, usage)
}

func (s *Set) Strings(name, usage string) *[]string {
	var strs []string
	s.StringsVar(&strs, name, usage)
	return &strs
}

func (s *Set) StringsVar(p *[]string, name, usage string) {
	s.Var(newStringsValue(p), name, usage)
}

func (s *Set) Int(name, usage string) *int {
	var i int
	s.IntVar(&i, name, usage)
	return &i
}

func (s *Set) IntVar(p *int, name, usage string) {
	s.Var(newIntValue(p), name, usage)
}

func (s *Set) Parse(args ...string) (err error) {
	defer func() {
		if err == nil {
			return
		} else if err != errHelp {
			_, _ = fmt.Fprintln(s.output, err)
		}
		s.PrintUsage()
	}()
	s.args = args
	for len(s.args) > 0 {
		if s.args[0] == "--" {
			s.args = s.args[1:]
			s.Args = append(s.Args, s.args...)
			return
		} else if s.args[0] == "-" {
			s.Args = append(s.Args, "-")
			s.args = s.args[1:]
		} else if len(s.args[0]) > 0 && s.args[0][0] == '-' {
			if err = s.parseFlag(); err != nil {
				return
			}
		} else {
			s.Args = append(s.Args, s.args[0])
			s.args = s.args[1:]
		}
	}
	return
}

func (s *Set) Set(name string) bool {
	if s.flags[name] != nil {
		return s.flags[name].isSet
	}
	return false
}

func (s *Set) parseFlag() error {
	arg := s.args[0]
	s.args = s.args[1:]
	if arg[0] != '-' {
		return fmt.Errorf("bad flag: %s", arg)
	} else if len(arg) > 2 && arg[:2] == "--" {
		return s.parseLongFlag(arg[2:])
	} else {
		return s.parseShortFlag(arg[1:])
	}
}

func (s *Set) parseLongFlag(arg string) error {
	name, val, _ := strings.Cut(arg, "=")
	if len(name) == 1 {
		return fmt.Errorf("bad flag: --%s", name) // Short flags are invalid.
	}
	if name == "help" {
		return errHelp
	}
	flag, ok := s.flags[name]
	if !ok {
		return fmt.Errorf("bad flag: --%s", name)
	}
	_, bool := flag.Value.(*boolValue)
	if val == "" && len(s.args) > 0 && !bool {
		val = s.args[0]
		s.args = s.args[1:]
	} else if val == "" && bool {
		val = "true"
	} else if val == "" && !bool {
		return fmt.Errorf("bad flag: needs value: --%s", name)
	}
	return flag.set(val)
}

func (s *Set) parseShortFlag(arg string) error {
	for len(arg) > 0 {
		name := arg[0]
		arg = arg[1:]
		flag, ok := s.flags[string(name)]
		if !ok {
			return fmt.Errorf("bad flag: '%s'", string(name))
		} else if _, bool := flag.Value.(*boolValue); bool {
			if err := flag.set("true"); err != nil {
				return err
			}
		} else if len(arg) > 0 {
			return flag.set(arg)
		} else if len(s.args) > 0 {
			arg = s.args[0]
			s.args = s.args[1:]
			return flag.set(arg)
		} else {
			return fmt.Errorf("bad flag: '%s' needs value", string(name))
		}
	}
	return nil
}

func (s *Set) PrintError(e string) {
	_, _ = fmt.Fprintln(s.output, e)
	s.PrintUsage()
}

func (s *Set) PrintUsage() {
	_, _ = fmt.Fprintln(s.output, "Usage:", s.usage)
	defaults := s.Defaults()
	if defaults != "" {
		_, _ = fmt.Fprintln(s.output)
		_, _ = fmt.Fprintln(s.output, defaults)
	}
}

func (s *Set) Defaults() string {
	var b strings.Builder
	visited := make(map[*Flag]bool)
	s.Visit(func(flag *Flag) {
		if visited[flag] {
			return
		}
		visited[flag] = true
		for i, name := range flag.Names {
			if len(name) > 1 && i == 0 {
				fmt.Fprintf(&b, "  --%s", name)
			} else if i == 0 {
				fmt.Fprintf(&b, "  -%s", name)
			} else if len(name) > 1 {
				fmt.Fprintf(&b, ",--%s", name)
			} else {
				fmt.Fprintf(&b, ",-%s", name)
			}
		}
		name, usage := unquoteUsage(flag)
		if len(name) > 0 {
			fmt.Fprintf(&b, " %s", name)
		}
		if b.Len() <= 4 {
			b.WriteString("\t")
		} else {
			b.WriteString("\n    \t")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))
		b.WriteString("\n")
	})
	return strings.TrimSuffix(b.String(), "\n")
}

func (s *Set) Visit(fn func(*Flag)) {
	for _, flag := range sortFlags(s.flags) {
		fn(flag)
	}
}

func (s *Set) Arg(i int) string {
	if i < 0 || i >= len(s.Args) {
		return ""
	}
	return s.Args[i]
}

func sortFlags(flags map[string]*Flag) []*Flag {
	result := make([]*Flag, len(flags))
	i := 0
	for _, f := range flags {
		result[i] = f
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Names[0] < result[j].Names[0]
	})
	return result
}

func unquoteUsage(flag *Flag) (name, usage string) {
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break
		}
	}
	name = "value"
	switch flag.Value.(type) {
	case *boolValue:
		name = ""
	case *intValue:
		name = "num"
	case *stringValue:
		name = "string"
	case *stringsValue:
		name = "string[,string...]"
	}
	return
}
