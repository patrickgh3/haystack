package database

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "fmt"
    "time"
    "github.com/patrickgh3/haystack/config"
)

type Stream struct {
    gorm.Model
    ChannelName         string  `gorm:"size:50"`
    ChannelDisplayName  string  `gorm:"size:50"`
    Title               string  `gorm:"size:150"`
    StartTime           time.Time
    LastUpdateTime      time.Time
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
    db.AutoMigrate(&Stream{}, &Thumb{})
}

func AddThumbToDB(roundTime time.Time, ChannelName string,
        ChannelDisplayName string, VODSeconds int, VOD string,
        ImagePath string, StartTime time.Time, Title string) {
    // Find existing stream for this new thumb, if there is one
    var s []Stream
    cutoff := roundTime.Add(-config.Timing.CutoffLeeway)
    db.Where("channel_name = ? AND last_update_time >= ?",
             ChannelName, cutoff).Find(&s)

    // Create a stream for this thumb if none already exist,
    // or set update time if it does exist
    var streamid uint
    if len(s) == 0 {
        newstream := Stream{
                ChannelName:ChannelName, ChannelDisplayName:ChannelDisplayName,
                StartTime:StartTime, LastUpdateTime:roundTime, Title:Title}
        db.Create(&newstream)
        streamid = newstream.ID
    } else {
        streamid = s[0].ID
        s[0].LastUpdateTime = roundTime
        db.Save(&s[0])
    }

    // Insert the thumb with ID of its stream
    db.Create(&Thumb{StreamID:streamid, VOD:VOD, VODSeconds:VODSeconds,
            ImagePath:ImagePath})
}

func GetStreamThumbs(streamId uint) []Thumb {
    var thumbs []Thumb
    db.Where("stream_id = ?", streamId).Find(&thumbs)
    return thumbs
}

func GetAllStreams() []Stream {
    var streams []Stream
    db.Find(&streams)
    return streams
}
