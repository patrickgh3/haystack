package database

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "fmt"
    "time"
    "os"
    "github.com/patrickgh3/haystack/config"
)

const (
    QueryTypeStreams int = 1
    QueryTypeFollows int = 2
)

type Filter struct {
    gorm.Model
    Name        string `gorm:"size:50"`
    Subpath     string `gorm:"size:50"`
    QueryType   int
    QueryParam  string `gorm:"size:150"`
    LastUpdateTime  time.Time
}

type Follow struct {
    gorm.Model
    FilterID    uint
    StreamID    uint
}

type Stream struct {
    gorm.Model
    ChannelName         string  `gorm:"size:50"`
    ChannelDisplayName  string  `gorm:"size:50"`
    Title               string  `gorm:"size:150"`
    StartTime           time.Time
    LastUpdateTime      time.Time
    AverageViewers      float32
    NumThumbs           uint
}

type Thumb struct {
    gorm.Model
    StreamID    uint
    VOD         string `gorm:"size:50"`
    VODSeconds  int
    ImagePath   string
}

type Clip struct {
    gorm.Model
    ClipID          string
    StreamID        uint
    ClipCreatedAt   time.Time
    ImageUrl        string
    ClipUrl         string
}

var db *gorm.DB

// InitDB initializes the database.
func InitDB() {
    dataSourceName := fmt.Sprintf("%v:%v@/%v?parseTime=True&loc=Local",
            config.DB.User, config.DB.Pass, config.DB.DBName)
    var err error
    db, err = gorm.Open("mysql", dataSourceName)
    if err != nil {
        panic(err)
    }

    // Migrate the schema
    db.AutoMigrate(&Filter{}, &Follow{}, &Stream{}, &Thumb{}, &Clip{})
}

// AddThumbToDB adds a new Thumb to the DB, possibly creating a new Stream.
func AddThumbToDB(roundTime time.Time, ChannelName string,
        ChannelDisplayName string, VODSeconds int, VOD string,
        ImagePath string, StartTime time.Time, Title string, viewers int,
        FilterIds []uint) {
    // Find most recently updated stream of this channel
    var foundStream *Stream
    var s []Stream
    db.Where("channel_name = ?", ChannelName).
            Order("last_update_time desc").Find(&s)
    if len(s) != 0 {
        // The stream is valid if it was last updated within the cutoff
        cutoff := roundTime.Add(-config.Timing.CutoffLeeway)
        if (s[0].LastUpdateTime.After(cutoff) ||
                s[0].LastUpdateTime.Equal(cutoff)) {
            foundStream = &s[0]
        }
    }

    // Create a stream for this thumb if none already exist,
    // or update current stream if it does exist
    var streamid uint
    if foundStream == nil {
        newstream := Stream{
                ChannelName:ChannelName, ChannelDisplayName:ChannelDisplayName,
                StartTime:StartTime, LastUpdateTime:roundTime, Title:Title,
                AverageViewers:float32(viewers), NumThumbs:1}
        db.Create(&newstream)
        streamid = newstream.ID
    } else {
        streamid = foundStream.ID
        foundStream.LastUpdateTime = roundTime
        foundStream.NumThumbs++
        // Running average = average*(6/7) + newpoint*(1/7), assuming 7 thumbs
        foundStream.AverageViewers =
            float32(foundStream.AverageViewers) *
            float32(foundStream.NumThumbs-1) / float32(foundStream.NumThumbs) +
            float32(viewers) / float32(foundStream.NumThumbs)
        db.Save(&foundStream)
    }

    // Insert the thumb with ID of its stream
    db.Create(&Thumb{StreamID:streamid, VOD:VOD, VODSeconds:VODSeconds,
            ImagePath:ImagePath})

    // Associate the stream with filters that picked it up, if they aren't
    // already associated
    for _, filterId:= range FilterIds {
        var f Follow
        db.FirstOrCreate(&f, &Follow{FilterID:filterId, StreamID:streamid})
    }
}

