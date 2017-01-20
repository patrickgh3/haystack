package main

import (
    "net/http"
    "os"
    "io"
)

// DownloadImage downloads and saves an image from a url to a path.
func DownloadImage(url string, filepath string) {
    // Download image
    response, err := http.Get(url)
    if err != nil {
        panic(err)
    }

    // Create file
    file, err := os.Create(filepath)
    defer file.Close()
    if err != nil {
        panic(err)
    }

    // Write image to file
    _, err = io.Copy(file, response.Body)
    if err != nil {
        panic(err)
    }
}
