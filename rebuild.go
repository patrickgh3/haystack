package main

import (
    "time"
    "fmt"
    "html/template"
    "os"
    "bufio"
)

var templ = template.Must(template.New("following").Parse(templateStr))

// RebuildPage generates a new HTML page with Twitch following data
// and writes it to file.
func RebuildPage() {
    fmt.Println("Rebuilding")
    timeStr := time.Now().Format(time.RFC850)

    f, err := os.Create(outFilename)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    w := bufio.NewWriter(f)
    templ.Execute(w, timeStr)
    w.Flush()
}

func main() {
    RebuildPage()
    ticker := time.NewTicker(5 * time.Second)
    for {
        select {
        case <- ticker.C:
            RebuildPage()
        }
    }
}

const outFilename = "/var/html/cwpat.me/following/out.html"

const templateStr = `
<html>
<head>
<title>Following</title>
</head>
<body>
{{.}}
</body>
</html>
`
