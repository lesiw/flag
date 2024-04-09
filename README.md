# lesiw.io/flag

[![Go Reference](https://pkg.go.dev/badge/lesiw.io/flag.svg)](https://pkg.go.dev/lesiw.io/flag)

A flag parser for the lazy. Follows the POSIX [utility conventions][utilconv]
plus GNU `--longopts`.

So, for example, `-abcd` could be four boolean flags, or `-b` could be a string
flag that considers `cd` to be the argument passed to it.

Pass in an `io.Writer` when creating a `FlagSet` and the `Parse` method will
write errors and usage help messages directly to that writer. If `Parse` returns
an error, it is recommended that the program author exit as soon as possible
without additional output.

Aliases can be made at flag definition time by passing in comma-separated flag
names.

## Example

``` go
package main

import (
    "fmt"
    "os"

    "lesiw.io/flag"
)

func main() {
    os.Exit(run())
}

func run() int {
    flags := flag.NewFlagSet(os.Stderr, "sandbox")

    // By default, the usage string will be "Usage of UTILNAME:"
    flags.Usage = "Usage: sandbox [-w WORD] [-n NUM] [-b] ARGS..."

    var (
        word = flags.String("w,word", "a string")
        num  = flags.Int("n,num", "an int")
        bool = flags.Bool("b,bool", "a bool")
    )

    // Replace os.Args with your own strings to test.
    if err := flags.Parse(os.Args[1:]...); err != nil {
        return 1
    }
    if len(flags.Args) < 1 {
        flags.PrintError("at least one arg is required")
        return 1
    }

    fmt.Println("word:", *word)
    fmt.Println("num:", *num)
    fmt.Println("bool:", *bool)
    fmt.Println("args:", flags.Args)

    return 0
}
```

[▶️ Run this example on the Go Playground](https://go.dev/play/p/zvTvgDYN-RP)

[utilconv]: https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html
