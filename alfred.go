package alfred

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/codemodus/rwap"
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
	w = newResponseWriterWrap(w, r, a.cnf)

	a.fs.ServeHTTP(w, r)
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

	f, st, err := openFileWithStats(path.Join(w.cnf.dir, w.cnf.index))
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return w.ResponseWriter.Write(w.cnf.notFound)
	}

	w.ResponseWriter.Header().Set("Content-Type", "text/html")
	http.ServeContent(w.ResponseWriter, w.r, f.Name(), st.ModTime(), f)

	return int(st.Size()), nil
}

func openFileWithStats(filename string) (*os.File, os.FileInfo, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	st, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	return f, st, nil
}

// LogAccess ...
func LogAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		rw := rwap.New(w)

		next.ServeHTTP(rw, r)

		var stts int
		rstts := rw.Status()
		if rstts > 0 {
			stts = rstts
		}

		afmt := "%3d\t%-7s\t%08.3f\t%s\n"
		fmt.Printf(afmt, stts, r.Method, floatDurSince(t), r.URL.Path)
	})
}

func floatDurSince(t time.Time) float64 {
	return float64(time.Since(t).Nanoseconds()) / 1000000
}
