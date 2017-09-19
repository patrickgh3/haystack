package webserver

import (
    "fmt"
    "net"
    "net/http"
    "net/http/fcgi"
    "net/url"
    "strings"
    "path"
    "html/template"
    "strconv"
    "github.com/kardianos/osext"
    "github.com/patrickgh3/haystack/config"
    //"github.com/patrickgh3/haystack/database"
)

const filterFilepath = "templates/filter.html"
var filterTempl *template.Template
const directoryFilepath = "templates/directory.html"
var directoryTempl *template.Template

// InitTemplates initializes page templates from the included files
func InitTemplates() {
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    filterTempl = template.Must(
            template.ParseFiles(ef + "/" + filterFilepath))
    directoryTempl = template.Must(
            template.ParseFiles(ef + "/" + directoryFilepath))
}

type FastCGIServer struct{}

// ServeHTTP computes responses for HTTP requests passed to the app.
func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("got request: %v\n", r.URL.Path)

    // Grab path prefix from config site url and clean
    siteURL, err := url.Parse(config.Path.SiteUrl)
    if err != nil {
        panic(err)
    }
    prefix := path.Clean(siteURL.Path)

    // Trim prefix from request path and clean
    reqPath := path.Clean(r.URL.Path)
    if !strings.HasPrefix(reqPath, prefix) {
        fmt.Printf("URL not prefixed, nginx probably shouldnt have given us")
        w.WriteHeader(http.StatusBadGateway)
        return
    }
    subPath := strings.TrimPrefix(reqPath, prefix)
    subPath = path.Clean(subPath)

    // Handle the various application requests
    if subPath == "." {
        ServeRootPage(w, r)
    } else if subPath == "/stream" {
        ServeStreamRequest(w, r)
    } else {
        w.WriteHeader(http.StatusNotFound)
    }
    /*    // Try serving filter page
        filterPath := strings.ToLower(subPath[1:len(subPath)])
        f := database.GetFilterWithSubpath(filterPath)
        if f != nil {
            ServeFilterPage(w, r, *f)
        } else {
            // If URL not otherwise handled, then 404
            w.WriteHeader(http.StatusNotFound)
        }
    }*/
}

// Serve starts the web server and blocks.
func Serve() {
    fmt.Printf("Starting server...\n")
    l, _ := net.Listen("tcp", config.WebServer.IP+":"+
            strconv.Itoa(config.WebServer.Port))
    b := new(FastCGIServer)
    fcgi.Serve(l, b)
}
