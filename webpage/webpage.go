package webpage

import (
    "html/template"
    "os"
    "bufio"
    "time"
    "github.com/kardianos/osext"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

const indexFilepath = "templates/newindex.html"
const labelTimeFormat = "Monday 2006-01-02"

var templ *template.Template

type WebpageData struct {
    StreamPanels []StreamPanel
}

type StreamPanel struct {
    StreamID uint
    ChannelDisplayName string
    ChannelName string
    CoverImages []string
    Title string
}

// ByStart implements sort.Interface for []Stream
/*type ByStart []Stream
func (b ByStart) Len() int { return len(b) }
func (b ByStart) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByStart) Less(i, j int) bool { return b[i].StartPos < b[j].StartPos }*/

func InitTemplate () {
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    templ = template.Must(template.ParseFiles(ef + "/" + indexFilepath))
}

func truncateString (s string) string {
    maxLength := 24
    if len(s) > maxLength {
        s = s[:maxLength-3] + "..."
    }
    return s
}

// BuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage (roundTime time.Time) {
    var wpd WebpageData
    // For each stream in the DB, create a panel
    streams := database.GetAllStreams()
    // TODO: sort streams first by day, then by currently live, then by viewers
    for _, stream := range streams {
        panel := StreamPanel{ChannelDisplayName:stream.ChannelDisplayName,
                Title:"todo: title", StreamID:stream.ID}

        // Grab 4 representative images from the stream for its panel
        // [----|--------|--------|--------|----] where | are chosen images
        thumbs := database.GetStreamThumbs(stream.ID)
        numCoverImages := 4
        for i := 0; i < numCoverImages; i++ {
            // For each i, calculate fraction it is through the stream
            fractionThroughStream := (float64(i)+0.5)/float64(numCoverImages)
            // The closest thumb index that fraction corresponds to
            chosenThumbIndex := int(
                    fractionThroughStream * float64(len(thumbs)-1))
            // Add image url of that thumb to the list
            imageUrl := config.Path.SiteUrl +
                    thumbs[chosenThumbIndex].ImagePath
            panel.CoverImages = append(panel.CoverImages, imageUrl)
        }
        wpd.StreamPanels = append(wpd.StreamPanels, panel)
    }

    // Write to html file
    f, err := os.Create(config.Path.Root + "/index.html")
    defer f.Close()
    if err != nil {
        panic(err)
    }

    w := bufio.NewWriter(f)
    templ.Execute(w, wpd)
    w.Flush()
}

