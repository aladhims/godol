package main

import (
	"flag"
	"log"
	"time"

	"github.com/aladhims/godol"
)

const (
	defaultWorker = 10
)

func main() {
	url := flag.String("url", "", "URL of the file that will be downloaded")
	destination := flag.String("dest", "./", "the path where the downloaded file will be placed")
	fileName := flag.String("name", "", "custom name for the file that will be downloaded")
	workers := flag.Int("worker", defaultWorker, "number of workers that will be spawned to complete the task")

	flag.Parse()

	var options []godol.Option

	if *url == "" {
		log.Fatal("URL should not be empty")
	}

	options = append(options, godol.WithURL(*url))
	options = append(options, godol.WithDestination(*destination))
	options = append(options, godol.WithFilename(*fileName))
	options = append(options, godol.WithWorker(*workers))

	g := godol.New(options...)

	timeStart := time.Now()

	log.Println("Download started")

	err := g.Start()
	if err != nil {
		log.Fatalf("An error has been occured: %s", err.Error())
	}

	log.Printf("elapsed %s", time.Since(timeStart).Round(time.Millisecond).String())
	log.Println("Download completed")
}
