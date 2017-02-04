package main

import (
    "time"
    "strconv"
    "strings"
    "fmt"
)

// main is the entry point into the program.
// It initializes stuff and then calls Update periodically.
func main () {
    ReadConfig()
    InitDB()

    // Do an initial update (useful to quickly verify it's working)
    Update()

    // Start periodic updates
    ticker := time.NewTicker(refreshDuration)
    for {
        <-ticker.C
        Update()
    }
}

// Update saves thumbnails of Twitch streams, deletes old ones, and
// builds a new webpage.
func Update () {
    curTime := time.Now()
    timeString := strconv.FormatInt(curTime.Unix(), 10);

    sr := TwitchAPIAllStreams("?game=I%20Wanna%20Be%20The%20Guy")

    for _, s := range sr.Streams {
        fmt.Printf("%v...\n", s.Channel.Display_name)

        channelName := strings.ToLower(s.Channel.Display_name)
        vodID := strconv.Itoa(s.Id)
        imageUrl := s.Preview.Medium

        subpath := imagesSubdir + "/" + channelName + "_" + timeString + ".jpg"
        path := outPath + "/" + subpath

        DownloadImage(imageUrl, path)

        InsertThumb(channelName, curTime, vodID, subpath)
    }

    BuildWebpage()

    numDeleted := DeleteOldThumbs()

    fmt.Printf("%v deleted\n", numDeleted)
    fmt.Printf("%v thumbs \n", NumThumbs())
    fmt.Printf("%v jpg files\n", NumFilesInDir(thumbsPath))
    fmt.Printf("%v unique channels\n", NumUniqueChannels())
}

