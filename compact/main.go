package main

import (
    "github.com/ddliu/go-compact"
    // "github.com/spf13/cobra"
    "os"
    "log"
    "io/ioutil"
    "net/url"
    "fmt"
    "flag"
)

func main() {
    var verbose bool

    flag.BoolVar(&verbose, "verbose", false, "verbose")
    flag.BoolVar(&verbose, "v", false, "verbose")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, 
`Compact v0.1.0
Usage: compact [flags] source [target]
Example: compact -v http://google.com google.html
`)
        flag.PrintDefaults()
    }

    flag.Parse()

    if !verbose {
        log.SetOutput(ioutil.Discard)
    }

    args := flag.Args()

    if len(args) < 1 || len(args) > 2 {
        flag.Usage()
        return
    }

    u, err := url.Parse(args[0])

    if err != nil {
        panic(err)
    }

    converter := &compact.Converter{}

    content, err := converter.Convert(u)

    if err != nil {
        panic(err)
    }

    if len(args) == 1 {
        fmt.Print(content)
    } else {
        err = ioutil.WriteFile(args[1], []byte(content), 0666)
        if err != nil {
            panic(err)
        } else {
            fmt.Println("Done")
        }
    }
}