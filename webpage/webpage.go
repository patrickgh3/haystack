package webpage

import (
    "html/template"
    "os"
    "bufio"
    "time"
    "sort"
    "github.com/kardianos/osext"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

const indexFilepath = "templates/newindex.html"
const groupTimeFormat = "Monday 2006-01-02"

var templ *template.Template

type WebpageData struct {
    AppBaseUrl string
    PanelGroups []PanelGroup
}

type PanelGroup struct {
    Title string
    StreamPanels []StreamPanel
}

type StreamPanel struct {
    StreamID uint
    ChannelDisplayName string
    ChannelName string
    CoverImages []string
    Title string
}

// Times implements sort.Interface
type Times []time.Time
func (t Times) Len() int { return len(t) }
func (t Times) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t Times) Less(i, j int) bool { return t[i].Before(t[j]) }

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
    wpd.AppBaseUrl = config.Path.SiteUrl

    // For each stream in the DB, create a panel
    streams := database.GetAllStreams()

    // TODO: sort streams first by day, then by currently live, then by viewers

    // Group streams by day
    streamgroups := make(map[time.Time][]database.Stream)
    for _, stream := range streams {
        t := stream.StartTime
        rounded := time.Date(
                t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
        streamgroups[rounded] = append(streamgroups[rounded],
                stream)
    }

    // Sort time groups by most recent
    var groupTimes Times
    for k, _ := range streamgroups { groupTimes = append(groupTimes, k) }
    sort.Sort(sort.Reverse(groupTimes))

    for gi, groupTime := range groupTimes {
        streams := streamgroups[groupTime]
        // TODO: further sort streams in group by currently live, then viewers

        panelgroup := PanelGroup{Title:groupTime.Format(groupTimeFormat)}
        if gi == 0 {
            panelgroup.Title = ""
        }

        for _, stream := range streams {
            panel := StreamPanel{ChannelDisplayName:stream.ChannelDisplayName,
                    Title:stream.Title, StreamID:stream.ID}

            // Grab 4 representative images from the stream for its panel
            // [----|--------|--------|--------|----] where | are chosen images
            thumbs := database.GetStreamThumbs(stream.ID)
            numCoverImages := 4
            for i := 0; i < numCoverImages; i++ {
                // For each i, calculate fraction it is through the stream
                fractionThroughStream :=
                        (float64(i) + 0.5) / float64(numCoverImages)
                // The closest thumb index that fraction corresponds to
                chosenThumbIndex := int(
                        fractionThroughStream * float64(len(thumbs)-1))
                // Add image url of that thumb to the list
                imageUrl := config.Path.SiteUrl +
                        thumbs[chosenThumbIndex].ImagePath
                panel.CoverImages = append(panel.CoverImages, imageUrl)
            }
            panelgroup.StreamPanels = append(panelgroup.StreamPanels, panel)
        }
        wpd.PanelGroups = append(wpd.PanelGroups, panelgroup)
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

