package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/codemodus/alfred"
)

func main() {
	var (
		dir    = "."
		port   = ":4001"
		acs    bool
		silent bool
	)

	flag.StringVar(&dir, "dir", dir, "directory")
	flag.StringVar(&port, "port", port, "http port")
	flag.BoolVar(&acs, "acs", acs, "log access")
	flag.BoolVar(&silent, "s", silent, "silent")
	flag.Parse()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var h http.Handler = alfred.New(absDir)

	if acs && !silent {
		h = alfred.LogAccess(h)
	}

	if !silent {
		fmt.Printf("listening on %s and serving %s\n", port, dir)
	}

	if err := http.ListenAndServe(port, h); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
