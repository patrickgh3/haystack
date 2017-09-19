package webserver

import (
    "time"
    "sort"
    "fmt"
    "math"
    "io"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

const groupTimeFormat = "Monday 2006-01-02"

type WebpageData struct {
    Title string
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
    Viewers int
    Length string
}

// ServeFilterPage serves a page listing all streams of a filter.
func ServeFilterPage(w io.Writer, f database.Filter) {
    roundTime := f.LastUpdateTime
    var wpd WebpageData
    wpd.AppBaseUrl = config.Path.SiteUrl
    wpd.Title = f.Name

    // Grab filter's streams from the DB
    streams := database.GetStreamsOfFilter(f.ID)

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
        sort.Sort(ByViewers(livePanels))
        sort.Sort(ByViewers(notlivePanels))

        // Add Live then Not Live to the actual group
        for _, panel := range livePanels {
            panelgroup.StreamPanels = append(panelgroup.StreamPanels, panel)
        }
        for _, panel := range notlivePanels {
            panelgroup.StreamPanels = append(panelgroup.StreamPanels, panel)
        }

        wpd.PanelGroups = append(wpd.PanelGroups, panelgroup)
    }

    // Execute template to HTTP response
    filterTempl.Execute(w, wpd)
}

// PanelOfStream generates a StreamPanel based on a stream
func PanelOfStream(stream database.Stream) StreamPanel {
    dur := time.Duration(stream.NumThumbs) * config.Timing.Period
    mins := int(math.Mod(float64(dur.Minutes()), 60))
    lengthStr := fmt.Sprintf("%d:%02d", int(dur.Hours()), mins)

    panel := StreamPanel{ChannelDisplayName:stream.ChannelDisplayName,
            Title:stream.Title, StreamID:stream.ID,
            Viewers:int(stream.AverageViewers), Length:lengthStr}

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

func truncateString (s string) string {
    maxLength := 24
    if len(s) > maxLength {
        s = s[:maxLength-3] + "..."
    }
    return s
}

// Times implements sort.Interface
type Times []time.Time
func (t Times) Len() int { return len(t) }
func (t Times) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t Times) Less(i, j int) bool { return t[i].Before(t[j]) }

// ByViewers implements sort.Interface
type ByViewers []StreamPanel
func (t ByViewers) Len() int { return len(t) }
func (t ByViewers) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t ByViewers) Less(i, j int) bool { return t[i].Viewers > t[j].Viewers }
