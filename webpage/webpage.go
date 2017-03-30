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
                title := truncateString(thumb.Status)
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

    // Insert streams into cells
    for _, stream := range streams {
        // Find or make a row r that we can insert this stream into
        spaceLeft := 6
        valid := func(row int, pos int) bool {
            for i := 0; i <= spaceLeft; i++ {
                if pos-i < 0 || pd.Cells[row][pos-i].Filled {
                    return false
                }
            }
            return true
        }
        var r int
        for r = 0; r < len(pd.Cells); r++ {
            if valid(r, stream.StartPos) {
                break
            }
        }
        if r == len(pd.Cells) {
            pd.Cells = append(pd.Cells, make([]*Cell, config.Timing.NumPeriods+1))
            for i := 0; i < config.Timing.NumPeriods+1; i++ {
                pd.Cells[len(pd.Cells)-1][i] = &Cell{};
            }
        }

        // Insert the stream into row r
        cell := pd.Cells[r][stream.StartPos-1]
        cell.Filled = true
        cell.Type = 1
        cell.ChannelName = stream.ChannelName
        cell.Title = stream.Title
        for d, thumb := range stream.Thumbs {
            cell := pd.Cells[r][stream.StartPos+d]
            cell.Filled = true
            cell.Type = 0
            cell.HasVod = thumb.VOD != ""
            //cell.HasVod = false
            cell.ImageUrl = config.Path.SiteUrl + thumb.Image
            vodTimeString := thumb.VODTimeTime.Format(vodUrlTimeFormat)
            cell.VodUrl = vodBaseUrl + thumb.VOD +
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

