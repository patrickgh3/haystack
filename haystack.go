package main

import (
    "time"
    "strconv"
    "fmt"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
    "github.com/patrickgh3/haystack/newdatabase"
    "github.com/patrickgh3/haystack/twitchapi"
    "github.com/patrickgh3/haystack/webpage"
    "github.com/patrickgh3/haystack/webserver"
)

// main initializes stuff, then calls Update periodically.
func main () {

    config.ReadConfig()

    // TEMP
    newdatabase.TestDB()
    //return

    database.InitDB()
    webpage.InitTemplate()

    // Initial page rebuild
    now := time.Now()
    roundTime := now.Round(config.Timing.Period)
    //database.DeleteOldThumbs(roundTime)
    webpage.BuildWebpage(roundTime)

    // TEMP: Start web server to handle HTTP requets
    // TODO: spawn as separate goroutine
    webserver.Serve()

    // Sleep until the next multiple of refresh period
    wakeup := roundTime.Add(config.Timing.Period)
    fmt.Print("Waiting...")
    time.Sleep(wakeup.Sub(now))
    fmt.Println("Go")

    // Periodic updates
    ticker := time.NewTicker(config.Timing.Period)
    Update() // Initial update
    for {
        <-ticker.C
        Update()
    }
}

// Update grabs new thumbnails, deletes old ones, and generates the webpage.
func Update () {
    roundTime := time.Now().Round(config.Timing.Period)
    unixTimeString := strconv.FormatInt(roundTime.Unix(), 10);

    sr := twitchapi.AllStreams("?game=I%20Wanna%20Be%20The%20Guy")
    //sr := TwitchAPIAllStreams(
    //        "?community_id=e7912cf2-1f61-46bd-91f8-9187fde84971")

    for _, stream := range sr.Streams {
        fmt.Printf("%v...\n", stream.Channel.Display_name)

        imagePath := config.Path.ImagesRelative + "/" +
                stream.Channel.Name + "_" + unixTimeString + ".jpg"
        channelName := stream.Channel.Display_name
        status := stream.Channel.Status
        fmt.Printf("status: %v\n", status)

        archive := twitchapi.ChannelRecentArchive(stream.Channel.Id)
        var vodID string
        var vodTime time.Time
        if archive == nil {
            fmt.Println("WARN: archive was nil")
            vodID = ""
            vodTime = time.Time{}
        } else {
            vodID = archive.Id
            vodTime = time.Time{}.Add(roundTime.Sub(archive.Created_At_Time))
        }

        imageDLPath := config.Path.Root + imagePath
        imageUrl := stream.Preview.Medium

        DownloadImage(imageUrl, imageDLPath)

        database.InsertThumb(
            channelName, roundTime, vodID, imagePath, vodTime, status)
    }

    numDeleted := database.DeleteOldThumbs(roundTime)

    webpage.BuildWebpage(roundTime)

    fmt.Printf("%v deleted\n", numDeleted)
    fmt.Printf("%v thumbs \n", database.NumThumbs())
    fmt.Printf("%v jpg files\n", NumFilesInDir(config.Path.Images))
    fmt.Printf("%v distinct channels\n", len(database.DistinctChannels()))
}

