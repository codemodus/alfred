package alfred

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"
)

// Alfred ...
type Alfred struct {
	fs  http.Handler
	cnf *responseConfig
}

// New ...
func New(dir string) *Alfred {
	return &Alfred{
		fs: http.FileServer(http.Dir(dir)),
		cnf: &responseConfig{
			dir:      dir,
			index:    "index.html",
			notFound: []byte("not found"),
		},
	}
}

func (a *Alfred) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rww := newResponseWriterWrap(w, r, a.cnf)

	a.fs.ServeHTTP(rww, r)
}

type responseConfig struct {
	dir      string
	index    string
	notFound []byte
}

type responseWriterWrap struct {
	http.ResponseWriter
	r   *http.Request
	cnf *responseConfig
	hit bool
}

func newResponseWriterWrap(w http.ResponseWriter, r *http.Request, cnf *responseConfig) *responseWriterWrap {
	return &responseWriterWrap{
		ResponseWriter: w,
		r:              r,
		cnf:            cnf,
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

	f, err := os.Open(path.Join(w.cnf.dir, w.cnf.index))
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write(w.cnf.notFound)
	}

	st, err := f.Stat()
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write(w.cnf.notFound)
	}

	w.ResponseWriter.Header().Set("Content-Type", "text/html")
	http.ServeContent(w.ResponseWriter, w.r, f.Name(), st.ModTime(), f)

	return int(st.Size()), nil
}

// LogAccess ...
func LogAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		next.ServeHTTP(w, r)

		dur := float64(time.Since(t).Nanoseconds()) / float64(1000000)

		afmt := "req(method): %s, req(path): %s, dur(ms): %f\n"
		fmt.Printf(afmt, r.Method, r.URL.Path, dur)
	})
}
