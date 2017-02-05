package main

import (
    "time"
    "os"
    "fmt"
)

// InsertThumb adds a specified entry to the thumbs table.
func InsertThumb (channelName string, created time.Time,
        vodID string, imagePath string) {

    timeString := created.Format(mysqlTimeFormat)
    _, err := db.Exec(
            "INSERT INTO thumbs (created, channel, VOD, imagePath)"+
            "VALUES (?, ?, ?, ?)",
            timeString, channelName, vodID, imagePath,
    )
    if err != nil {
        panic(err)
    }
}

// NumThumbs returns the number of rows in the thumbs table.
func NumThumbs () int {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM thumbs").Scan(&count)
    if err != nil {
        panic(err)
    }
    return count
}

// DistinctChannels returns a list of unique channel names.
func DistinctChannels () []string {
    var channels []string

    rows, err := db.Query(
            "SELECT DISTINCT channel FROM thumbs ORDER BY channel")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        var c string
        err := rows.Scan(&c)
        if err != nil {
            panic(err)
        }
        channels = append(channels, c)
    }

    return channels
}

// ChannelThumbs returns a list of all thumbs belonging to a given channel.
func ChannelThumbs (channel string) []ThumbRow {
    var thumbs []ThumbRow

    rows, err := db.Query("SELECT * FROM thumbs WHERE channel=(?)", channel)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        r := CurrRowStruct(rows)
        thumbs = append(thumbs, *r)
    }

    return thumbs
}

// DeleteOldThumbs deletes thumbs entries older than a certain time.
func DeleteOldThumbs(roundTime time.Time) int {
    cutoffTime := roundTime.Add(thumbDeleteDuration)
    timeString := cutoffTime.Format(mysqlTimeFormat)

    // Delete image files of matching thumbs
    rows, err := db.Query("SELECT * FROM thumbs WHERE created <= (?)",
            timeString)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        r := CurrRowStruct(rows)
        filepath := outPath + r.Image
        err = os.Remove(filepath)
        if err != nil {
            fmt.Println("Error removing old thumb image file")
            fmt.Println(err)
        }
    }

    // Delete rows
    result, err := db.Exec(
            "DELETE FROM thumbs WHERE created <= (?)",
            timeString,
    )
    if err != nil {
        panic(err)
    }

    num, err := result.RowsAffected()
    if err != nil {
        panic(err)
    }
    return int(num)
}

