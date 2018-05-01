package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/buger/jsonparser"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		printUsage()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "LOG OPEN ERROR:", err.Error())
		os.Exit(1)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	buf := bytes.NewBuffer(nil)

	for {
		line, pfx, err := r.ReadLine()
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, "LOG READ ERROR:", err.Error())
			os.Exit(1)
		}
		if err == io.EOF {
			break
		}

		buf.Write(line)
		if !pfx {
			if buf.Len() > 0 {
				e := new(entry)
				err = jsonparser.ObjectEach(buf.Bytes(), e.parse)
				if err != nil {
					fmt.Fprintln(os.Stderr, "LOG PARSE ERROR:", err.Error())
					os.Exit(1)
				}
				e.print(os.Stdout)
			}
			buf.Reset()
		}
	}
}

func printUsage() {
	flag.CommandLine.SetOutput(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage: zzz [options] LOG_FILE")
	flag.PrintDefaults()
}
