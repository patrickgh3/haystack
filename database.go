package main

import (
    "fmt"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "time"
)

var db *sql.DB

const mysqlTimeFormat = "2006-01-02 15:04:05"
const thumbDeleteDuration = time.Duration(-30) * time.Second

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
func InsertThumb (channelName string, vodID string, imageUrl string) {
    // MySQL can auto-populate the timestamp field, but let's do it
    // manually ourselves
    timeString := time.Now().Format(mysqlTimeFormat)
    _, err := db.Exec(
            "INSERT INTO thumbs (created, channel, VOD, imageUrl)"+
            "VALUES (?, ?, ?, ?)",
            timeString, channelName, vodID, imageUrl,
    )
    if err != nil {
        panic(err)
    }
}

// DeleteOldThumbs deletes thumbs entries older than a certain time.
func DeleteOldThumbs() {
    cutoffTime := time.Now().Add(thumbDeleteDuration)
    timeString := cutoffTime.Format(mysqlTimeFormat)
    _, err := db.Exec(
            "DELETE FROM thumbs WHERE created < (?)",
            timeString,
    )
    if err != nil {
        panic(err)
    }
}

