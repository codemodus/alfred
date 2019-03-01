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
	dir string
	fs  http.Handler
}

// New ...
func New(dir string) *Alfred {
	return &Alfred{
		dir: dir,
		fs:  http.FileServer(http.Dir(dir)),
	}
}

func (n *Alfred) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rww := newResponseWriterWrap(w, r, n.dir)

	n.fs.ServeHTTP(rww, r)
}

type responseWriterWrap struct {
	http.ResponseWriter
	r     *http.Request
	dir   string
	hit   bool
	index string
	nfBs  []byte
}

func newResponseWriterWrap(w http.ResponseWriter, r *http.Request, dir string) *responseWriterWrap {
	return &responseWriterWrap{
		ResponseWriter: w,
		r:              r,
		dir:            dir,
		index:          "index.html",
		nfBs:           []byte("not found"),
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

	f, err := os.Open(path.Join(w.dir, w.index))
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write(w.nfBs)
	}

	st, err := f.Stat()
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write(w.nfBs)
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
