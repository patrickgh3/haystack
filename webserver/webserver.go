package webserver

import (
    "fmt"
    "net/http"
    "net/http/fcgi"
    "net"
    "strings"
    "path"
)

type FastCGIServer struct{}

// ServeHTTP computes responses for HTTP requests passed to the app.
func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("got request: %v\n", r.URL.Path)

    // Strip app root url prefix from url, and clean it
    prefix := "/haystack-dev"
    if strings.HasPrefix(r.URL.Path, prefix) {
        r.URL.Path = path.Clean(r.URL.Path)
        r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
    } else {
        fmt.Printf("URL not prefixed, nginx probably shouldnt have given us")
        w.WriteHeader(http.StatusBadGateway)
        return
    }

    /*

    Requests:

    /                   Returns main browsing webpage
    /stream?id=N        Returns <a><img></img></a> list of thumbs for stream X

    */

    // Handle the various application requests
    vals := r.URL.Query()
    if r.URL.Path == "/stream" {
        // TODO:
        // Query DB for all thumbs of a given stream
        // Construct (from template?) series of <a><img></img></a> tags
        writeStr := "Stream: " + vals.Get("id")
        w.Write([]byte(writeStr))
    // TODO: home page request
    } else {
        w.WriteHeader(http.StatusNotFound)
    }
}

// Serve starts the web server.
func Serve() {
    fmt.Printf("Starting server...")
    l, _ := net.Listen("tcp", "127.0.0.1:4424") // TODO: port (IP?) configurable
    b := new(FastCGIServer)
    fcgi.Serve(l, b)
}
