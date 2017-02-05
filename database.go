package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "fmt"
)

var db *sql.DB

const mysqlTimestampFormat = "2006-01-02 15:04:05"
const mysqlTimeFormat = "15:04:05"

type ThumbRow struct {
    Id string
    Created string
    Channel string
    VOD string
    Image string
    VODTime string
}

// InitDB creates the Database object in the package db variable.
func InitDB () {
    // Don't open new database if db is already set
    if db != nil {
        return
    }

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
}

func CurrRowStruct(rows *sql.Rows) *ThumbRow {
    r := new(ThumbRow)
    rows.Scan(&r.Id, &r.Created, &r.Channel, &r.VOD, &r.Image, &r.VODTime)
    return r
}

