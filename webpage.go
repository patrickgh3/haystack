package main

import (
    "html/template"
    "os"
    "bufio"
    "time"
)

const indexFilepath = "html/index.html"
var templ = template.Must(template.ParseFiles(indexFilepath))

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

// RebuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage () {
    var pd WebpageData
    pd.BuildTimeStr = time.Now().Format(time.RFC850)

    channelNames := UniqueChannels()
    for i := 0; i < len(channelNames); i++ {
        c := WChannel{Name: channelNames[i]}
        for i := 0; i < 10; i++ {
            t := WThumb{}
            t.Filled = false
            if (i != 3 && i != 5) {
                t.ImageUrl = siteBaseUrl + "/images/100x80"
                t.VodUrl = vodBaseUrl + "/" + "12341234"
                t.Filled = true
            }
            c.Thumbs = append(c.Thumbs, t)
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

