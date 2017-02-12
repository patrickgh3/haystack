package config

import (
    "github.com/go-yaml/yaml"
    "os"
    "io/ioutil"
    "github.com/kardianos/osext"
    "time"
)

const configFilename = "config.yaml"

var config struct {
    Path *ConfPath `yaml:"path"`
    Twitch *ConfTwitch `yaml:"twitch"`
    DB *ConfDB `yaml:"db"`
    Timing *ConfTiming `yaml:"timing"`
}

type ConfPath struct {
    Root string `yaml:"root"`
    ImagesRelative string `yaml:"images-relative"`
    SiteUrl string `yaml:"site-url"`
    Images string
}

type ConfTwitch struct {
    ClientKey string `yaml:"client-key"`
}

type ConfDB struct {
    User string `yaml:"user"`
    Pass string `yaml:"pass"`
    DBName string `yaml:"dbname"`
}

type ConfTiming struct {
    PeriodSeconds int `yaml:"period-seconds"`
    LabelSeconds int `yaml:"label-seconds"`
    NumPeriods int `yaml:"num-periods"`
    Period time.Duration
    LabelPeriod time.Duration
}

var Path ConfPath
var Twitch ConfTwitch
var DB ConfDB
var Timing ConfTiming

// ReadConfig populates config structs from the config file.
func ReadConfig() {
    ef, err := osext.ExecutableFolder()
    if err != nil {
        panic(err)
    }
    reader, err := os.Open(ef + "/" + configFilename)
    if err != nil {
        panic(err)
    }
    buf, err := ioutil.ReadAll(reader)
    if err != nil {
        panic(err)
    }

    config.Path = &Path
    config.Twitch = &Twitch
    config.DB = &DB
    config.Timing = &Timing

    yaml.Unmarshal(buf, &config)

    // Additional calculation
    Timing.Period = time.Duration(Timing.PeriodSeconds) * time.Second
    Timing.LabelPeriod = time.Duration(Timing.LabelSeconds) * time.Second
    Path.Images = Path.Root + Path.ImagesRelative
}

