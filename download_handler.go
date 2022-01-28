package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DownloadHandler struct {
	root http.FileSystem
}

func (d *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	now := time.Now()
	msg := fmt.Sprintf("%s [%s] [%s %s] [%s]\n", getUserIP(r), now.String(), r.Method, r.URL, r.UserAgent())
	fmt.Println(msg)

	switch r.Method {
	case "GET":
		serveFile(w, r, d.root, upath)
	default:
		fmt.Printf("error: not supprot method")
	}
}

func getUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	return IPAddress
}

func serveFile(w http.ResponseWriter, r *http.Request, root http.FileSystem, name string) {
	f, err := root.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	if fileInfo.IsDir() {
		dirList(w, r, f)
	} else {
		downloadFile(w, r, f, fileInfo)
	}
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

type anyDirs interface {
	len() int
	name(int) string
	isDir(int) bool
}

type fileInfos []fs.FileInfo

func (f fileInfos) len() int          { return len(f) }
func (f fileInfos) name(i int) string { return f[i].Name() }
func (f fileInfos) isDir(i int) bool  { return f[i].IsDir() }

func dirList(w http.ResponseWriter, r *http.Request, f http.File) {
	var entries anyDirs
	if list, err := f.Readdir(-1); err != nil {
		fmt.Printf("http: error reading directory: %v\n", err)
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	} else {
		entries = fileInfos(list)
	}

	sort.Slice(entries, func(i, j int) bool {return entries.name(i) < entries.name(j)})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache")

	fmt.Fprintf(w, "<html><body>")
	fmt.Fprintf(w, "<h2>%s</h2><hr>\n", "Directory listing for /")
	fmt.Fprint(w, "<ul>")
	for i, n := 0, entries.len(); i < n; i++ {
		name := entries.name(i)
		if entries.isDir(i) {
			name += "/"
		}

		fileUrl := url.URL{Path: "/" + name}
		fmt.Fprintf(w, `<li><a href="%s">%s</a>`, fileUrl.String(), name)
	}
	fmt.Fprint(w, "</ul><hr></body></html>")
}

func downloadFile(w http.ResponseWriter, r *http.Request, f http.File, stat fs.FileInfo) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	// copy file
	if _, err := io.Copy(w, f); err != nil {
		fmt.Printf("copy file error: %s\n", err)
	}
}

