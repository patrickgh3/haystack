package main

import (
    "time"
    "os"
    "fmt"
)

// InsertThumb adds a specified entry to the thumbs table.
func InsertThumb (channelName string, created time.Time,
        vodID string, imageUrl string) {

    timeString := created.Format(mysqlTimeFormat)
    _, err := db.Exec(
            "INSERT INTO thumbs (created, channel, VOD, imageUrl)"+
            "VALUES (?, ?, ?, ?)",
            timeString, channelName, vodID, imageUrl,
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

// NumUniqueChannels returns the number of unique channels.
func NumUniqueChannels () int {
    var count int
    err := db.QueryRow(
            "SELECT COUNT(DISTINCT channel) FROM thumbs").Scan(&count)
    if err != nil {
        panic(err)
    }
    return count
}

func UniqueChannels () []string {
    var channels []string

    rows, err := db.Query("SELECT DISTINCT channel FROM thumbs")
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

// DeleteOldThumbs deletes thumbs entries older than a certain time.
func DeleteOldThumbs() int {
    cutoffTime := time.Now().Add(ThumbDeleteDuration)
    timeString := cutoffTime.Format(mysqlTimeFormat)

    // Delete image files of matching thumbs
    rows, err := db.Query("SELECT * FROM thumbs WHERE created < (?)", timeString)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        r := CurrRowStruct(rows)
        err = os.Remove(r.image)
        if err != nil {
            fmt.Println("Error removing old thumb image file")
            fmt.Println(err)
        }
    }

    // Delete rows
    result, err := db.Exec(
            "DELETE FROM thumbs WHERE created < (?)",
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

