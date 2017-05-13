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
const vodUrlTimeFormat = "15h04m05s"
const vodBaseUrl = "https://www.twitch.tv/videos/"
const labelTimeFormat = "Monday 2006-01-02"
const labelSpace = 8

var templ *template.Template

type WebpageData struct {
    BuildTimeStr string
    NumChannels int
    Cells [][]*Cell
    TimeLabels []string
}

type Cell struct {
    Filled bool
    Type int

    ChannelName string
    Title string

    HasVod bool
    ImageUrl string
    VodUrl string
}

type Stream struct {
    StartPos int
    ChannelName string
    Title string
    Thumbs []database.ThumbRow
}

type NewWebpageData struct {
    StreamPanels []StreamPanel
}

type StreamPanel struct {
    StreamID int
    ChannelDisplayName string
    ChannelName string
    CoverImages []string
    Title string
}

// ByStart implements sort.Interface for []Stream
type ByStart []Stream
func (b ByStart) Len() int { return len(b) }
func (b ByStart) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByStart) Less(i, j int) bool { return b[i].StartPos < b[j].StartPos }

func InitTemplate () {
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    templ = template.Must(template.ParseFiles(ef + "/" + indexFilepath))
}

func columnOfTime (t time.Time, roundTime time.Time) int {
    x := int(roundTime.Sub(t).Seconds() / config.Timing.Period.Seconds())
    return (config.Timing.NumPeriods - 1) - x + 1
}

func timeOfColumn (col int, roundTime time.Time) time.Time {
    columnsFromRight := (config.Timing.NumPeriods - 1) - (col - 1)
    deltaT := config.Timing.Period * time.Duration(-columnsFromRight)
    return roundTime.Add(deltaT)
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
    var wpd NewWebpageData
    /*channelNames := database.DistinctChannels()

    // Create list of streams from the database
    var streams []Stream
    var curStream Stream
    var lastpos int
    for _, channelName := range channelNames {
        thumbs := database.ChannelThumbsTimeAscending(channelName)
        for i, thumb := range thumbs {
            // TODO: fix timezone offset
            t := thumb.CreatedTime.Add(time.Duration(4) * time.Hour)
            curpos := columnOfTime(t, roundTime)
            // Detect gap and start new stream
            if curpos - lastpos > 1 || i == 0 {
                if i != 0 {
                    streams = append(streams, curStream)
                }
                //title := truncateString(thumb.Status)
                title := thumb.Status
                curStream = Stream{ChannelName:channelName, Title:title}
                curStream.StartPos = curpos
            }
            // Always append thumb to current stream
            curStream.Thumbs = append(curStream.Thumbs, thumb)
            lastpos = curpos
        }
        streams = append(streams, curStream)
    }

    sort.Sort(ByStart(streams))

    for _, stream := range streams {
        panel := StreamPanel{ChannelDisplayName:stream.ChannelName,
                            Title:stream.Title, StreamID:4}
        numCoverImages := 4
        numThumbs := len(stream.Thumbs)
        for i := 0; i < numCoverImages; i++ {
            t := int(float64(i)/float64(numCoverImages-1) * float64(numThumbs-1))
            panel.CoverImages = append(panel.CoverImages, config.Path.SiteUrl+stream.Thumbs[t].Image)
        }
        wpd.StreamPanels = append(wpd.StreamPanels, panel)
    }*/

    for x := 0; x < 5; x++ {
        panel := StreamPanel{ChannelDisplayName:"name", Title:"title", StreamID:x}
        numCoverImages := 4
        for i := 0; i < numCoverImages; i++ {
            panel.CoverImages = append(panel.CoverImages, config.Path.SiteUrl+"/images/test.jpg")
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

