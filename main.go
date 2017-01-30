package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

type options struct {
	dir    string
	port   string
	acs    bool
	silent bool
}

type responseWriterWrap struct {
	http.ResponseWriter
	r   *http.Request
	dir string
	hit bool
}

func newResponseWriterWrap(w http.ResponseWriter, r *http.Request, dir string) *responseWriterWrap {
	return &responseWriterWrap{
		ResponseWriter: w,
		r:              r,
		dir:            dir,
	}
}

func (w *responseWriterWrap) WriteHeader(code int) {
	if code == http.StatusNotFound {
		w.hit = true
		return
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrap) Write(b []byte) (int, error) {
	if !w.hit {
		return w.ResponseWriter.Write(b)
	}

	f, err := os.Open(path.Join(w.dir, "index.html"))
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write([]byte("not found"))
	}

	st, err := f.Stat()
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write([]byte("not found"))
	}

	w.ResponseWriter.Header().Set("Content-Type", "text/html")
	http.ServeContent(w.ResponseWriter, w.r, f.Name(), st.ModTime(), f)

	return int(st.Size()), nil
}

type node struct {
	dir string
	fs  http.Handler
}

func newNode(dir string) *node {
	return &node{
		dir: dir,
		fs:  http.FileServer(http.Dir(dir)),
	}
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

func (n *node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rww := newResponseWriterWrap(w, r, n.dir)
	n.fs.ServeHTTP(rww, r)
}

func main() {
	opts := &options{}

	flag.StringVar(
		&opts.dir, "dir", ".",
		"directory",
	)
	flag.StringVar(
		&opts.port, "port", ":4001",
		"http port",
	)
	flag.BoolVar(
		&opts.acs, "acs", false,
		"log access",
	)
	flag.BoolVar(
		&opts.silent, "s", false,
		"silent",
	)
	flag.Parse()

	absDir, err := filepath.Abs(opts.dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	n := newNode(absDir)

	var h http.Handler = n
	if opts.acs && !opts.silent {
		h = n.access(n)
	}

	if !opts.silent {
		fmt.Printf("listening on %s and serving %s\n", opts.port, n.dir)
	}

	if err := http.ListenAndServe(opts.port, h); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
