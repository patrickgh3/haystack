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

const indexFilepath = "html/index.html"
const vodUrlTimeFormat = "15h04m05s"
const vodBaseUrl = "https://www.twitch.tv/videos/"
const labelTimeFormat = "Monday 3pm MST"

var templ *template.Template

type WebpageData struct {
    BuildTimeStr string
    NumChannels int
    Channels []Channel
    TimeLabels []string
}

type Channel struct {
    Name string
    Thumbs []Thumb
}

type Thumb struct {
    Filled bool
    HasVod bool
    ImageUrl string
    VodUrl string
}

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

// RebuildWebpage generates an HTML page with up-to-date thumbnail content.
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
    for _, channelName := range channelNames {
        c := Channel{Name: channelName}
        for i := 0; i < config.Timing.NumPeriods; i++ {
            t := Thumb{}
            t.Filled = false
            c.Thumbs = append(c.Thumbs, t)
        }

        thumbs := database.ChannelThumbs(channelName)
        for _, thumb := range thumbs {
            // TODO: fix timezone offset
            t := thumb.CreatedTime.Add(time.Duration(5) * time.Hour)
            vodTimeString := thumb.VODTimeTime.Format(vodUrlTimeFormat)

            col := columnOfTime(t, roundTime)
            // TODO: warn if too old or too new thumb is present at this point
            c.Thumbs[col].HasVod = thumb.VOD != ""
            /*if thumb.Channel == "DestinationMystery" {
                c.Thumbs[col].HasVod = false
            }*/
            c.Thumbs[col].Filled = true
            c.Thumbs[col].ImageUrl = config.Path.SiteUrl + thumb.Image
            c.Thumbs[col].VodUrl = vodBaseUrl + thumb.VOD +
                    "?t=" + vodTimeString
        }
        pd.Channels = append(pd.Channels, c)
    }
    pd.NumChannels = len(channelNames)

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

