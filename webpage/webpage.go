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

const indexFilepath = "html/index.html"
const vodUrlTimeFormat = "15h04m05s"
const vodBaseUrl = "https://www.twitch.tv/videos/"
const labelTimeFormat = "Monday 3pm MST"

var templ *template.Template

type WebpageData struct {
    BuildTimeStr string
    NumChannels int
    Cells [][]Cell
    TimeLabels []string
}

type Cell struct {
    Filled bool
    Type int

    ChannelName string

    HasVod bool
    ImageUrl string
    VodUrl string
}

type Stream struct {
    StartPos int
    ChannelName string
    Thumbs []database.ThumbRow
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
    return (config.Timing.NumPeriods - 1) - x
}

func timeOfColumn (col int, roundTime time.Time) time.Time {
    columnsFromRight := (config.Timing.NumPeriods - 1) - col
    deltaT := config.Timing.Period * time.Duration(-columnsFromRight)
    return roundTime.Add(deltaT)
}

// BuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage (roundTime time.Time) {
    var pd WebpageData
    pd.BuildTimeStr = time.Now().Format(time.RFC850)

    for i := 0; i < config.Timing.NumPeriods; i++ {
        label := ""
        t := timeOfColumn(i, roundTime)
        if t.Truncate(config.Timing.LabelPeriod) == t {
            label = t.Format(labelTimeFormat)
        }
        pd.TimeLabels = append(pd.TimeLabels, label)
    }

    channelNames := database.DistinctChannels()
    pd.NumChannels = len(channelNames)

    // Create list of streams from the database
    var streams []Stream
    for _, channelName := range channelNames {
        thumbs := database.ChannelThumbsTimeAscending(channelName)
        var curStream Stream
        var lastpos int
        for i, thumb := range thumbs {
            // TODO: fix timezone offset
            t := thumb.CreatedTime.Add(time.Duration(5) * time.Hour)
            curpos := columnOfTime(t, roundTime)
            if curpos - lastpos > 1 || i == 0 {
                if i != 0 {
                    streams = append(streams, curStream)
                }
                curStream = Stream{ChannelName:channelName}
                curStream.StartPos = curpos
            }
            curStream.Thumbs = append(curStream.Thumbs, thumb)
            lastpos = curpos
        }
        streams = append(streams, curStream)
    }

    sort.Sort(ByStart(streams))

    // Insert each stream into available rows, or a newly created row.
    for _, stream := range streams {
        // Find or make a row r that we can insert this stream into
        valid := func(row int, pos int) bool {
            return !pd.Cells[row][pos].Filled &&
                (pos-1 < 0 || !pd.Cells[row][pos-1].Filled) &&
                (pos-2 < 0 || !pd.Cells[row][pos-2].Filled)
        }
        var r int
        for r = 0; r < len(pd.Cells); r++ {
            if valid(r, stream.StartPos) {
                break
            }
        }
        if r == len(pd.Cells) {
            pd.Cells = append(pd.Cells, make([]Cell, config.Timing.NumPeriods))
        }

        // Insert the stream into row r
        for d, thumb := range stream.Thumbs {
            pd.Cells[r][stream.StartPos+d].Filled = true
            pd.Cells[r][stream.StartPos+d].HasVod = thumb.VOD != ""
            pd.Cells[r][stream.StartPos+d].ImageUrl = config.Path.SiteUrl + thumb.Image
            vodTimeString := thumb.VODTimeTime.Format(vodUrlTimeFormat)
            pd.Cells[r][stream.StartPos+d].VodUrl = vodBaseUrl + thumb.VOD +
                    "?t=" + vodTimeString
        }
    }

    // Write to html file
    f, err := os.Create(config.Path.Root + "/index.html")
    defer f.Close()
    if err != nil {
        panic(err)
    }

    w := bufio.NewWriter(f)
    templ.Execute(w, pd)
    w.Flush()
}

