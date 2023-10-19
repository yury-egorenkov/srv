package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/mitranim/srv"
)

const help = `srv starts a local HTTP server, serving files from the given
directory (default .), approximary like it would be served by
default Nginx, GitHub Pages, Netlify, and so on. Uses a random
port by default.

Usage:

	srv
	srv -d <dir>
	srv -p <port>
	srv -h <host>
	srv -c

Settings:

`

var conf = struct {
	Port int
	Host string
	Server
}{
	Host:   "localhost",
	Server: Server{Dir: "."},
}

type Server struct {
	Dir  string
	Cors bool
}

var logger = log.New(os.Stderr, "", 0)

func main() {
	flag.StringVar(&conf.Dir, "d", conf.Dir, "dir")
	flag.IntVar(&conf.Port, "p", conf.Port, "port")
	flag.StringVar(&conf.Host, "h", conf.Host, "host")
	flag.BoolVar(&conf.Cors, "c", conf.Cors, "cors")

	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), help)
		flag.PrintDefaults()
	}

	flag.Parse()
	args()

	crit(serve())
}

func serve() error {
	// This allows us to find the OS-provided port.
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", conf.Host, conf.Port))
	if err != nil {
		return err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{Handler: conf.Server}

	logger.Printf("[srv] serving %q on %v\n", conf.Dir, fmt.Sprintf("http://%v:%v", conf.Host, port))
	return server.Serve(listener)
}

func args() {
	args := flag.Args()
	if len(args) == 0 {
		return
	}

	if args[0] == "help" {
		flag.Usage()
		os.Exit(0)
	}

	crit(fmt.Errorf(`[srv] unexpected arguments %q`, args))
}

func crit(err error) {
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%+v\n", err)
		os.Exit(1)
	}
}

func (self Server) ServeHTTP(rew http.ResponseWriter, req *http.Request) {
	if self.Cors {
		allowCors(rew.Header())
	}
	srv.FileServer(self.Dir).ServeHTTP(rew, req)
}

func allowCors(header http.Header) {
	header.Add("access-control-allow-credentials", "true")
	header.Add("access-control-allow-headers", "content-type")
	header.Add("access-control-allow-methods", "OPTIONS, GET, HEAD, POST, PUT, PATCH, DELETE")
	header.Add("access-control-allow-origin", "*")
}
