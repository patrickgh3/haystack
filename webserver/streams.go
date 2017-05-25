package webserver

import (
    "fmt"
    "time"
    "net/http"
    "html/template"
    "strconv"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

const twitchVodBaseUrl = "https://www.twitch.tv/videos/"

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


