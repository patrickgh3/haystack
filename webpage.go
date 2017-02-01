package main

import (
    "html/template"
    "os"
    "bufio"
    "time"
)

const templateStr = `
<html>
<head>
<title>Pitchfork</title>
</head>
<body>
{{.}}
</body>
</html>
`

// RebuildWebpage generates an HTML page with up-to-date thumbnail content.
func BuildWebpage () {
    // Compute contents
    timeStr := time.Now().Format(time.RFC850)

    //content, err := ioutil.ReadFile(
    var templ = template.Must(template.New("following").Parse(templateStr))

    // Open html file, write contents, close it
    f, err := os.Create(outPath + "/index.html")
    defer f.Close()
    if err != nil {
        panic(err)
    }

    w := bufio.NewWriter(f)
    templ.Execute(w, timeStr)
    w.Flush()
}

