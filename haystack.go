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

    // Sleep until the next multiple of refresh period
    now := time.Now()
    wakeupTime := now.Add(refreshDuration).Truncate(refreshDuration)
    fmt.Print("Waiting...")
    time.Sleep(wakeupTime.Sub(now))
    fmt.Println("Go")

    // Initial update
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
    roundTime := time.Now().Round(refreshDuration)
    unixTimeString := strconv.FormatInt(roundTime.Unix(), 10);

    sr := TwitchAPIAllStreams("?game=I%20Wanna%20Be%20The%20Guy")

    for _, s := range sr.Streams {
        fmt.Printf("%v...\n", s.Channel.Display_name)

        channelName := strings.ToLower(s.Channel.Display_name)
        imageUrl := s.Preview.Medium

        subpath := imagesSubdir + "/" +
                channelName + "_" + unixTimeString + ".jpg"
        path := outPath + "/" + subpath

        DownloadImage(imageUrl, path)

        archive := TwitchAPIChannelRecentArchive(s.Channel.Id)
        vodID := string(archive.Id)[1:]

        InsertThumb(channelName, roundTime, vodID, subpath)
    }

    numDeleted := DeleteOldThumbs(roundTime)

    BuildWebpage(roundTime)

    fmt.Printf("%v deleted\n", numDeleted)
    fmt.Printf("%v thumbs \n", NumThumbs())
    fmt.Printf("%v jpg files\n", NumFilesInDir(thumbsPath))
    fmt.Printf("%v distinct channels\n", len(DistinctChannels()))
}

