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
var templ *template.Template
const vodUrlTimeFormat = "15h04m05s"
const vodBaseUrl = "https://www.twitch.tv/videos"

type WebpageData struct {
    BuildTimeStr string
    NumChannels int
    Channels []WChannel
}

type WChannel struct {
    Name string
    Thumbs []WThumb
}

type WThumb struct {
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

// ColumnOfTime returns which column a certain time corresponds to.
func ColumnOfTime (t time.Time, roundTime time.Time) int {
    x := int(roundTime.Sub(t).Seconds() / config.Timing.Period.Seconds())
    return (config.Timing.NumPeriods - 1) - x
}

// RebuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage (roundTime time.Time) {
    var pd WebpageData
    pd.BuildTimeStr = time.Now().Format(time.RFC850)

    channelNames := database.DistinctChannels()
    for i := 0; i < len(channelNames); i++ {
        c := WChannel{Name: channelNames[i]}
        for i := 0; i < config.Timing.NumPeriods; i++ {
            t := WThumb{}
            t.Filled = false
            c.Thumbs = append(c.Thumbs, t)
        }
        thumbs := database.ChannelThumbs(channelNames[i])
        for i := 0; i < len(thumbs); i++ {
            // TODO: fix timezone offset
            t := thumbs[i].CreatedTime.Add(time.Duration(5) * time.Hour)
            vodTimeString := thumbs[i].VODTimeTime.Format(vodUrlTimeFormat)

            col := ColumnOfTime(t, roundTime)
            c.Thumbs[col].HasVod = thumbs[i].VOD != ""
            /*if thumbs[i].Channel == "paragusrants" {
                c.Thumbs[col].HasVod = false
            }*/
            c.Thumbs[col].Filled = true
            c.Thumbs[col].ImageUrl = config.Path.SiteUrl + "/" + thumbs[i].Image
            c.Thumbs[col].VodUrl = vodBaseUrl + thumbs[i].VOD +
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

