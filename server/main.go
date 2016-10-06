package main

import (
	crawl "crawlfarm/common"
	"crawlfarm/server/data"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

func init() {
	crawl.RegisterGobs()
}

func main() {
	fmt.Println("Crawl Farm 1.0 - Server")

	if file, err := os.OpenFile("config.json", os.O_RDONLY, os.ModePerm); err == nil {
		var conf data.Config
		var site crawl.Site

		dec := json.NewDecoder(file)
		dec.Decode(&conf)
		site.Root = conf.Root
		site.Headers = conf.Headers

		fmt.Println("Job:", site.Root)

		server := data.NewServer(site)

		server.UnprocessedLinks <- crawl.Link{Url: conf.Root, Referrer: "/"}

		t := time.NewTicker(1 * time.Second)
		go func() {
			for tick := range t.C {
				fmt.Printf("\r%v Workers: %d. Url queue: %d. Results queue: %v. Processed: %d. Errors: %d\t\t", 
					tick.Format("15:04:05"), len(server.Workers), len(server.UnprocessedLinks), len(server.Results), len(server.CrawledUrls), *server.ErrorCount)
			}
		}()

		fmt.Println("Listen on", conf.Listen)

		if listener, err := net.Listen("tcp", conf.Listen); err == nil {
			go startListenLoop(listener, server)
		}

		<-server.Finished
		t.Stop()

		fmt.Println("\n\nFinished Job", site.Root)
	}
}

func clientConnected(conn net.Conn, server data.Server) {
	worker := data.NewWorker(conn)
	server.Connected(worker)

	crawl.WriteSite(worker.Conn, worker.Gob, server.Site, worker.Disconnect)
	go crawl.WriteLinks(worker.Conn, worker.Gob, worker.OutLinks, worker.Disconnect)
	go crawl.WriteUrlCount(worker.Conn, worker.Gob, worker.FoundLinks, worker.Disconnect)
	go crawl.ReadLoop(worker.Conn, worker.Gob, nil, server.Results, nil, worker.Disconnect)

	go func() {
		<-worker.Disconnect
		server.Disconnected(worker)
	}()

	go func() {
		<-worker.Finished

		server.WorkerFinished(worker)

		if len(server.Workers) == 0 {
			server.Finished <- true
		}
	}()

	for {
		select {
		case link := <-server.UnprocessedLinks:
			worker.SentLinks[link.Url] = link
			worker.OutLinks <- link
		case result := <-server.Results:
			server.MarkComplete(&worker, result.Link)

			worker.QueueLength = result.QueueLength
			found := 0

			if result.Code != 200 {
				server.RecordError(result)
			} else {
				for _, url := range result.Links {
					if server.AddUrl(url) {
						found++
					}
				}
				worker.FoundLinks <- found
			}

			if server.CheckFinished(worker) {
				worker.Finished <- true
			}
		}
	}

}

func startListenLoop(listener net.Listener, server data.Server) {
	for {
		if conn, err := listener.Accept(); err == nil {
			go clientConnected(conn, server)
		}
	}
}
