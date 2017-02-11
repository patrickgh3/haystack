package config

import (
    "github.com/spf13/viper"
    "github.com/kardianos/osext"
    "path"
    "time"
)

var OutPath string
var ThumbsPath string
var SiteBaseUrl string
var ApiClientId string
var DbUser string
var DbPass string
var DbDatabase string
var RefreshDuration time.Duration
var NumRefreshPeriods int

const ImagesSubdir = "/images/t"
const VodBaseUrl = "https://www.twitch.tv/videos"

const configFilename = "config"

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

    ApiClientId = viper.GetString("client_id")
    DbUser = viper.GetString("db_user")
    DbPass = viper.GetString("db_pass")
    DbDatabase = viper.GetString("db_database")
    SiteBaseUrl = viper.GetString("base_url")
    OutPath = path.Clean(viper.GetString("out_path"))
    ThumbsPath = OutPath + ImagesSubdir

    seconds := viper.GetInt("period_seconds")
    RefreshDuration = time.Duration(seconds) * time.Second
    NumRefreshPeriods = viper.GetInt("num_periods")
}

