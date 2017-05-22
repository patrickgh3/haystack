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
    "time"
    "io"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

var streamTemplate = template.Must(template.New("streamTemplate").Parse(
`{{range .Thumbs}}
<a {{if .HasVOD}}href="{{.LinkUrl}}"{{else}}class="novodlink"{{end}} target="_blank">
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
    HasVOD bool
}

const twitchVodBaseUrl = "https://www.twitch.tv/videos/"

type FastCGIServer struct{}

// ServeHTTP computes responses for HTTP requests passed to the app.
func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("got request: %v\n", r.URL.Path)

    // Strip app root url prefix from url, and clean it
    siteURL, err := url.Parse(config.Path.SiteUrl)
    if err != nil {
        panic(err)
    }
    prefix := siteURL.Path
    if !strings.HasPrefix(r.URL.Path, prefix) {
        fmt.Printf("URL not prefixed, nginx probably shouldnt have given us")
        w.WriteHeader(http.StatusBadGateway)
        return
    }
    subPath := path.Clean(r.URL.Path)
    subPath = strings.TrimPrefix(r.URL.Path, prefix)

    // Handle the various application requests
    if subPath == "/" {
        ServeRootPage(w, r)
    } else if subPath == "/stream" {
        ServeStreamRequest(w, r)
    } else {
        // Try serving filter page
        f := database.GetFilterWithSubpath(strings.ToLower(subPath[1:len(subPath)]))
        if f != nil {
            ServeFilterPage(w, r, *f)
        } else {
            // If URL not otherwise handled, then 404
            w.WriteHeader(http.StatusNotFound)
        }
    }
}

// ServeStreamRequest serves a series of <a><img></a> tags for thumbs of a
// specified stream.
func ServeStreamRequest(w http.ResponseWriter, r *http.Request) {
    // Parse id param as uint
    vals := r.URL.Query()
    sid, err := strconv.ParseUint(vals.Get("id"), 10, 64)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Printf("Bad stream ID\n")
    } else {
        streamId := uint(sid)
        // Generate data for template
        var td StreamResponseData
        thumbs := database.GetStreamThumbs(streamId)
        for _, thumb := range thumbs {
            // Format time like "15h04m05s"
            timeStr := time.Duration(
                    time.Duration(thumb.VODSeconds)*time.Second).String()
            linkUrl := ""
            hasVOD := len(thumb.VOD) > 0
            if hasVOD {
                linkUrl = twitchVodBaseUrl+thumb.VOD+"?t="+timeStr
            }
            td.Thumbs = append(td.Thumbs, StreamResponseThumb{
                    LinkUrl:linkUrl,
                    ImageUrl:config.Path.SiteUrl+thumb.ImagePath,
                    HasVOD:hasVOD})
        }

        // Fill thumbs into response HTML
        streamTemplate.Execute(w, td)
        // Disable client-side caching
        w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expries", "0")
    }
}

// ServeRootPage serves a page listing all filters.
func ServeRootPage(w http.ResponseWriter, r *http.Request) {
    // TODO
    filters := database.GetAllFilters()
    for _, filter := range filters {
        io.WriteString(w, fmt.Sprintf("%v: %v<br>", filter.Name, filter.Subpath))
    }
}

// Serve starts the web server.
func Serve() {
    fmt.Printf("Starting server...\n")
    l, _ := net.Listen("tcp", config.WebServer.IP+":"+
            strconv.Itoa(config.WebServer.Port))
    b := new(FastCGIServer)
    fcgi.Serve(l, b)
}
