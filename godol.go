package godol

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

// Godol is a module for downloading a file concurrently
type Godol struct {
	url         string
	destination string
	workers     int
	fileName    string
	wg          sync.WaitGroup
	result      []string
	writer      *bufio.Writer
}

// Option to give to Godol
type Option func(*Godol)

// New creates an instance of Godol with provided options
func New(options ...Option) *Godol {
	g := &Godol{}

	for _, option := range options {
		option(g)
	}

	return g
}

// WithURL option
func WithURL(url string) Option {
	return func(g *Godol) {
		g.url = url
	}
}

// WithDestination option
func WithDestination(dest string) Option {
	return func(g *Godol) {
		g.destination = dest
	}
}

// WithWorker option
func WithWorker(n int) Option {
	return func(g *Godol) {
		g.workers = n
		g.result = make([]string, n+1)
	}
}

// WithFilename option
func WithFilename(name string) Option {
	return func(g *Godol) {
		g.fileName = name
	}
}

// Start starts the task
func (g *Godol) Start() error {
	res, err := http.Head(g.url)
	if err != nil {
		return err
	}

	if g.fileName == "" {
		g.fileName = path.Base(res.Request.URL.Path)
	}

	f, err := os.Create(fmt.Sprintf("%s/%s", g.destination, g.fileName))
	if err != nil {
		return err
	}

	defer f.Close()

	g.writer = bufio.NewWriter(f)

	contentLength := res.ContentLength
	subBytes := contentLength / int64(g.workers)
	diff := contentLength % int64(g.workers)

	for i := 0; i < g.workers; i++ {
		g.wg.Add(1)

		min := subBytes * int64(i)
		max := subBytes * int64(i+1)
		if i == (g.workers - 1) {
			max += diff
		}

		go g.spawnWorker(min, max, i)

	}
	g.wg.Wait()
	return g.finalize()
}

func (g *Godol) spawnWorker(min, max int64, i int) {
	defer g.wg.Done()
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, g.url, nil)
	if err != nil {
		log.Fatalf("An error has been occured %s", err.Error())
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", min, max-1)
	req.Header.Add("Range", rangeHeader)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("An error has been occured %s", err.Error())
	}

	defer res.Body.Close()

	reader, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("An error has been occured %s", err.Error())
	}

	g.result[i] = string(reader)

}

func (g *Godol) finalize() error {
	for i, s := range g.result {
		n, err := g.writer.WriteString(s)
		if err != nil {
			return err
		}
		log.Printf("buffer[%d] : %d bytes", i, n)
	}
	if err := g.writer.Flush(); err != nil {
		return err
	}
	return nil
}
