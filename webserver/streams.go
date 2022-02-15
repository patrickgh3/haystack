package webserver

import (
    "fmt"
    "time"
    "net/http"
    "html/template"
    "strconv"
    "math"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

const twitchVodBaseUrl = "https://www.twitch.tv/videos/"

var streamTemplate = template.Must(template.New("streamTemplate").Parse(
//`<div class="detailslength">{{.Length}}</div>
`{{range .Thumbs}}
<a {{if .HasVOD}}href="{{.LinkUrl}}"{{else}}class="novodlink"{{end}} target="_blank">
    <img src="{{.ImageUrl}}" onmousemove="magnify(event, this, true)" onmouseout="unmagnify()">
</a>
{{end}}
`))
/*
<div class=".clips">
{{range .Clips}}
<a href="{{.ClipUrl}}"><img src="{{.ImageUrl}}"></a>
{{end}}
</div>
*/

type StreamResponseData struct {
    Length string
    Thumbs []StreamResponseThumb
    Clips  []StreamResponseClip
}
type StreamResponseThumb struct {
    LinkUrl string
    ImageUrl string
    HasVOD bool
}
type StreamResponseClip struct {
    ClipUrl string
    ImageUrl string
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

        // Stream length
        stream := database.GetStreamByID(streamId)
        dur := time.Duration(stream.NumThumbs) * config.Timing.Period
        mins := int(math.Mod(float64(dur.Minutes()), 60))
        td.Length = fmt.Sprintf("%d:%02d", int(dur.Hours()), mins)

        // Thumbs
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

        // Clips
        /*clips := database.GetStreamClips(streamId)
        for _, clip := range clips {
            td.Clips = append(td.Clips, StreamResponseClip{
                    ClipUrl:clip.ClipUrl, ImageUrl:clip.ImageUrl})
        }*/

        // Fill thumbs into response HTML
        streamTemplate.Execute(w, td)
        // Disable client-side caching
        w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expries", "0")
    }
}


