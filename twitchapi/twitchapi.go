package twitchapi

import (
    "net/http"
    "encoding/json"
    "fmt"
    "time"
    "strconv"
    "github.com/patrickgh3/haystack/config"
)

// Custom marshal type since Twitch is sometimes inconsistent with ints/strings
// for IDs.
// http://stackoverflow.com/questions/31625511/
type jsonInt int
func (i *jsonInt) UnmarshalJSON(data []byte) error {
    if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
        data = data[1 : len(data)-1]
    }

    var tmp int
    err := json.Unmarshal(data, &tmp)
    if err != nil {
        return err
    }

    *i = jsonInt(tmp)
    return nil
}

type StreamsResponse struct {
    Total   int `json:"_total"`
    Streams []*Stream
}

type StreamResponse struct {
    Stream *Stream
}

type Stream struct {
    IdInt   int `json:"_id"`
    Id      string `json:"-"`
    Channel *Channel
    Preview *Preview
    Viewers int
}

type Channel struct {
    IdInt           jsonInt `json:"_id"`
    Id              string `json:"-"`
    Display_name    string
    Name            string
    Status          string
}

type Preview struct {
    Medium  string
}

type VideosResponse struct {
    Total   int `json:"_total"`
    Videos  []*Video
}

type Video struct {
    Id              string `json:"_id"`
    Broadcast_IdInt int `json:"broadcast_id"`
    Broadcast_Id    string
    Created_At      string
    Created_At_Time time.Time `json:"-"`
}

type ClipsResponse struct {
    Clips []*Clip
}

type Clip struct {
    TrackingId      string `json:"tracking_id"`
    Url             string
    Created_At      string
    Created_At_Time time.Time
    Thumbnails      Thumbnails
}

type Thumbnails struct {
    Small string
}

const videoTimeString = "2006-01-02T15:04:05Z"

// These "convert struct types" functions perform type converions
// into useful forms for IDs, times, etc.
func convertChannelTypes(channel *Channel) {
    if channel != nil {
        channel.Id = strconv.Itoa(int(channel.IdInt))
    }
}

func convertStreamTypes(stream *Stream) {
    if stream != nil {
        stream.Id = strconv.Itoa(stream.IdInt)
        convertChannelTypes(stream.Channel)
    }
}

func convertVideoTypes(video *Video) {
    if video != nil {
        video.Broadcast_Id = strconv.Itoa(video.Broadcast_IdInt)
        t, err := time.Parse(videoTimeString, video.Created_At)
        if err != nil {
            panic(err)
        }
        video.Created_At_Time = t
        video.Id = video.Id[1:] // Strip leading "v"
    }
}

func convertClipTypes(clip *Clip) {
    if clip != nil {
        t, err := time.Parse(videoTimeString, clip.Created_At)
        if err != nil {
            panic(err)
        }
        clip.Created_At_Time = t
    }
}

// AllStreams returns all streams which match a given query.
// See https://dev.twitch.tv/docs/v5/reference/streams/#get-all-streams
func AllStreams (queryString string) *StreamsResponse {
    urlTail := fmt.Sprintf("/streams/%v", queryString)
    r := new(StreamsResponse)
    generalQuery(urlTail, "", &r)

    for _, stream := range r.Streams {
        convertStreamTypes(stream)
    }
    return r
}

// ChannelVideos returns the videos from a channel.
// See https://dev.twitch.tv/docs/v5/reference/channels/#get-channel-videos
func ChannelVideos (channelID string, queryString string) *VideosResponse {
    urlTail := fmt.Sprintf("/channels/%v/videos%v", channelID, queryString)
    r := new(VideosResponse)
    generalQuery(urlTail, "", &r)

    for _, video := range r.Videos {
        convertVideoTypes(video)
    }
    return r
}

// ChannelRecentArchive returns a channel's most recent archive video.
func ChannelRecentArchive (channelID string) *Video {
    vr := ChannelVideos (channelID,
            "?broadcast_type=archive&sort=time&limit=1")
    if vr.Total == 0 {
        return nil
    }
    return vr.Videos[0]
}

// StreamByChannel returns the stream of a specified channel.
// See https://dev.twitch.tv/docs/v5/reference/streams/#get-stream-by-user
func StreamByChannel (channelID string) *Stream {
    urlTail := fmt.Sprintf("/streams/%v", channelID)
    r := new(StreamResponse)
    generalQuery(urlTail, "", &r)
    convertStreamTypes(r.Stream)
    return r.Stream
}

// FollowedStreams returns all live streams a user follows.
// See https://dev.twitch.tv/docs/v5/reference/users/#get-user-follows
func FollowedStreams (authorization string) *StreamsResponse {
    urlTail := fmt.Sprintf("/streams/followed")
    r := new(StreamsResponse)
    generalQuery(urlTail, authorization, &r)
    for _, s := range r.Streams {
        convertStreamTypes(s)
    }
    return r
}

func TopClipsDay (channelID string) *ClipsResponse {
    urlTail := fmt.Sprintf("/clips/top?channel=%v&period=day", channelID)
    r := new(ClipsResponse)
    generalQuery(urlTail, "", &r)

    for _, clip := range r.Clips {
        convertClipTypes(clip)
    }
    return r
}

// generalQuery performs an API query and parses the JSON response.
func generalQuery (urlTail string, authorization string, v interface{}) {
    url := fmt.Sprintf(
            "https://api.twitch.tv/kraken%v",
            urlTail,
    )

    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        panic(err)
    }
    req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
    req.Header.Add("Client-ID", config.Twitch.ClientKey)
    if len(authorization) > 0 {
        req.Header.Add("Authorization", "OAuth "+authorization)
    }

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

