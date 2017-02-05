package main

import (
    "html/template"
    "os"
    "bufio"
    "time"
)

const indexFilepath = "html/index.html"
var templ = template.Must(template.ParseFiles(indexFilepath))
const vodUrlTimeFormat = "15h04m05s"

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
    ImageUrl string
    VodUrl string
}

// ColumnOfTime returns which column a certain time corresponds to.
func ColumnOfTime (t time.Time, roundTime time.Time) int {
    x := int(roundTime.Sub(t).Seconds() / refreshDuration.Seconds())
    return (numRefreshPeriods - 1) - x
}

// RebuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage (roundTime time.Time) {
    var pd WebpageData
    pd.BuildTimeStr = time.Now().Format(time.RFC850)

    channelNames := DistinctChannels()
    for i := 0; i < len(channelNames); i++ {
        c := WChannel{Name: channelNames[i]}
        for i := 0; i < numRefreshPeriods; i++ {
            t := WThumb{}
            t.Filled = false
            c.Thumbs = append(c.Thumbs, t)
        }
        thumbs := ChannelThumbs(channelNames[i])
        for i := 0; i < len(thumbs); i++ {
            t, err := time.Parse(mysqlTimestampFormat, thumbs[i].Created)
            if err != nil {
                panic(err)
            }
            t = t.Add(time.Duration(5) * time.Hour) // TODO: fix timezone offset
            vodTime, err := time.Parse(mysqlTimeFormat, thumbs[i].VODTime)
            if err != nil {
                panic(err)
            }
            vodTimeString := vodTime.Format(vodUrlTimeFormat)

            col := ColumnOfTime(t, roundTime)
            c.Thumbs[col].Filled = true
            c.Thumbs[col].ImageUrl = siteBaseUrl + thumbs[i].Image
            c.Thumbs[col].VodUrl = vodBaseUrl + "/" + thumbs[i].VOD +
                    "?t=" + vodTimeString
        }
        pd.Channels = append(pd.Channels, c)
    }
    pd.NumChannels = len(channelNames)

    // Write to html file
    f, err := os.Create(outPath + "/index.html")
    defer f.Close()
    if err != nil {
        panic(err)
    }
    w := bufio.NewWriter(f)
    templ.Execute(w, pd)
    w.Flush()
}
