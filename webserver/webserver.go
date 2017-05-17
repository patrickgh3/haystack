package webserver

import (
    "fmt"
    "net/http"
    "net/http/fcgi"
    "net"
    "strings"
    "path"
    "html/template"
    "strconv"
    "time"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

var streamTemplate = template.Must(template.New("streamTemplate").Parse(
`{{range .Thumbs}}
<a href="{{.LinkUrl}}" target="_blank">
    <img src="{{.ImageUrl}}" onmousemove="magnify(this, true)" onmouseout="unmagnify()">
</a>
{{end}}
`))

type StreamResponseData struct {
    Thumbs []StreamResponseThumb
}

type StreamResponseThumb struct {
    LinkUrl string
    ImageUrl string
}

const twitchVodBaseUrl = "https://www.twitch.tv/videos/"

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
        // Parse id param as uint
        sid, err := strconv.ParseUint(vals.Get("id"), 10, 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Printf("Bad stream ID\n")
        } else {
            streamId := uint(sid)
            thumbs := database.GetStreamThumbs(streamId)
            var td StreamResponseData
            for _, thumb := range thumbs {
                // String() formats like "15h04m05s"
                timeStr := time.Duration(
                        time.Duration(thumb.VODSeconds)*time.Second).String()
                td.Thumbs = append(td.Thumbs, StreamResponseThumb{
                        LinkUrl:twitchVodBaseUrl+thumb.VOD+"?t="+timeStr,
                        ImageUrl:config.Path.SiteUrl+thumb.ImagePath})
            }
            // Fill thumbs into response HTML
            streamTemplate.Execute(w, td)
        }

    } else {
        w.WriteHeader(http.StatusNotFound)
    }
}

// Serve starts the web server.
func Serve() {
    fmt.Printf("Starting server...\n")
    l, _ := net.Listen("tcp", "127.0.0.1:4424") // TODO: port (IP?) configurable
    b := new(FastCGIServer)
    fcgi.Serve(l, b)
}
