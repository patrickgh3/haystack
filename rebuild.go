package main

import (
    "fmt"
    "os"
    "bufio"
    "html/template"
    "time"
    "github.com/spf13/viper"
    "github.com/kardianos/osext"
    "path"
)

var templ = template.Must(template.New("following").Parse(templateStr))

// RebuildPage generates a new HTML page with Twitch following data
// and writes it to file.
func RebuildPage(outPath string) {
    fmt.Print("Rebuilding...")

    // Compute contents
    timeStr := time.Now().Format(time.RFC850)

    // Open html file, write contents, close it
    f, err := os.Create(outPath + "/index.html")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    w := bufio.NewWriter(f)
    templ.Execute(w, timeStr)
    w.Flush()

    fmt.Println("done")
}

func main() {
    // Set config file location to current directory
    viper.SetConfigName("config")
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    viper.AddConfigPath(ef)

    // Read config values
    viper.ReadInConfig()
    outPath := viper.GetString("out_path")
    outPath = path.Clean(outPath)
    seconds := viper.GetInt("interval_seconds")
    refreshDuration := time.Duration(seconds) * time.Second

    // Do an initial refresh
    // (useful to quickly verify it's working)
    RebuildPage(outPath)

    // Start periodic refreshes
    ticker := time.NewTicker(refreshDuration)
    for {
        select {
        case <- ticker.C:
            RebuildPage(outPath)
        }
    }
}

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