// PruneOldStreams removes old streams along with their thumbs, follows, and
// image files.
func PruneOldStreams(roundTime time.Time) {
    // Find all streams started before cutoff
    // Remove all Thumbs of those streams (and delete their image files)
    // Remove all Follows corresponding to those streams
    // Remove the streams

    t := roundTime
    oldestDay := time.Date(
            t.Year(), t.Month(), t.Day()-config.Timing.PruneDays,
            0, 0, 0, 0, t.Location())

    // TODO: Do this properly!
    var thumbs []Thumb
    db.Joins("join streams on streams.id = thumbs.stream_id AND streams.start_time < ?", oldestDay).Find(&thumbs)
    for _, thumb := range thumbs {
        DeleteImage(config.Path.Root + thumb.ImagePath)
        db.Unscoped().Delete(&thumb)
    }
    fmt.Printf("%v thumbs deleted\n", len(thumbs))

    var follows []Follow
    db.Joins("join streams on streams.id = follows.stream_id AND streams.start_time < ?", oldestDay).Find(&follows)
    for _, follow := range follows {
        db.Unscoped().Delete(&follow)
    }

    db.Table("streams").Unscoped().Where("start_time < ?", oldestDay).Delete(&Stream{})
}

// TODO: move this logic elsewhere?
func DeleteImage(filepath string) {
    err := os.Remove(filepath)
    if err != nil {
        fmt.Printf("couldn't remove image %v\n", filepath)
    }
}

// GetStreamThumbs returns all thumbs corresponding to a stream id.
func GetStreamThumbs(streamId uint) []Thumb {
    var thumbs []Thumb
    db.Where("stream_id = ?", streamId).Order("id asc").Find(&thumbs)
    return thumbs
}

func GetStreamByID(streamId uint) Stream {
    var stream Stream
    db.First(&stream, streamId)
    return stream
}

func GetAllStreams() []Stream {
    var streams []Stream
    db.Find(&streams)
    return streams
}

func GetStreamsOfFilter(filterId uint) []Stream {
    var streams []Stream
    db.Joins("join follows on follows.stream_id = streams.id AND follows.filter_id = ?", filterId).
            Order("id desc").Find(&streams)
    return streams
}

// GetAllFilters returns all filters in the DB.
func GetAllFilters() []Filter {
    var filters []Filter
    db.Find(&filters)
    return filters
}

func UpdateFilter(filterId uint, roundTime time.Time) {
    db.Table("filters").Where("id = ?", filterId).Update("last_update_time",
            roundTime)
}

func GetFilterWithSubpath(Subpath string) *Filter {
    var filters []Filter
    db.Where("subpath = ?", Subpath).Find(&filters)
    if len(filters) > 0 {
        return &(filters[0])
    }
    return nil
}

// AddClipToDB inserts a clip into the database if it does not already exist.
func AddClipToDB(ClipID string, CreatedAt time.Time,
        ImageUrl string, ClipUrl string, ChannelName string) {
    var c []Clip
    db.Where("clip_id = ?", ClipID).Find(&c)
    if len(c) == 0 {
        fmt.Printf("Clip NOT found\n")

        // Find the most recent stream.
        var s []Stream
        db.Where("channel_name = ?", ChannelName).
                Order("last_update_time desc").Find(&s)
        if len(s) != 0 {
            // Insert clip.
            db.Create(&Clip{ClipID:ClipID, StreamID:s[0].ID,
                    ClipCreatedAt:CreatedAt, ImageUrl:ImageUrl, ClipUrl:ClipUrl})
        } else {
            fmt.Printf("OMG WTF we are adding a clip but we couldn't find a stream of that channel!!\n")
        }
    } else {
        fmt.Printf("Clip found\n")
    }
}

func GetStreamClips(streamId uint) []Clip {
    var clips []Clip
    db.Where("stream_id = ?", streamId).Order("id asc").Find(&clips)
    return clips
}

