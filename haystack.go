package main

import (
    "time"
    "strconv"
    "fmt"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
    "github.com/patrickgh3/haystack/twitchapi"
    "github.com/patrickgh3/haystack/webpage"
    "github.com/patrickgh3/haystack/webserver"
)

// main initializes stuff, then calls Update periodically.
func main () {
    // Load app configuration
    config.ReadConfig()
    // Initialize database
    database.InitDB()
    // Set up webpage stuff
    webpage.InitTemplate()

    // Rebuild the page upon new startup (useful for debugging)
    /*now := time.Now()
    roundTime := now.Round(config.Timing.Period)
    //database.DeleteOldThumbs(roundTime)
    webpage.BuildWebpage(roundTime)*/

    // Start web server to handle HTTP requets
    go webserver.Serve()

    // Start tracking streams
    TrackStreams()
}

// TrackStreams blocks and periodically grabs stream data from Twitch.
func TrackStreams() {
    // Sleep until the time is a multiple of the refresh period
    now := time.Now()
    wakeUpTime := now.Truncate(config.Timing.Period).Add(config.Timing.Period)
    fmt.Print("Waiting...")
    time.Sleep(wakeUpTime.Sub(now))
    fmt.Println("Go")

    // Start periodic updates
    ticker := time.NewTicker(config.Timing.Period)
    Update() // Update immediately, since ticker waits for next interval
    for {
        <-ticker.C
        Update()
    }
}

// Update grabs new thumbnails, deletes old ones, and generates the webpage.
func Update () {
    // Compute the current time rounded to the interval
    roundTime := time.Now().Round(config.Timing.Period)
    unixTimeString := strconv.FormatInt(roundTime.Unix(), 10);

    // Query Twitch for all the currently live streams
    sr := twitchapi.AllStreams("?game=I%20Wanna%20Be%20The%20Guy")

    // Iterate over all currently live streams, saving their info to the DB
    for _, stream := range sr.Streams {
        channelName := stream.Channel.Display_name
        status := stream.Channel.Status
        fmt.Printf("%v: %v\n", channelName, status)

        // Query Twitch for the channel's most recent (current) archive video ID
        archive := twitchapi.ChannelRecentArchive(stream.Channel.Id)
        var vodID string
        var vodOffset time.Duration
        if archive == nil {
            fmt.Println("WARN: archive was nil")
            vodID = ""
            vodOffset = time.Duration(0)
        } else {
            vodID = archive.Id
            vodOffset = roundTime.Sub(archive.Created_At_Time)
        }

        imagePath := config.Path.ImagesRelative + "/" +
                stream.Channel.Name + "_" + unixTimeString + ".jpg"
        imageDLPath := config.Path.Root + imagePath
        imageUrl := stream.Preview.Medium

        // Download stream preview image from Twitch
        DownloadImage(imageUrl, imageDLPath)

        // Finally, store new info for this stream in the DB
        // If VOD ID in streams table, add thumb to that
        // Else (new stream just started) add stream and this thumb to it

        // Want: DB Stream{ChannelName, ChannelDisplayName, Title}
        //       DB Thumb{StreamID, VODSeconds, VOD, ImagePath, Viewers}

        vodSeconds := int(vodOffset.Seconds())
        database.AddThumbToDB(
                stream.Channel.Name, stream.Channel.Display_name,
                vodSeconds, vodID, imagePath)

        //database.InsertThumb(
        //    channelName, roundTime, vodID, imagePath, vodTime, status)
    }

    // Prune old streams from the DB
    // For all streams with created < cutoff, delete, and delete their thumbs
    //numDeleted := database.DeleteOldThumbs(roundTime)

    // Regenerate the main webpage
    webpage.BuildWebpage(roundTime)

    /*fmt.Printf("%v deleted\n", numDeleted)
    fmt.Printf("%v thumbs \n", database.NumThumbs())
    fmt.Printf("%v jpg files\n", NumFilesInDir(config.Path.Images))
    fmt.Printf("%v distinct channels\n", len(database.DistinctChannels()))*/
}

