/*
Extremely simple Go tool that serves files out of a given folder, using a file
resolution algorithm similar to GitHub Pages, Netlify, or the default Nginx
config. Useful for local development. Provides a Go "library" (less than 100
LoC) and an optional CLI tool.

See `readme.md` for examples and additional details.
*/
package srv

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	ZIP_EXT = `.zip`
)

/*
Serves static files, resolving URL/HTML in a fashion similar to the default
Nginx config, Github Pages, and Netlify. Implements `http.Handler`. Can be used
as an almost drop-in replacement for `http.FileServer`.
*/
type FileServer struct {
	Dir      string
	Handlers []FileServable
}

func (self *FileServer) AppendHandler(handler FileServable) FileServer {
	self.Handlers = append(self.Handlers, handler)
	return *self
}

/*
Implements `http.Hander`.

Minor note: this has a race condition between checking for a file's existence
and actually serving it. Serving a file is not an atomic operation; the file
may be deleted or changed midway. In a production-grade version, this condition
would probably be addressed.
*/
func (self FileServer) ServeHTTP(rew http.ResponseWriter, req *http.Request) {
	switch req.Method {
	default:
		http.Error(rew, "", http.StatusMethodNotAllowed)
		return
	case http.MethodHead, http.MethodOptions:
		return
	case http.MethodGet:
	}

	dir := string(self.Dir)
	reqPath := req.URL.Path
	filePath := fpj(dir, reqPath)

	/**
	Ends with slash? Return error 404 for hygiene. Directory links must not end
	with a slash. It's unnecessary, and GH Pages will do a 301 redirect to a
	non-slash URL, which is a good feature but adds latency.
	*/
	// if len(reqPath) > 1 && reqPath[len(reqPath)-1] == '/' {
	// 	goto notFound
	// }

	hans := append([]FileServable{PlainFileServer(self.Dir)}, self.Handlers...)
	for _, han := range hans {
		if han.FileExists(filePath) {
			han.ServeFile(rew, req, filePath)
			return
		}

		// Has extension? Don't bother looking for +".html" or +"/index.html".
		// if path.Ext(reqPath) != "" {
		// 	NotFoundFileServer(self.Dir).ServeFile(rew, req, filePath)
		// 	return
		// }

		// Try +".html".
		{
			candidatePath := filePath + ".html"
			if han.FileExists(candidatePath) {
				han.ServeFile(rew, req, candidatePath)
				return
			}
		}

		// Try +"/index.html".
		{
			candidatePath := fpj(filePath, "index.html")
			if han.FileExists(candidatePath) {
				han.ServeFile(rew, req, candidatePath)
				return
			}
		}
	}
}

type FileServable interface {
	FileExists(filePath string) bool
	ServeFile(rew http.ResponseWriter, req *http.Request, filePath string)
}

type NotFoundFileServer PlainFileServer

func (self NotFoundFileServer) filePath() string {
	return fpj(string(self), "404.html")
}

func (self NotFoundFileServer) FileExists(_ string) bool {
	return fileExists(self.filePath())
}

// Minor issue: sends code 200 instead of 404 if "404.html" is found; not
// worth fixing for local development.
func (self NotFoundFileServer) ServeFile(rew http.ResponseWriter, req *http.Request, _ string) {
	http.ServeFile(rew, req, self.filePath())
}

type PlainFileServer string

func (PlainFileServer) FileExists(filePath string) bool {
	return fileExists(filePath)
}

func (self PlainFileServer) ServeFile(rew http.ResponseWriter, req *http.Request, filePath string) {
	http.ServeFile(rew, req, filePath)
}

type ZipFileServer string

func (self ZipFileServer) FileExists(filePath string) bool {
	filePath, _ = splitFilePathWithExt(filePath, ZIP_EXT)
	return fileExists(filePath)
}

func (self ZipFileServer) ServeFile(rew http.ResponseWriter, req *http.Request, filePath string) {
	zipFilePath, inZipFilePath := splitFilePathWithExt(filePath, ZIP_EXT)

	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		if errNotExistOrPanic(err) {
			NotFoundFileServer(self).ServeFile(rew, req, filePath)
			return
		}
	}
	defer zipReader.Close()

	req.URL.Path = inZipFilePath

	file, err := zipReader.Open(inZipFilePath)
	if err != nil {
		if errNotExistOrPanic(err) {
			NotFoundFileServer(self).ServeFile(rew, req, filePath)
			return
		}
	}

	rew.Header().Set(`Content-Type`, mime.TypeByExtension(filepath.Ext(inZipFilePath)))
	io.Copy(rew, file)
	return
}

func fpj(path ...string) string { return filepath.Join(path...) }

func fileExists(filePath string) bool {
	stat, _ := os.Stat(filePath)
	return stat != nil && !stat.IsDir()
}

/*
Splits a given file path into two parts: the archive part and the file part. The arch
variable holds the part of the path up to and including the first occurrence of the
provided extension. The file variable holds the remainder of the path after the
provided extension.

	splitFilePathWithExt(`/report/archive.zip/public/index.html`, `zip`)

	Returns:
		arch := `/report/archive.zip`
		file := `/public/index.html`
*/
func splitFilePathWithExt(val string, ext string) (arch string, file string) {
	vals := strings.Split(val, string(filepath.Separator))
	for ind, val := range vals {
		if filepath.Ext(val) == ext {
			arch = filepath.Join(vals[:ind+1]...)
			file = filepath.Join(vals[ind+1:]...)
			break
		}
	}
	return
}

func errNotExistOrPanic(err error) bool {
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
		return false
	}
	panic(err)
}
