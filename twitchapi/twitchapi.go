// IMPORTANT NOTE
// IMPORTANT NOTE
// IMPORTANT NOTE
// Migrated to helix twitch API on 2022-02-14 and not all
// queries on this page are tested, such as user followed streams and clips stuff.
// See this guide for migration if we ever want to use those:
// https://dev.twitch.tv/docs/api/migration
// IMPORTANT NOTE
// IMPORTANT NOTE
// IMPORTANT NOTE

package twitchapi

import (
    "net/http"
    "encoding/json"
    "fmt"
    "time"
    "io"
    "github.com/patrickgh3/haystack/config"
    "strings"
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

//type StreamResponse struct {
//    Stream *Stream
//}

type GetStreamsResponse struct {
    Streams []*Stream `json:"data"`
}

type Stream struct {
    Id           string `json:"id"`
    ChannelId    string `json:"user_id"`
    Status       string `json:"title"`
    Viewers      int `json:"viewer_count"`
    Preview      string `json:"thumbnail_url"`
    ChannelName          string `json:"user_login"`
    ChannelDisplayName   string `json:"user_name"`
}

type GetVideosResponse struct {
    Videos []*Video `json:"data"`
}

type Video struct {
    Id              string `json:"_id"`
    Broadcast_Id    string `json:"stream_id"`
    Created_At      string
    Created_At_Time time.Time `json:"-"`
}

/*type ClipsResponse struct {
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
}*/

const videoTimeString = "2006-01-02T15:04:05Z"

// These "convert struct types" functions perform type converions
// into useful forms for IDs, times, etc.
func convertStreamTypes(stream *Stream) {
    if stream != nil {
        stream.Preview = strings.Replace(stream.Preview, "{width}", "320", 1)
        stream.Preview = strings.Replace(stream.Preview, "{height}", "180", 1)
    }
}

func convertVideoTypes(video *Video) {
    if video != nil {
        t, err := time.Parse(videoTimeString, video.Created_At)
        if err != nil {
            panic(err)
        }
        video.Created_At_Time = t
        video.Id = video.Id[1:] // Strip leading "v"
    }
}

/*func convertClipTypes(clip *Clip) {
    if clip != nil {
        t, err := time.Parse(videoTimeString, clip.Created_At)
        if err != nil {
            panic(err)
        }
        clip.Created_At_Time = t
    }
}*/

// AllStreams returns all streams which match a given query.
// See https://dev.twitch.tv/docs/v5/reference/streams/#get-all-streams
func AllStreams (queryString string) *GetStreamsResponse {
    urlTail := fmt.Sprintf("/streams%v", queryString)
    r := new(GetStreamsResponse)
    generalQuery(urlTail, config.Twitch.AppAccessToken, &r)

    for _, stream := range r.Streams {
        convertStreamTypes(stream)
    }
    return r
}

// ChannelVideos returns the videos from a channel.
// See https://dev.twitch.tv/docs/v5/reference/channels/#get-channel-videos
func GetVideos (queryString string) *GetVideosResponse {
    urlTail := fmt.Sprintf("/videos%v", queryString)
    r := new(GetVideosResponse)
    generalQuery(urlTail, config.Twitch.AppAccessToken, &r)

    for _, video := range r.Videos {
        convertVideoTypes(video)
    }
    return r
}

// ChannelRecentArchive returns a channel's most recent archive video.
func ChannelRecentArchive (channelID string) *Video {
    vr := GetVideos (fmt.Sprintf("?user_id=%v&type=archive&sort=time&first=1", channelID))
    if len(vr.Videos) == 0 {
        return nil
    }
    return vr.Videos[0]
}

// StreamByChannel returns the stream of a specified channel.
// See https://dev.twitch.tv/docs/v5/reference/streams/#get-stream-by-user
/*func StreamByChannel (channelID string) *Stream {
    urlTail := fmt.Sprintf("/streams/%v", channelID)
    r := new(StreamResponse)
    generalQuery(urlTail, config.Twitch.AppAccessToken, &r)
    convertStreamTypes(r.Stream)
    return r.Stream
}*/

// FollowedStreams returns all live streams a user follows.
// See https://dev.twitch.tv/docs/v5/reference/users/#get-user-follows
/*func FollowedStreams (authorization string) *StreamsResponse {
    urlTail := fmt.Sprintf("/streams/followed")
    r := new(StreamsResponse)
    generalQuery(urlTail, authorization, &r)
    for _, s := range r.Streams {
        convertStreamTypes(s)
    }
    return r
}*/

/*func TopClipsDay (channelID string) *ClipsResponse {
    urlTail := fmt.Sprintf("/clips/top?channel=%v&period=day", channelID)
    r := new(ClipsResponse)
    generalQuery(urlTail, config.Twitch.AppAccessToken, &r)

    for _, clip := range r.Clips {
        convertClipTypes(clip)
    }
    return r
}*/

// generalQuery performs an API query and parses the JSON response.
// IMPORTANT NOTE
// IMPORTANT NOTE
// IMPORTANT NOTE
// Migrated to helix twitch API on 2022-02-14 and not all
// queries on this page are tested, such as user followed streams and clips stuff.
// See this guide for migration if we ever want to use those:
// https://dev.twitch.tv/docs/api/migration
// IMPORTANT NOTE
// IMPORTANT NOTE
// IMPORTANT NOTE
func generalQuery (urlTail string, authorization string, v interface{}) {
    url := fmt.Sprintf(
            "https://api.twitch.tv/helix%v",
            urlTail,
    )

    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        panic(err)
    }
    req.Header.Add("Client-ID", config.Twitch.ClientKey)
    if len(authorization) > 0 {
        req.Header.Add("Authorization", "Bearer "+authorization)
    }

    // Make request
    response, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer response.Body.Close()
    if response.StatusCode != 200 {
        bodyBytes, err := io.ReadAll(response.Body)
        if err != nil {
            panic(err)
        }
        panic(fmt.Sprintf("Bad HTTP status code: %v body: %v", response.StatusCode, string(bodyBytes)))
    }

    // Parse response JSON into struct
    decoder := json.NewDecoder(response.Body)
    err = decoder.Decode(&v)
    if err != nil {
        panic(err)
    }
}

