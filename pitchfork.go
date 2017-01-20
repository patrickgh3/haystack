package main

import (
    "fmt"
    "time"
    "strconv"
    "strings"
)

// main is the entry point into the program.
// It initializes stuff and then calls Update periodically.
func main () {
    ReadConfig()

    InitDB()

    // Do an initial update
    // (useful to quickly verify it's working)
    Update()

    // Start periodic updates
    ticker := time.NewTicker(refreshDuration)
    for {
        select {
        case <- ticker.C:
            Update()
        }
    }
}

// Update saves thumbnails of Twitch streams, deletes old ones, and
// builds a new webpage.
func Update () {
    sr := TwitchAPIAllStreams("?game=I%20Wanna%20Be%20The%20Guy")
    for _, s := range sr.Streams[:] {
        fmt.Printf("%v...", s.Channel.Display_name)

        channelName := strings.ToLower(s.Channel.Display_name)
        vodID := strconv.Itoa(s.Id)
        imageUrl := s.Preview.Medium

        filepath := fmt.Sprintf("%v/images/%v_thumb.jpg",
                outPath,
                channelName,
        )
        DownloadImage(imageUrl, filepath)

        InsertThumb(channelName, vodID, filepath)

        fmt.Println("done")
    }

    DeleteOldThumbs()

    BuildPage()
}

