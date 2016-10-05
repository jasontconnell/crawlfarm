package main

import (
	crawl "crawlfarm/common"
	"crawlfarm/worker/data"
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
	fmt.Println("Crawl Farm 1.0 - Worker")

	if file, err := os.OpenFile("config.json", os.O_RDONLY, os.ModePerm); err == nil {
		var conf data.Config
		dec := json.NewDecoder(file)
		dec.Decode(&conf)

		finished := make(chan bool)
		disconnect := make(chan bool)

		fmt.Println("Dialing server on ", conf.Server)

		if conn, err := net.Dial("tcp", conf.Server); err == nil {
			client := data.NewClient(conn)
			go handleClient(conn, client, finished, disconnect)
		} else {
			go func() {
				fmt.Println("No server listening at", conf.Server)
				finished <- true
			}()
		}

		<-finished
	}
}

func handleClient(conn net.Conn, client data.Client, finished, disconnect chan bool) {
	client.Site = crawl.ReadSite(conn, client.Gob, disconnect)
	fmt.Println("Job: ", client.Site.Root)

	go crawl.ReadLoop(conn, client.Gob, client.InLinks, client.Results, client.FoundLinks, disconnect)
	go crawl.WriteResults(conn, client.Gob, client.OutResults, disconnect)

	go func() {
		t := time.NewTicker(1 * time.Second)
		go func() {
			for tick := range t.C {
				fmt.Printf("\r%v Processed: %d. Found: %d. Unique: %d. Queue: %d\t\t", tick.Format("15:04:05"), client.ProcessedLinks, client.ReportedLinks, client.UniqueLinks, len(client.InLinks))
			}
		}()

		<-client.Finished
		t.Stop()

		fmt.Println("Job finished")
		finished <- true
	}()

	go func() {
		<-disconnect
		fmt.Println("\n\nServer closed. Closing")
		finished <- true
	}()

	for {
		select {
		case link := <-client.InLinks:
			reschan := crawl.Fetch(client.Site, link)
			for res := range reschan {
				client.Results <- res
			}
		case count := <-client.FoundLinks:
			client.UniqueLinks += count
		case result := <-client.Results:
			if result.JobFinished {
				client.Finished <- true
			} else {
				urls := crawl.Parse(client.Site, result.Link.Url, result.Content)

				for url := range urls {
					client.ReportedLinks++
					result.Links = append(result.Links, url)
				}

				result.Content = "" // clear content, the server doesn't care, save network bandwidth
				result.QueueLength = len(client.InLinks)
				client.ProcessedLinks++
				client.OutResults <- result
			}
		}
	}
}
