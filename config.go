package main

import (
    "github.com/spf13/viper"
    "github.com/kardianos/osext"
    "path"
    "time"
)

var outPath string
var thumbsPath string
var siteBaseUrl string
var refreshDuration time.Duration
var apiClientId string
var dbUser string
var dbPass string
var dbDatabase string

const thumbDeleteDuration = time.Duration(30) * time.Second * -1
const imagesSubdir = "/images/t"
const configFilename = "config"
const vodBaseUrl = "https://www.twitch.tv/videos"

// ReadConfig sets various variables from the config file.
func ReadConfig() {
    viper.SetConfigName(configFilename)
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    viper.AddConfigPath(ef)

    err = viper.ReadInConfig()
    if err != nil {
        panic(err)
    }

    apiClientId = viper.GetString("client_id")
    dbUser = viper.GetString("db_user")
    dbPass = viper.GetString("db_pass")
    dbDatabase = viper.GetString("db_database")
    siteBaseUrl = viper.GetString("base_url")
    outPath = viper.GetString("out_path")
    outPath = path.Clean(outPath)
    thumbsPath = outPath + imagesSubdir
    seconds := viper.GetInt("interval_seconds")
    refreshDuration = time.Duration(seconds) * time.Second
}

