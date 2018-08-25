package data

import (
	crawl "github.com/jasontconnell/crawlfarm/common"
	"net"
)

type Client struct {
	Conn           net.Conn
	Site           crawl.Site
	InLinks        chan crawl.Link
	Results        chan crawl.Result
	OutResults     chan crawl.Result
	FoundLinks     chan int
	ReportedLinks  int
	ProcessedLinks int
	UniqueLinks    int
	Finished       chan bool
	Disconnect     chan bool
	Gob            crawl.Gob
}

func NewClient(conn net.Conn) (client Client) {
	client.Conn = conn
	client.InLinks = make(chan crawl.Link, crawl.UrlMaxLength)
	client.Results = make(chan crawl.Result, crawl.UrlMaxLength)
	client.OutResults = make(chan crawl.Result, crawl.UrlMaxLength)
	client.FoundLinks = make(chan int, crawl.UrlMaxLength)
	client.Finished = make(chan bool)
	client.Disconnect = make(chan bool)

	client.Gob = crawl.GetGob(client.Conn)

	return
}
