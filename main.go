package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/codemodus/mixmux"
)

type options struct {
	dir    string
	port   string
	acs    bool
	silent bool
}

type node struct {
	dir string
}

func (n *node) assetsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(n.dir, "assets", r.URL.Path[7:]))
}

func (n *node) iconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(n.dir, "assets", "ico"+r.URL.Path))
}

func (n *node) htmlHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("html", r.URL.Path))
}

func (n *node) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		http.ServeFile(w, r, filepath.Join(n.dir, "html", "index.html"))
		return
	}

	if len(r.URL.Path) >= 5 {
		sfx := r.URL.Path[len(r.URL.Path)-5:]

		if sfx == ".html" || sfx[1:] == ".htm" {
			n.htmlHandler(w, r)
			return
		}

		if sfx[1:] == ".ico" {
			n.iconHandler(w, r)
			return
		}
	}

	http.ServeFile(w, r, filepath.Join(n.dir, "html", r.URL.Path, "index.html"))
}

func (n *node) access(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		next.ServeHTTP(w, r)

		fmt.Printf(
			"req(method): %s, req(path): %s, dur(ms): %f\n",
			r.Method,
			r.URL.Path,
			float64(time.Since(t).Nanoseconds())/float64(1000000),
		)
	})
}

func main() {
	opts := &options{}
	flag.StringVar(&opts.dir, "dir", ".", "directory")
	flag.StringVar(&opts.port, "port", ":4001", "http port")
	flag.BoolVar(&opts.acs, "acs", false, "log access")
	flag.BoolVar(&opts.silent, "s", false, "silent")
	flag.Parse()

	absDir, err := filepath.Abs(opts.dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	n := &node{
		dir: absDir,
	}

	mOpts := &mixmux.Options{
		NotFound: http.HandlerFunc(n.indexHandler),
	}
	m := mixmux.NewRouter(mOpts)

	m.Get("/assets/*x", http.HandlerFunc(n.assetsHandler))

	var h http.Handler = m
	if opts.acs && !opts.silent {
		h = n.access(h)
	}

	if !opts.silent {
		fmt.Printf("listening on %s and serving %s\n", opts.port, n.dir)
	}

	if err := http.ListenAndServe(opts.port, h); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
