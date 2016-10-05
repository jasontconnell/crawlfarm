package crawlfarmCommon

import (
    "fmt"
)

const (
    UrlMaxLength = 500
)

type Packet struct {
    Link *Link
    Result *Result
    Site *Site
    Urls *int
}

type Link struct {
    Referrer string
    Url string
}

type Result struct {
    Link Link
    Code int
    Content string
    Links []Link
    QueueLength int
    JobFinished bool
}

type Site struct {
    Root string `json:"root"`
    Headers map[string]string `json:"headers"`
}

func (link Link) String() string {
    return fmt.Sprintf("Url: %v Referrer: %v", link.Url, link.Referrer)
}