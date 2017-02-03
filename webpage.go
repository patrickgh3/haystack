package main

import (
    "html/template"
    "os"
    "bufio"
    "time"
    "bytes"
)

const indexFilepath = "html/index.html"
var templ = template.Must(template.ParseFiles(indexFilepath))

const rowTemplateStr = `
<h2>{{.}}</h2>
`
var rowTempl = template.Must(template.New("row").Parse(rowTemplateStr))

// RebuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage () {
    // Compute contents
    timeStr := time.Now().Format(time.RFC850)
    timeStr += "a"

    channelNames := UniqueChannels()
    var c bytes.Buffer
    for i := 0; i < len(channelNames); i++ {
        rowTempl.Execute(&c, channelNames[i])
    }

    // Open html file, write contents, close it
    f, err := os.Create(outPath + "/index.html")
    defer f.Close()
    if err != nil {
        panic(err)
    }

    w := bufio.NewWriter(f)
    //tplVars := []interface{} {timeStr, template.HTML(c.String())}
    templ.Execute(w, template.HTML(c.String()))
    w.Flush()
}

