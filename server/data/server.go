package data

import (
    crawl "crawlfarm/common"
    "os"
    "sync"
    "fmt"
    "log"
)

type Server struct {
    Site crawl.Site
    UnprocessedLinks chan crawl.Link
    Results chan crawl.Result
    Workers map[string]bool
    CrawledUrls map[string]string
    ProcessQueue map[string]string
    Finished chan bool
    ErrorCodes []int
    Mutex *sync.Mutex
    ErrorLog *log.Logger
    ProcessLog *log.Logger
    MessageLog *log.Logger
    ErrorCount *int
}

func NewServer(site crawl.Site, errcodes []int) (server Server){
    server.Site = site
    server.UnprocessedLinks = make(chan crawl.Link, 13000)
    server.Workers = make(map[string]bool)
    server.Results = make(chan crawl.Result, crawl.UrlMaxLength)
    server.CrawledUrls = make(map[string]string)
    server.ProcessQueue = make(map[string]string)
    server.Finished = make(chan bool)
    server.Mutex = new(sync.Mutex)
    server.ErrorCount = new(int)
    *server.ErrorCount = 0
    server.ErrorCodes = errcodes

    if file, err := os.Create("processed.log"); err == nil {
        server.ProcessLog = log.New(file, "", 0)
    }

    if file, err := os.Create("errors.log"); err == nil {
        server.ErrorLog = log.New(file, "", 0)
    }

    if file, err := os.Create("messages.log"); err == nil {
        server.MessageLog = log.New(file, "", 0)
    }

    return
}

func (server Server) MarkComplete(worker *Worker, link crawl.Link){
    server.ProcessLog.Println(link.Url)

    server.Mutex.Lock()
    defer server.Mutex.Unlock()


    if _,ok := server.CrawledUrls[link.Url]; !ok {
        server.CrawledUrls[link.Url] = link.Url
    }

    if _,ok := server.ProcessQueue[link.Url]; ok {
        delete(server.ProcessQueue, link.Url)
    }

    if _,ok := (*worker).SentLinks[link.Url]; ok {
        delete((*worker).SentLinks, link.Url)
    }
}

func (server Server) AddUrl(link crawl.Link) bool {
    server.Mutex.Lock()
    defer server.Mutex.Unlock()

    if _,ok := server.ProcessQueue[link.Url]; !ok {
        server.ProcessQueue[link.Url] = link.Url
        server.UnprocessedLinks <- link
        return true
    }
    return false
}

func (server Server) RecordError(result crawl.Result){
    server.ErrorLog.Println(fmt.Sprintf("%v, %v, %v", result.Link.Url, result.Link.Referrer, result.Code))

    server.Mutex.Lock()
    *server.ErrorCount++
    server.Mutex.Unlock()
}

func (server *Server) Connected(worker Worker){
    server.Mutex.Lock()
    defer server.Mutex.Unlock()

    server.Workers[worker.RemoteAddr] = true
}

func (server *Server) Disconnected(worker Worker){
    server.Mutex.Lock()
    defer server.Mutex.Unlock()

    for _, link := range worker.SentLinks {
        server.UnprocessedLinks <- link
    }

    if _,ok := server.Workers[worker.RemoteAddr]; ok {
        delete(server.Workers, worker.RemoteAddr)
    }
}

func (server *Server) WorkerFinished(worker Worker){
    server.MessageLog.Println("worker finished", worker.RemoteAddr)

    server.Mutex.Lock()
    defer server.Mutex.Unlock()

    crawl.WriteFinished(worker.Conn, worker.Gob, true, worker.Disconnect)

    if _,ok := server.Workers[worker.RemoteAddr]; ok {
        delete(server.Workers, worker.RemoteAddr)
    }
}

func (server Server) CheckFinished(worker Worker) bool {
    return len(server.ProcessQueue) == 0 && len(worker.SentLinks) == 0
}