package database

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "fmt"
    "github.com/patrickgh3/haystack/config"
)

type Stream struct {
    gorm.Model
    ChannelName         string  `gorm:"size:50"`
    ChannelDisplayName  string  `gorm:"size:50"`
    VOD                 string  `gorm:"size:50"`
    //StartTime           time.Time

    //Title               string  `gorm:"size:150"`
    //EndTime             time.Time
}

type Thumb struct {
    gorm.Model
    StreamID    uint
    VODSeconds  int
    ImagePath   string
    VOD string `gorm:"-"`
    //Viewers     int
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

func AddThumbToDB(ChannelName string, ChannelDisplayName string,
        VODSeconds int, VOD string, ImagePath string) {
    // Find existing 
    var s []Stream
    db.Where("VOD = ?", VOD).Find(&s)
    var streamid uint
    if len(s) == 0 {
        fmt.Printf("adding stream\n")
        newstream := Stream{
                ChannelName:ChannelName, ChannelDisplayName:ChannelDisplayName,
                VOD:VOD}
        db.Save(&newstream)
        streamid = newstream.ID
    } else {
        fmt.Printf("not adding stream\n")
        streamid = s[0].ID
    }

    db.Create(&Thumb{
            VODSeconds:VODSeconds, ImagePath:ImagePath,
            StreamID:streamid})
}

func GetStreamThumbs(streamId uint) []Thumb {
    var stream Stream
    db.First(&stream, streamId) // uses primary key
    var thumbs []Thumb
    db.Where("stream_id = ?", streamId).Find(&thumbs)
    for i, _ := range thumbs {
        thumbs[i].VOD = stream.VOD
    }
    return thumbs
}

func GetAllStreams() []Stream {
    var streams []Stream
    db.Find(&streams)
    return streams
}
