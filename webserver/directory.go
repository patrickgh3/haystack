package webserver

import (
    "net/http"
    "github.com/patrickgh3/haystack/database"
    "github.com/patrickgh3/haystack/config"
)

type dirPageData struct {
    Title string
    AppBaseUrl string
    Links []link
}

type link struct {
    Link string
    Name string
}

// ServeRootPage serves a page listing all filters.
func ServeRootPage(w http.ResponseWriter, r *http.Request) {
    var pd dirPageData
    pd.AppBaseUrl = config.Path.SiteUrl
    filters := database.GetAllFilters()
    for _, filter := range filters {
        pd.Links = append(pd.Links, link{
                Link:config.Path.SiteUrl+"/"+filter.Subpath, Name:filter.Name})
    }
    directoryTempl.Execute(w, pd)
}
