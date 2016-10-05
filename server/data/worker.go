package data

import (
    "net"
    crawl "crawlfarm-common"
)

type Worker struct {
    RemoteAddr string
    Conn net.Conn
    Sites chan crawl.Site
    InLinks chan crawl.Link
    OutLinks chan crawl.Link
    Results chan crawl.Result
    FoundLinks chan int
    Disconnect chan bool

    SentLinks map[string]crawl.Link
    QueueLength int
    Finished chan bool

    Gob crawl.Gob
}

func NewWorker(conn net.Conn) (worker Worker){
    worker.Conn = conn
    worker.RemoteAddr = conn.RemoteAddr().String()
    worker.Sites = make(chan crawl.Site)
    worker.InLinks = make(chan crawl.Link, crawl.UrlMaxLength)
    worker.OutLinks = make(chan crawl.Link, crawl.UrlMaxLength)
    worker.Results = make(chan crawl.Result, crawl.UrlMaxLength)
    worker.FoundLinks = make(chan int, crawl.UrlMaxLength)
    worker.SentLinks = make(map[string]crawl.Link)
    worker.Disconnect = make(chan bool)
    worker.Finished = make(chan bool)
    worker.Gob = crawl.GetGob(worker.Conn)

    return
}