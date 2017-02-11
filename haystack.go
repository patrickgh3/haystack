package main

import (
    "time"
    "strconv"
    "fmt"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
)

// main is the entry point into the program.
// It initializes stuff and then calls Update periodically.
func main () {
    config.ReadConfig()
    database.InitDB()

    // Sleep until the next multiple of refresh period
    now := time.Now()
    wakeup := now.Add(config.RefreshDuration).Truncate(config.RefreshDuration)
    fmt.Print("Waiting...")
    time.Sleep(wakeup.Sub(now))
    fmt.Println("Go")

    // Initial update
    Update()

    // Start periodic updates
    ticker := time.NewTicker(config.RefreshDuration)
    for {
        <-ticker.C
        Update()
    }
}

// Update saves thumbnails of Twitch streams, deletes old ones, and
// builds a new webpage.
func Update () {
    roundTime := time.Now().Round(config.RefreshDuration)
    unixTimeString := strconv.FormatInt(roundTime.Unix(), 10);

    sr := TwitchAPIAllStreams("?game=I%20Wanna%20Be%20The%20Guy")
    //sr := TwitchAPIAllStreams(
    //        "?community_id=e7912cf2-1f61-46bd-91f8-9187fde84971")

    for _, s := range sr.Streams {
        fmt.Printf("%v...\n", s.Channel.Display_name)

        channelName := s.Channel.Display_name
        imageUrl := s.Preview.Medium

        subpath := config.ImagesSubdir + "/" +
                s.Channel.Name + "_" + unixTimeString + ".jpg"
        path := config.OutPath + "/" + subpath

        DownloadImage(imageUrl, path)

        archive := TwitchAPIChannelRecentArchive(s.Channel.Id)
        vodID := ""
        vodTime := time.Time{}
        if archive == nil {
            fmt.Println("WARN: archive was nil")
        } else {
            vodID = string(archive.Id)[1:]
            vodCreateTime := archive.Created_At_Time
            vodTime = time.Time{}.Add(roundTime.Sub(vodCreateTime))
        }

        database.InsertThumb(channelName, roundTime, vodID, subpath, vodTime)
    }

    numDeleted := database.DeleteOldThumbs(roundTime)

    BuildWebpage(roundTime)

    fmt.Printf("%v deleted\n", numDeleted)
    fmt.Printf("%v thumbs \n", database.NumThumbs())
    fmt.Printf("%v jpg files\n", NumFilesInDir(config.ThumbsPath))
    fmt.Printf("%v distinct channels\n", len(database.DistinctChannels()))
}

