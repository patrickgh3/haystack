package main

import (
    "fmt"
    "net/http"
    "encoding/json"
)

type StreamsResponse struct {
    Total int `json:"_total"`
    Streams []*Stream
}

type StreamResponse struct {
    Stream *Stream
}

type Stream struct {
    Id int `json:"_id"`
    Channel *Channel
    Preview *Preview
}

type Channel struct {
    Display_name string
}

type Preview struct {
    Medium string
}

// TwitchAPIAllStreams returns all streams which match a given query.
// https://dev.twitch.tv/docs/v5/reference/streams/#get-all-streams
func TwitchAPIAllStreams (queryString string) *StreamsResponse {
    url := fmt.Sprintf(
            "https://api.twitch.tv/kraken/streams/%v",
            queryString,
    )
    sr := new(StreamsResponse)
    TwitchAPIGeneralQuery(url, &sr)
    return sr
}

// TwitchAPIStreamByChannel returns a stream corresponding to a channel ID.
// https://dev.twitch.tv/docs/v5/reference/streams/#get-stream-by-channel
func TwitchAPIStreamByChannel (channelID int) *StreamResponse {
    url := fmt.Sprintf(
            "https://api.twitch.tv/kraken/streams/%v",
            channelID,
    )
    sr := new(StreamResponse)
    TwitchAPIGeneralQuery(url, &sr)
    return sr
}

// TwitchAPIGeneralQuery performs an API query and parses the JSON response.
func TwitchAPIGeneralQuery (url string, v interface{}) {
    // Create client and request
    client := &http.Client{}
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
        panic(fmt.Sprintf("Bad HTTP status code: %v", response.StatusCode))
    }

    // Parse response JSON into struct
    defer response.Body.Close()
    decoder := json.NewDecoder(response.Body)
    err = decoder.Decode(&v)
    if err != nil {
        panic(err)
    }
}

