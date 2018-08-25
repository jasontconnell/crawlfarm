package main

import (
	"encoding/json"
	"fmt"
	crawl "github.com/jasontconnell/crawlfarm/common"
	"github.com/jasontconnell/crawlfarm/worker/data"
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

		fmt.Println("Dialing server on ", conf.Server)

		if conn, err := net.Dial("tcp", conf.Server); err == nil {
			client := data.NewClient(conn)
			go handleClient(conn, client)

			<-client.Finished
		} else {
			fmt.Println("No server listening at", conf.Server)
		}
	}
}

func startOutputLoop(client *data.Client) {
	t := time.NewTicker(1 * time.Second)
	go func() {
		for tick := range t.C {
			fmt.Printf("\r%v Processed: %d. Found: %d. Unique: %d. Queue: %d\t\t",
				tick.Format("15:04:05"), client.ProcessedLinks, client.ReportedLinks, client.UniqueLinks, len(client.InLinks))
		}
	}()

	<-client.Finished
	t.Stop()

	fmt.Println("Job finished")

	client.Finished <- true // send finished to the listener in main()
}

func handleClient(conn net.Conn, client data.Client) {
	client.Site = crawl.ReadSite(conn, client.Gob, client.Disconnect)
	fmt.Println("Job: ", client.Site.Root)

	go crawl.ReadLoop(conn, client.Gob, client.InLinks, client.Results, client.FoundLinks, client.Disconnect)
	go crawl.WriteResults(conn, client.Gob, client.OutResults, client.Disconnect)

	go startOutputLoop(&client)

	go func() {
		<-client.Disconnect
		fmt.Println("\n\nServer closed. Closing")
		client.Finished <- true
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
