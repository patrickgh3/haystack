package main

import (
    "io/ioutil"
)

// NumFilesInDir returns the number of files in a directory.
func NumFilesInDir (dirPath string) int {
    info, err := ioutil.ReadDir(dirPath)
    if err != nil {
        panic(err)
    }
    return len(info)
}

