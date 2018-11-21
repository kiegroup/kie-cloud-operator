package packd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MemoryBox(t *testing.T) {
	r := require.New(t)

	box := NewMemoryBox()

	r.False(box.Has("a/a.txt"))
	r.NoError(box.AddString("b/b.txt", "B"))
	r.NoError(box.AddBytes("a/a.txt", []byte("A")))
	r.True(box.Has("a/a.txt"))

	b, err := box.Find("b/b.txt")
	r.NoError(err)
	r.Equal([]byte("B"), b)

	s, err := box.FindString("a/a.txt")
	r.NoError(err)
	r.Equal("A", s)

	r.Equal([]string{"a/a.txt", "b/b.txt"}, box.List())

	wm := map[string]string{}
	box.Walk(func(path string, file File) error {
		bb := &bytes.Buffer{}
		io.Copy(bb, file)
		wm[path] = bb.String()
		return nil
	})

	r.Len(wm, 2)
	r.Equal("A", wm["a/a.txt"])
	r.Equal("B", wm["b/b.txt"])

	box.Remove("b/b.txt")

	_, err = box.Find("b/b.txt")
	r.Error(err)
}

var httpBox = func() *MemoryBox {
	box := NewMemoryBox()
	box.AddString("hello.txt", "hello world!")
	box.AddString("index.html", "<h1>Index!</h1>")
	return box
}()

func Test_HTTPBox(t *testing.T) {
	r := require.New(t)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(httpBox))

	req, err := http.NewRequest("GET", "/hello.txt", nil)
	r.NoError(err)

	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	r.Equal(200, res.Code)
	r.Equal("hello world!", strings.TrimSpace(res.Body.String()))
}

func Test_HTTPBox_NotFound(t *testing.T) {
	r := require.New(t)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(httpBox))

	req, err := http.NewRequest("GET", "/notInBox.txt", nil)
	r.NoError(err)

	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	r.Equal(404, res.Code)
}

func Test_HTTPBox_Handles_IndexHTML(t *testing.T) {
	r := require.New(t)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(httpBox))

	req, err := http.NewRequest("GET", "/", nil)
	r.NoError(err)

	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	r.Equal(200, res.Code)

	r.Equal("<h1>Index!</h1>", strings.TrimSpace(res.Body.String()))
}

func Test_HTTPBox_CaseInsensitive(t *testing.T) {
	mux := http.NewServeMux()
	httpBox.AddString("myfile.txt", "this is my file")
	mux.Handle("/", http.FileServer(httpBox))

	for _, path := range []string{"/MyFile.txt", "/myfile.txt", "/Myfile.txt"} {
		t.Run(path, func(st *testing.T) {
			r := require.New(st)

			req, err := http.NewRequest("GET", path, nil)
			r.NoError(err)

			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)
			res.Flush()

			r.Equal(200, res.Code)
			r.Equal("this is my file", strings.TrimSpace(res.Body.String()))
		})
	}
}
