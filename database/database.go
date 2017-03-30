package database

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "fmt"
    "time"
    "github.com/patrickgh3/haystack/config"
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
    CreatedTime time.Time
    VODTimeTime time.Time
    Status string
}

// InitDB initializes the database.
func InitDB () {
    // Don't open new database if db is already set
    if db != nil {
        return
    }

    dataSourceName := fmt.Sprintf("%v:%v@/%v",
            config.DB.User, config.DB.Pass, config.DB.DBName)
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

func CurrRowStruct(rows *sql.Rows) (*ThumbRow, error) {
    r := new(ThumbRow)
    rows.Scan(&r.Id, &r.Created, &r.Channel, &r.VOD, &r.Image, &r.VODTime, &r.Status)

    t, err := time.Parse(mysqlTimestampFormat, r.Created)
    if err != nil {
        panic(err)
    }
    r.CreatedTime = t

    t, err = time.Parse(mysqlTimeFormat, r.VODTime)
    if err != nil {
        return nil, err
    }
    r.VODTimeTime = t

    return r, nil
}

