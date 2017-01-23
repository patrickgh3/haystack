package main

import (
    "github.com/spf13/viper"
    "github.com/kardianos/osext"
    "path"
    "time"
)

var outPath string
var thumbsPath string
var refreshDuration time.Duration
var apiClientId string
var dbUser string
var dbPass string
var dbDatabase string

const ThumbDeleteDuration = time.Duration(30) * time.Second * -1
const imagesSubdir = "/images/t"

// ReadConfig sets various variables from the config file.
func ReadConfig() {
    // Set config file location to current directory
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    viper.AddConfigPath(ef)
    viper.SetConfigName("config")

    // Read config values
    err = viper.ReadInConfig()
    if err != nil {
        panic(err)
    }

    outPath = viper.GetString("out_path")
    outPath = path.Clean(outPath)
    thumbsPath = outPath + imagesSubdir
    seconds := viper.GetInt("interval_seconds")
    refreshDuration = time.Duration(seconds) * time.Second
    apiClientId = viper.GetString("client_id")
    dbUser = viper.GetString("db_user")
    dbPass = viper.GetString("db_pass")
    dbDatabase = viper.GetString("db_database")
}

