package flag_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"lesiw.io/flag"
)

type config struct {
	a []string
	s string
	n int
	x bool
	y bool
	z bool
}

func TestFlag(t *testing.T) {
	tests := []struct {
		args []string
		want config
	}{{
		args: []string{"-x"},
		want: config{[]string{}, "", 0, true, false, false},
	}, {
		args: []string{"--zee"},
		want: config{[]string{}, "", 0, false, false, true},
	}, {
		args: []string{"--zee=true"},
		want: config{[]string{}, "", 0, false, false, true},
	}, {
		args: []string{"--zee=false"},
		want: config{[]string{}, "", 0, false, false, false},
	}, {
		args: []string{"--zee", "false"},
		want: config{[]string{"false"}, "", 0, false, false, true},
	}, {
		args: []string{"-x", "-y"},
		want: config{[]string{}, "", 0, true, true, false},
	}, {
		args: []string{"-xy"},
		want: config{[]string{}, "", 0, true, true, false},
	}, {
		args: []string{"-xs", "foo"},
		want: config{[]string{}, "foo", 0, true, false, false},
	}, {
		args: []string{"-xsfoo"},
		want: config{[]string{}, "foo", 0, true, false, false},
	}, {
		args: []string{"-s", "foo", "bar"},
		want: config{[]string{"bar"}, "foo", 0, false, false, false},
	}, {
		args: []string{"--zee", "foo", "-yxsbar", "baz"},
		want: config{[]string{"foo", "baz"}, "bar", 0, true, true, true},
	}, {
		args: []string{"-x", "--", "-y"},
		want: config{[]string{"-y"}, "", 0, true, false, false},
	}, {
		args: []string{"-n", "42"},
		want: config{[]string{}, "", 42, false, false, false},
	}, {
		args: []string{"-n", "-42"},
		want: config{[]string{}, "", -42, false, false, false},
	}, {
		args: []string{"-n", "0"},
		want: config{[]string{}, "", 0, false, false, false},
	}, {
		args: []string{"-n", "-0"},
		want: config{[]string{}, "", 0, false, false, false},
	}}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			fs := flag.NewFlagSet(new(strings.Builder), "test")
			var s string
			var n int
			var x, y, z bool
			fs.StringVar(&s, "s", "")
			fs.IntVar(&n, "n", "")
			fs.BoolVar(&x, "x", "")
			fs.BoolVar(&y, "y", "")
			fs.BoolVar(&z, "zee", "")
			if err := fs.Parse(tt.args...); err != nil {
				t.Error(err)
			}
			if x != tt.want.x {
				t.Errorf("x: got %v, want %v", x, tt.want.x)
			}
			if y != tt.want.y {
				t.Errorf("y: got %v, want %v", y, tt.want.y)
			}
			if z != tt.want.z {
				t.Errorf("z: got %v, want %v", z, tt.want.z)
			}
			if s != tt.want.s {
				t.Errorf("s: got %v, want %v", s, tt.want.s)
			}
			if n != tt.want.n {
				t.Errorf("n: got %v, want %v", n, tt.want.n)
			}
			for i := range tt.want.a {
				if tt.want.a[i] != fs.Arg(i) {
					t.Errorf("a[%d]: got %v, want %v",
						i, fs.Arg(i), tt.want.a[i])
				}
			}
		})
	}
}

func ExampleFlagSet_Bool() {
	fs := flag.NewFlagSet(os.Stderr, "example")
	bool := fs.Bool("b", "some bool")
	if err := fs.Parse("example", "-b"); err != nil {
		panic(err)
	}
	fmt.Println("bool:", *bool)
	// Output: bool: true
}

func ExampleFlagSet_String() {
	fs := flag.NewFlagSet(os.Stderr, "example")
	word := fs.String("w", "a string")
	if err := fs.Parse("example", "-wfoo"); err != nil {
		panic(err)
	}
	fmt.Println("word:", *word)
	// Output: word: foo
}

func ExampleFlagSet_String_with_bools() {
	fs := flag.NewFlagSet(os.Stderr, "example")
	boola := fs.Bool("a", "some bool")
	boolb := fs.Bool("b", "some other bool")
	word := fs.String("c", "a string")
	if err := fs.Parse("example", "-abcde"); err != nil {
		panic(err)
	}
	fmt.Println("bool 'a':", *boola)
	fmt.Println("bool 'b':", *boolb)
	fmt.Println("word:", *word)
	// Output:
	// bool 'a': true
	// bool 'b': true
	// word: de
}