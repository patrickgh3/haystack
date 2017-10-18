package main

import (
    "time"
    "strconv"
    "fmt"
    "os"
    "bufio"
    "path"
    "github.com/patrickgh3/haystack/config"
    "github.com/patrickgh3/haystack/database"
    "github.com/patrickgh3/haystack/twitchapi"
    "github.com/patrickgh3/haystack/webserver"
)

// main initializes stuff, then calls Update periodically.
func main () {
    // Load app configuration
    config.ReadConfig()
    // Initialize database
    database.InitDB()
    // Set up webpage stuff
    webserver.InitTemplates()

    /*fmt.Print("Debug regenerating filters pages...")
    RegenerateFilterPages()
    fmt.Print("Done\n")*/

    // Start web server to handle HTTP requets
    // NOTE: removed "go"
    webserver.Serve()

    // Start tracking streams
    //TrackStreams()
}

func RegenerateFilterPages() {
    filters := database.GetAllFilters()
    for _, filter := range filters {
        wpd := webserver.FilterPageData(filter)

        dir := path.Join(config.Path.Root, filter.Subpath)
        err := os.Mkdir(dir, os.ModePerm)
        if err != nil && !os.IsExist(err) {
            panic(err)
        }
        f, err := os.Create(path.Join(dir, "index.html"))
        defer f.Close()
        if err != nil {
            panic(err)
        }
        w := bufio.NewWriter(f)
        webserver.WriteFilterPage(w, wpd)
        w.Flush()
    }
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

type taggedStream struct {
    Stream      *twitchapi.Stream
    FilterIds   []uint
}
func (ts *taggedStream) AppendFilter(f uint) {
    (*ts).FilterIds = append((*ts).FilterIds, f)
}

// Update grabs new thumbnails, deletes old ones, and generates the webpage.
func Update () {
    // Compute the current time rounded to the interval
    roundTime := time.Now().Round(config.Timing.Period)
    unixTimeString := strconv.FormatInt(roundTime.Unix(), 10);

    // Query Twitch for each filter's streams, and assemble a map of unique
    // streams along with all filters which found them
    taggedStreams := make(map[string]*taggedStream)
    filters := database.GetAllFilters()
    for _, filter := range filters {
        // Make appropriate Twitch query
        var sr *twitchapi.StreamsResponse
        if filter.QueryType == database.QueryTypeStreams {
            sr = twitchapi.AllStreams(filter.QueryParam)
        } else if filter.QueryType == database.QueryTypeFollows {
            sr = twitchapi.FollowedStreams(filter.QueryParam)
        }
        // Assimilate all recieved streams into our tagged stream map
        for _, s := range sr.Streams {
            id := s.Channel.Id
            if _, seen := taggedStreams[id]; !seen { // Idiom to check if in map
                taggedStreams[id] = &taggedStream{Stream:s}
            }
            taggedStreams[id].AppendFilter(filter.ID)
        }
        database.UpdateFilter(filter.ID, roundTime)
    }

    // For each (stream, filters) pair, grab and save a snapshot to the DB
    for _, sf := range taggedStreams {
        stream := sf.Stream
        channelName := stream.Channel.Display_name
        status := stream.Channel.Status
        fmt.Printf("(%v) %v: %v\n", len(sf.FilterIds), channelName, status)

        // Query Twitch for the channel's most recent (current) archive video ID
        archive := twitchapi.ChannelRecentArchive(stream.Channel.Id)

        // If this snapshot doesn't correspond to the most recent archive, then
        // either the streamer has disabled archiving, the archive somehow isn't
        // accessible yet, or something else. So, store no VOD for this thumb.
        vodID := ""
        vodSeconds := 0
        vodTime := roundTime
        if archive != nil {
            if archive.Broadcast_Id == stream.Id {
                vodID = archive.Id
                vodSeconds = int(roundTime.Sub(archive.Created_At_Time).Seconds())
                //vodTime = archive.Created_At_Time
            } else {
                fmt.Printf("recent archive is not current stream\n")
            }
        } else {
            fmt.Printf("recent archive was nil\n")
        }

        // Download stream preview image from Twitch
        imagePath := config.Path.ImagesRelative + "/" +
                stream.Channel.Name + "_" + unixTimeString + ".jpg"
        imageDLPath := config.Path.Root + imagePath
        imageUrl := stream.Preview.Medium
        DownloadImage(imageUrl, imageDLPath)

        // Finally, store new info for this stream in the DB
        database.AddThumbToDB(
                roundTime, stream.Channel.Name, stream.Channel.Display_name,
                vodSeconds, vodID, imagePath, vodTime, stream.Channel.Status,
                stream.Viewers, sf.FilterIds)
    }

    // Delete old streams (and their thumbs, follows, image files) from the DB
    database.PruneOldStreams(roundTime)

    RegenerateFilterPages()

    // TODO: Occasionally check for "stray" data:
    // (Also perform this check on app startup)
    // Follows whose stream or filter no longer exists
    //     Have to check for each one.
    // Thumbs whose stream no longer exists
    //     Total of streams NumThumbs != # thumbs
    // Image files whose thumb no longer exists
    //     # image files != # thumbs

    fmt.Printf("update finish\n")
    /*fmt.Printf("%v deleted\n", numDeleted)
    fmt.Printf("%v thumbs \n", database.NumThumbs())
    fmt.Printf("%v jpg files\n", NumFilesInDir(config.Path.Images))
    fmt.Printf("%v distinct channels\n", len(database.DistinctChannels()))*/
}

