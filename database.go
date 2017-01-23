package main

import (
    "fmt"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "time"
    "os"
)

var db *sql.DB

const mysqlTimeFormat = "2006-01-02 15:04:05"

// InitDB creates the Database object in the package db variable.
func InitDB () error {
    // Don't open new database if db is already set
    if db != nil {
        return nil
    }

    // Open database
    dataSourceName := fmt.Sprintf("%v:%v@/%v", dbUser, dbPass, dbDatabase)
    var err error
    db, err = sql.Open("mysql", dataSourceName)
    if err != nil {
        panic(err)
    }

    // Verify the DB is working
    if err := db.Ping(); err != nil {
        panic(err)
    }

    return nil
}

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

type ThumbRow struct {
    created string
    channel string
    VOD string
    time string
    image string
}

func CurrRowStruct(rows *sql.Rows) *ThumbRow {
    r := new(ThumbRow)
    rows.Scan(&r.created, &r.channel, &r.VOD, &r.time, &r.image)
    return r
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
            fmt.Println("Error removing old image file")
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

