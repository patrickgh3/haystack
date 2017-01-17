package main

// Twitch API Reference for Get Stream By Channel
// https://dev.twitch.tv/docs/v5/reference/streams/#get-stream-by-channel

import (
    "fmt"
    "net/http"
    "encoding/json"
    "os"
    "io"
)

// SaveStreamThumbnail downloads a thumbnail of a stream and saves it to file.
func SaveStreamThumbnail (channelId int) {
    // Create client and request
    client := &http.Client{}
    url := fmt.Sprintf("https://api.twitch.tv/kraken/streams/%v", channelId)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        panic(err)
    }
    // Set request headers
    req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
    req.Header.Add("Client-ID", apiClientId)

    // Make request
    response, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    if response.StatusCode != 200 {
        panic(fmt.Sprintf("Response status code %v", response.StatusCode))
    }

    // Parse structs from JSON
    defer response.Body.Close()
    s := new(StreamResponse)
    decoder := json.NewDecoder(response.Body)
    err = decoder.Decode(&s)
    if err != nil {
        panic(err)
    }

    // API definition for channel being offline
    if s.Stream == nil {
        panic("Stream is offline")
    }

    // Download and save image
    filename := fmt.Sprintf("/var/html/cwpat.me/misc/%v_thumb.jpg",
            s.Stream.Channel.Display_name)
    imageUrl := s.Stream.Preview.Medium

    response, err = http.Get(imageUrl)
    if err != nil {
        panic(err)
    }
    file, err := os.Create(filename)
    defer file.Close()
    if err != nil {
        panic(err)
    }
    _, err = io.Copy(file, response.Body)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Source: %v\nDest: %v\n", imageUrl, filename)
}

type StreamResponse struct {
    Stream *Stream
}
type Stream struct {
    Preview *Preview
    Channel *Channel
}
type Preview struct {
    Medium string
}
type Channel struct {
    Display_name string
}


