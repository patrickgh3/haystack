package database

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "fmt"
    "time"
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
    db.AutoMigrate(&Filter{}, &Follow{}, &Stream{}, &Thumb{})
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

// GetStreamThumbs returns all thumbs corresponding to a stream id.
func GetStreamThumbs(streamId uint) []Thumb {
    var thumbs []Thumb
    db.Where("stream_id = ?", streamId).Order("id asc").Find(&thumbs)
    return thumbs
}

func GetAllStreams() []Stream {
    var streams []Stream
    db.Find(&streams)
    return streams
}

func GetStreamsOfFilter(filterId uint) []Stream {
    var streams []Stream
    db.Joins("join follows on follows.stream_id = streams.id AND follows.filter_id = ?", filterId).Find(&streams)
    return streams
}

// GetAllFilters returns all filters in the DB.
func GetAllFilters() []Filter {
    var filters []Filter
    db.Find(&filters)
    return filters
}

func UpdateAllFilters(roundTime time.Time) {
    db.Table("filters").Update("last_update_time", roundTime)
}

func GetFilterWithSubpath(Subpath string) *Filter {
    var filters []Filter
    db.Where("subpath = ?", Subpath).Find(&filters)
    if len(filters) > 0 {
        return &(filters[0])
    }
    return nil
}

