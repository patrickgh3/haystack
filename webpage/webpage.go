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

const indexFilepath = "templates/index.html"
const groupTimeFormat = "Monday 2006-01-02"

var templ *template.Template

type WebpageData struct {
    AppBaseUrl string
    PanelGroups []PanelGroup
}

type PanelGroup struct {
    StreamPanels []StreamPanel
    Title string
}

type StreamPanel struct {
    StreamID uint
    ChannelDisplayName string
    CoverImages []string
    Title string
    Live bool
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

    // Grab all streams from the DB
    streams := database.GetAllStreams()

    // Group streams by day into a Time -> []Stream map
    streamGroups := make(map[time.Time][]database.Stream)
    for _, stream := range streams {
        // Round time to nearest day in its time zone
        t := stream.StartTime
        rounded := time.Date(
                t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
        streamGroups[rounded] = append(streamGroups[rounded], stream)
    }

    // Find sequential ordering of all group times
    var groupTimes Times
    for k, _ := range streamGroups { groupTimes = append(groupTimes, k) }
    sort.Sort(sort.Reverse(groupTimes))

    // For each stream group, create a panel group
    for gi, groupTime := range groupTimes {
        var panelgroup PanelGroup
        // Top group has no title
        if gi != 0 {
            panelgroup.Title = groupTime.Format(groupTimeFormat)
        }

        // For each stream, create a stream panel and put it into either
        // live or not live array
        groupStreams := streamGroups[groupTime]
        var livePanels []StreamPanel
        var notlivePanels []StreamPanel

        for _, stream := range groupStreams {
            // Create panel for this stream
            panel := PanelOfStream(stream)
            // Determine whether this stream is considered Live or not and
            // add it to the appropriate list
            cutoff := roundTime.Add(-config.Timing.CutoffLeeway)
            if stream.LastUpdateTime.After(cutoff) ||
                    stream.LastUpdateTime.Equal(cutoff) {
                panel.Live = true
                livePanels = append(livePanels, panel)
            } else {
                panel.Live = false
                notlivePanels = append(notlivePanels, panel)
            }
        }
        // Sort live and not live individually by viewer count
        // TODO

        // Add Live then Not Live to the actual group
        for _, panel := range livePanels {
            panelgroup.StreamPanels = append(panelgroup.StreamPanels, panel)
        }
        for _, panel := range notlivePanels {
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

// PanelOfStream generates a StreamPanel based on a stream
func PanelOfStream(stream database.Stream) StreamPanel {
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
        // Add image url of that thumb to the array
        imageUrl := config.Path.SiteUrl +
                thumbs[chosenThumbIndex].ImagePath
        panel.CoverImages = append(panel.CoverImages, imageUrl)
    }
    return panel
}

